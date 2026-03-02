use once_cell::sync::OnceCell;
use serde::{Deserialize, Serialize};
use smtc_suite::{MediaCommand, MediaManager, MediaUpdate};
use std::ffi::CStr;
use std::os::raw::c_char;
use std::sync::Mutex;
use tokio::runtime::Runtime;

type Callback = extern "C" fn(*mut u8, usize);

static RUNTIME: OnceCell<Mutex<Option<Runtime>>> = OnceCell::new();
static CALLBACK: OnceCell<Mutex<Option<Callback>>> = OnceCell::new();
static CONTROLLER: OnceCell<Mutex<Option<smtc_suite::MediaController>>> = OnceCell::new();

#[derive(Serialize, Deserialize)]
struct Message {
    #[serde(rename = "type")]
    msg_type: String,
    data: serde_json::Value,
}

// 定义消息类型的枚举
#[repr(u8)]
enum MessageType {
    Json = 0,
    Audio = 1,
    Cover = 2,
}

#[unsafe(no_mangle)]
pub extern "C" fn start_media_service(cb: Callback) -> bool {
    if RUNTIME.get().is_some() {
        return false;
    }

    let rt = Runtime::new().unwrap();

    RUNTIME.set(Mutex::new(Some(rt))).ok();
    CALLBACK.set(Mutex::new(Some(cb))).ok();
    CONTROLLER.set(Mutex::new(None)).ok();

    let runtime_mutex = RUNTIME.get().unwrap();
    let runtime = runtime_mutex.lock().unwrap();
    let rt_ref = runtime.as_ref().unwrap();

    rt_ref.spawn(async move {
        let (controller, mut update_rx) = match MediaManager::start() {
            Ok(v) => v,
            Err(e) => {
                println!("MediaManager error: {:?}", e);
                return;
            }
        };

        // 存储 controller
        if let Some(ctrl_mutex) = CONTROLLER.get() {
            *ctrl_mutex.lock().unwrap() = Some(controller);
        }

        // lastCoverDataHash
        let mut last_cover_data_hash = None;

        while let Some(update) = update_rx.recv().await {
            match update {
                MediaUpdate::TrackChanged(info) => {

                    // 创建 JSON 消息
                    let track_data = serde_json::json!({
                        "title": info.title.unwrap_or_default(),
                        "artist": info.artist.unwrap_or_default(),
                        "album_title": info.album_title.unwrap_or_default(),
                        "album_artist": info.album_artist.unwrap_or_default(),
                        "genres": info.genres,
                        "track_number": info.track_number,
                        "album_track_count" : info.album_track_count,
                        "duration": info.duration_ms,
                        "position": info.position_ms,
                        "smtc_position_ms" : info.smtc_position_ms,
                        "playback_status": match info.playback_status {
                             Some(status) => match status {
                                smtc_suite::PlaybackStatus::Playing => "playing",
                                smtc_suite::PlaybackStatus::Paused => "paused",
                                smtc_suite::PlaybackStatus::Stopped => "stopped",
                            },
                            None => "unknown",
                        },
                        "repeat_mode": match info.repeat_mode {
                            Some(mode) => match mode {
                                smtc_suite::RepeatMode::Off => "off",
                                smtc_suite::RepeatMode::One => "one",
                                smtc_suite::RepeatMode::All => "all",
                            },
                            None => "unknown",
                        },
                        //"cover_data_hash": info.cover_data_hash,
                    });

                    let message = Message::new("TrackChanged", track_data);
                    let json_msg = serde_json::to_string(&message).unwrap();
                    send_to_go_with_type(MessageType::Json, json_msg.into_bytes());

                    if info.cover_data_hash != last_cover_data_hash {
                        last_cover_data_hash = info.cover_data_hash;
                        if let Some(cover_data) = info.cover_data {
                            send_to_go_with_type(MessageType::Cover, cover_data);
                        }
                    }
                }
                MediaUpdate::SessionsChanged(info) => {
                    let session_array: Vec<_> = info
                        .iter()
                        .map(|session_info| {
                            serde_json::json!({
                                "session_id": session_info.session_id,
                                "source_app_user_model_id": session_info.source_app_user_model_id,
                                "display_name": session_info.display_name,

                            })
                        })
                        .collect();

                    let playback_data = serde_json::Value::Array(session_array);

                    let message = Message::new("SessionsChanged", playback_data);
                    let json_msg = serde_json::to_string(&message).unwrap();
                    send_to_go_with_type(MessageType::Json, json_msg.into_bytes());
                }
                MediaUpdate::AudioData(mut info) => {
                    send_to_go_with_type(MessageType::Audio, info);
                }
                MediaUpdate::VolumeChanged {
                    session_id,
                    volume,
                    is_muted,
                } => {
                    let td = serde_json::json!({
                        "session_id": session_id,
                        "volume": volume,
                        "is_muted": is_muted,
                    });
                    let message = Message::new("VolumeChanged", td);
                    let json_msg = serde_json::to_string(&message).unwrap();
                    send_to_go_with_type(MessageType::Json, json_msg.into_bytes());
                }
                MediaUpdate::SelectedSessionVanished(info) => {
                    let td = serde_json::json!({
                        "info": info,
                    });
                    let message = Message::new("SelectedSessionVanished", td);
                    let json_msg = serde_json::to_string(&message).unwrap();
                    send_to_go_with_type(MessageType::Json, json_msg.into_bytes());
                }
                MediaUpdate::Error(info) => {
                    let td = serde_json::json!({
                        "info": info,
                    });
                    let message = Message::new("Error", td);
                    let json_msg = serde_json::to_string(&message).unwrap();
                    send_to_go_with_type(MessageType::Json, json_msg.into_bytes());
                }
                _ => {}
            }
        }
    });

    true
}

