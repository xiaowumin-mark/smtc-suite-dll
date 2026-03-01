package smtc

/*
#cgo LDFLAGS: -L. -lsmtc_suite_dll
#include <stdlib.h>
#include "smtc_bridge.h"
*/
import "C"
import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"unsafe"
)

// MessageType 定义消息类型
const (
	MessageTypeJSON = iota
	MessageTypeBinary
)

// Message 基础消息结构
type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// TrackInfo 轨道信息结构
type TrackInfo struct {
	Title           string   `json:"title"`
	Artist          string   `json:"artist"`
	AlbumTitle      string   `json:"album_title"`
	AlbumArtist     string   `json:"album_artist"`
	Genres          []string `json:"genres"`
	TrackNumber     int      `json:"track_number"`
	AlbumTrackCount int      `json:"album_track_count"`
	Duration        int64    `json:"duration"`
	Position        int64    `json:"position"`
	SMTCPostionMS   int64    `json:"smtc_position_ms"`
	PlaybackStatus  string   `json:"playback_status"`
	RepeatMode      string   `json:"repeat_mode"`
}

// SessionInfo 会话信息结构
type SessionInfo struct {
	SessionID            string `json:"session_id"`
	SourceAppUserModelID string `json:"source_app_user_model_id"`
	DisplayName          string `json:"display_name"`
}

// VolumeInfo 音量信息结构
type VolumeInfo struct {
	SessionID string  `json:"session_id"`
	Volume    float32 `json:"volume"`
	IsMuted   bool    `json:"is_muted"`
}

// ErrorInfo 错误信息结构
type ErrorInfo struct {
	Info string `json:"info"`
}

// SelectedSessionVanishedInfo 选中会话消失信息
type SelectedSessionVanishedInfo struct {
	Info string `json:"info"`
}

// EventHandler 事件处理器接口
type EventHandler interface {
	OnTrackChanged(info *TrackInfo)
	OnSessionsChanged(sessions []SessionInfo)
	OnVolumeChanged(info *VolumeInfo)
	OnSelectedSessionVanished(info *SelectedSessionVanishedInfo)
	OnError(info *ErrorInfo)
	OnAudioData(data []byte)
}

// SMTCWrapper SMTC套件包装器
type SMTCWrapper struct {
	eventHandler EventHandler
	isRunning    atomic.Bool
}

var (
	globalEventHandler EventHandler
	eventHandlerMutex  sync.RWMutex
)

// NewSMTCWrapper 创建新的SMTC包装器
func NewSMTCWrapper() *SMTCWrapper {
	return &SMTCWrapper{}
}

// SetEventHandler 设置事件处理器
func (w *SMTCWrapper) SetEventHandler(handler EventHandler) {
	eventHandlerMutex.Lock()
	defer eventHandlerMutex.Unlock()
	globalEventHandler = handler
	w.eventHandler = handler
}

// Start 启动媒体服务
func (w *SMTCWrapper) Start() error {
	if w.isRunning.Swap(true) {
		return errors.New("media service is already running")
	}

	// 注意这里强制转换类型以匹配 .h 文件
	ok := C.start_media_service((C.Callback)(C.GoCallback))
	if !ok {
		w.isRunning.Store(false)
		return errors.New("failed to start media service")
	}
	return nil
}

// Stop 停止媒体服务
func (w *SMTCWrapper) Stop() {
	if w.isRunning.Swap(false) {
		C.stop_media_service()
	}
}

// SelectSession 选择会话
func (w *SMTCWrapper) SelectSession(sessionID string) error {
	if !w.isRunning.Load() {
		return errors.New("media service is not running")
	}

	cSessionID := C.CString(sessionID)
	defer C.free(unsafe.Pointer(cSessionID))

	ok := C.SelectSession(cSessionID)
	if !ok {
		return errors.New("failed to select session")
	}

	return nil
}

// StartAudioCapture 开始音频捕获
func (w *SMTCWrapper) StartAudioCapture() error {
	if !w.isRunning.Load() {
		return errors.New("media service is not running")
	}

	ok := C.StartAudioCapture()
	if !ok {
		return errors.New("failed to start audio capture")
	}

	return nil
}

// StopAudioCapture 停止音频捕获
func (w *SMTCWrapper) StopAudioCapture() error {
	if !w.isRunning.Load() {
		return errors.New("media service is not running")
	}

	ok := C.StopAudioCapture()
	if !ok {
		return errors.New("failed to stop audio capture")
	}

	return nil
}

// SetHighFrequencyProgressUpdates 设置高频进度更新
func (w *SMTCWrapper) SetHighFrequencyProgressUpdates(enable bool) error {
	if !w.isRunning.Load() {
		return errors.New("media service is not running")
	}

	ok := C.SetHighFrequencyProgressUpdates(C.bool(enable))
	if !ok {
		return errors.New("failed to set high frequency progress updates")
	}

	return nil
}

// SetProgressOffset 设置进度偏移
func (w *SMTCWrapper) SetProgressOffset(offset int64) error {
	if !w.isRunning.Load() {
		return errors.New("media service is not running")
	}

	ok := C.SetProgressOffset(C.longlong(offset))
	if !ok {
		return errors.New("failed to set progress offset")
	}

	return nil
}

// SetAppleMusicOptimization 设置Apple Music优化
func (w *SMTCWrapper) SetAppleMusicOptimization(enable bool) error {
	if !w.isRunning.Load() {
		return errors.New("media service is not running")
	}

	ok := C.SetAppleMusicOptimization(C.bool(enable))
	if !ok {
		return errors.New("failed to set apple music optimization")
	}

	return nil
}