// 同时需要添加 Message 结构体的 new 方法
impl Message {
    fn new(msg_type: &str, data: serde_json::Value) -> Self {
        Message {
            msg_type: msg_type.to_string(),
            data,
        }
    }
}
#[unsafe(no_mangle)]
pub extern "C" fn stop_media_service() {
    if let Some(runtime_mutex) = RUNTIME.get() {
        let mut guard = runtime_mutex.lock().unwrap();
        if let Some(rt) = guard.take() {
            drop(rt);
        }
    }
}

#[unsafe(no_mangle)]
pub extern "C" fn SelectSession(id: *const c_char) -> bool {
    if id.is_null() {
        return false;
    }

    let c_str = unsafe { CStr::from_ptr(id) };
    let rust_str = match c_str.to_str() {
        // 使用 match 处理可能的错误
        Ok(s) => s,
        Err(_) => return false, // 如果转换失败则返回 false
    };
    send_command(MediaCommand::SelectSession(rust_str.to_string())); // 转换为 String
    true
}

#[unsafe(no_mangle)]
pub extern "C" fn StartAudioCapture() -> bool {
    send_command(MediaCommand::StartAudioCapture);
    true
}

#[unsafe(no_mangle)]
pub extern "C" fn StopAudioCapture() -> bool {
    send_command(MediaCommand::StopAudioCapture);
    true
}

#[unsafe(no_mangle)]
pub extern "C" fn SetHighFrequencyProgressUpdates(y: bool) -> bool {
    send_command(MediaCommand::SetHighFrequencyProgressUpdates(y));
    true
}

//SetProgress0ffset
#[unsafe(no_mangle)]
pub extern "C" fn SetProgressOffset(offset: i64) -> bool {
    send_command(MediaCommand::SetProgressOffset(offset));
    true
}

//SetAppleMusicOptimization
#[unsafe(no_mangle)]
pub extern "C" fn SetAppleMusicOptimization(y: bool) -> bool {
    send_command(MediaCommand::SetAppleMusicOptimization(y));
    true
}

// Pause
#[unsafe(no_mangle)]
pub extern "C" fn Pause() -> bool {
    send_command(MediaCommand::Control(smtc_suite::SmtcControlCommand::Pause));
    true
}

// Play
#[unsafe(no_mangle)]
pub extern "C" fn Play() -> bool {
    send_command(MediaCommand::Control(smtc_suite::SmtcControlCommand::Play));
    true
}

// SkipNext
#[unsafe(no_mangle)]
pub extern "C" fn SkipNext() -> bool {
    send_command(MediaCommand::Control(
        smtc_suite::SmtcControlCommand::SkipNext,
    ));
    true
}

// SkipPrevious
#[unsafe(no_mangle)]
pub extern "C" fn SkipPrevious() -> bool {
    send_command(MediaCommand::Control(
        smtc_suite::SmtcControlCommand::SkipPrevious,
    ));
    true
}

//SeekTo u64
#[unsafe(no_mangle)]
pub extern "C" fn SeekTo(position: u64) -> bool {
    send_command(MediaCommand::Control(
        smtc_suite::SmtcControlCommand::SeekTo(position),
    ));
    true
}

// SetVolume f32
#[unsafe(no_mangle)]
pub extern "C" fn SetVolume(volume: f32) -> bool {
    send_command(MediaCommand::Control(
        smtc_suite::SmtcControlCommand::SetVolume(volume),
    ));
    true
}

//SetShuffle
#[unsafe(no_mangle)]
pub extern "C" fn SetShuffle(y: bool) -> bool {
    send_command(MediaCommand::Control(
        smtc_suite::SmtcControlCommand::SetShuffle(y),
    ));
    true
}

//SetRepeatMode
#[unsafe(no_mangle)]
pub extern "C" fn SetRepeatMode(mode: i32) -> bool {
    let repeat_mode = match mode {
        0 => smtc_suite::RepeatMode::Off,
        1 => smtc_suite::RepeatMode::One,
        2 => smtc_suite::RepeatMode::All,
        _ => smtc_suite::RepeatMode::Off, // 默认值
    };

    send_command(MediaCommand::Control(
        smtc_suite::SmtcControlCommand::SetRepeatMode(repeat_mode),
    ));
    true
}
fn send_command(cmd: MediaCommand) {
    if let Some(ctrl_mutex) = CONTROLLER.get() {
        if let Some(controller) = &*ctrl_mutex.lock().unwrap() {
            let tx = controller.command_tx.clone();

            // 在 runtime 中发送 async 命令
            if let Some(rt_mutex) = RUNTIME.get() {
                if let Some(rt) = rt_mutex.lock().unwrap().as_ref() {
                    rt.spawn(async move {
                        let _ = tx.send(cmd).await;
                    });
                }
            }
        }
    }
}

// 修改 send_to_go 函数，添加消息类型标识符
fn send_to_go_with_type(msg_type: MessageType, mut data: Vec<u8>) {
    if let Some(cb_mutex) = CALLBACK.get() {
        if let Some(cb) = *cb_mutex.lock().unwrap() {
            // 创建新的字节数组，包含类型标识符 + 原始数据
            let mut full_data = Vec::with_capacity(1 + data.len());
            full_data.push(msg_type as u8);
            full_data.extend_from_slice(&data);

            let ptr = full_data.as_mut_ptr();
            let len = full_data.len();
            std::mem::forget(full_data);
            cb(ptr, len);
        }
    }
}

#[unsafe(no_mangle)]
pub extern "C" fn free_rust_buffer(ptr: *mut u8, len: usize) {
    if ptr.is_null() {
        return;
    }
    unsafe {
        let _ = Vec::from_raw_parts(ptr, len, len);
    }
}