// Pause 暂停播放
func (w *SMTCWrapper) Pause() error {
	if !w.isRunning.Load() {
		return errors.New("media service is not running")
	}

	ok := C.Pause()
	if !ok {
		return errors.New("failed to pause")
	}

	return nil
}

// Play 播放
func (w *SMTCWrapper) Play() error {
	if !w.isRunning.Load() {
		return errors.New("media service is not running")
	}

	ok := C.Play()
	if !ok {
		return errors.New("failed to play")
	}

	return nil
}

// SkipNext 下一首
func (w *SMTCWrapper) SkipNext() error {
	if !w.isRunning.Load() {
		return errors.New("media service is not running")
	}

	ok := C.SkipNext()
	if !ok {
		return errors.New("failed to skip next")
	}

	return nil
}

// SkipPrevious 上一首
func (w *SMTCWrapper) SkipPrevious() error {
	if !w.isRunning.Load() {
		return errors.New("media service is not running")
	}

	ok := C.SkipPrevious()
	if !ok {
		return errors.New("failed to skip previous")
	}

	return nil
}

// SeekTo 跳转到指定位置
func (w *SMTCWrapper) SeekTo(position uint64) error {
	if !w.isRunning.Load() {
		return errors.New("media service is not running")
	}

	ok := C.SeekTo(C.ulonglong(position))
	if !ok {
		return errors.New("failed to seek to position")
	}

	return nil
}

// SetVolume 设置音量
func (w *SMTCWrapper) SetVolume(volume float32) error {
	if !w.isRunning.Load() {
		return errors.New("media service is not running")
	}

	ok := C.SetVolume(C.float(volume))
	if !ok {
		return errors.New("failed to set volume")
	}

	return nil
}

// SetShuffle 设置随机播放
func (w *SMTCWrapper) SetShuffle(enable bool) error {
	if !w.isRunning.Load() {
		return errors.New("media service is not running")
	}

	ok := C.SetShuffle(C.bool(enable))
	if !ok {
		return errors.New("failed to set shuffle")
	}

	return nil
}

// SetRepeatMode 设置重复模式
func (w *SMTCWrapper) SetRepeatMode(mode int) error {
	if !w.isRunning.Load() {
		return errors.New("media service is not running")
	}

	if mode < 0 || mode > 2 {
		return errors.New("invalid repeat mode: must be 0 (Off), 1 (One), or 2 (All)")
	}

	ok := C.SetRepeatMode(C.int(mode))
	if !ok {
		return errors.New("failed to set repeat mode")
	}

	return nil
}

// IsRunning 检查服务是否运行
func (w *SMTCWrapper) IsRunning() bool {
	return w.isRunning.Load()
}

//export GoCallback
func GoCallback(data *C.uchar, length C.int) { // 确保是 C.uint64_t
	// C.GoBytes 的第二个参数需要 int，所以这里需要转换
	bytes := C.GoBytes(unsafe.Pointer(data), C.int(length))

	if len(bytes) < 1 {
		return
	}

	msgType := bytes[0]
	payload := bytes[1:]

	switch msgType {
	case MessageTypeJSON:
		handleJSONMessage(payload)
	case MessageTypeBinary:
		handleBinaryMessage(payload)
	}

	// 释放 Rust 分配的内存
	// 注意：Rust 端的 free_rust_buffer 也接收 usize (uint64_t)
	C.free_rust_buffer(data, C.size_t(length))
}

func handleJSONMessage(data []byte) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		fmt.Printf("Failed to parse JSON message: %v\n", err)
		return
	}

	eventHandlerMutex.RLock()
	handler := globalEventHandler
	eventHandlerMutex.RUnlock()

	if handler == nil {
		fmt.Printf("Received message type: %s (no event handler registered)\n", msg.Type)
		fmt.Printf("Data: %s\n", string(msg.Data))
		return
	}

	switch msg.Type {
	case "TrackChanged":
		var trackInfo TrackInfo
		if err := json.Unmarshal(msg.Data, &trackInfo); err == nil {
			handler.OnTrackChanged(&trackInfo)
		} else {
			fmt.Printf("Failed to parse TrackChanged data: %v\n", err)
		}
	case "SessionsChanged":
		var sessions []SessionInfo
		if err := json.Unmarshal(msg.Data, &sessions); err == nil {
			handler.OnSessionsChanged(sessions)
		} else {
			fmt.Printf("Failed to parse SessionsChanged data: %v\n", err)
		}
	case "VolumeChanged":
		var volumeInfo VolumeInfo
		if err := json.Unmarshal(msg.Data, &volumeInfo); err == nil {
			handler.OnVolumeChanged(&volumeInfo)
		} else {
			fmt.Printf("Failed to parse VolumeChanged data: %v\n", err)
		}
	case "SelectedSessionVanished":
		var vanishedInfo SelectedSessionVanishedInfo
		if err := json.Unmarshal(msg.Data, &vanishedInfo); err == nil {
			handler.OnSelectedSessionVanished(&vanishedInfo)
		} else {
			fmt.Printf("Failed to parse SelectedSessionVanished data: %v\n", err)
		}
	case "Error":
		var errorInfo ErrorInfo
		if err := json.Unmarshal(msg.Data, &errorInfo); err == nil {
			handler.OnError(&errorInfo)
		} else {
			fmt.Printf("Failed to parse Error data: %v\n", err)
		}
	default:
		fmt.Printf("Unknown message type: %s\n", msg.Type)
		fmt.Printf("Data: %s\n", string(msg.Data))
	}
}

func handleBinaryMessage(data []byte) {
	eventHandlerMutex.RLock()
	handler := globalEventHandler
	eventHandlerMutex.RUnlock()

	if handler == nil {
		fmt.Printf("Received binary data: %d bytes (no event handler registered)\n", len(data))
		return
	}

	handler.OnAudioData(data)
}
