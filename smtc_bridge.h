#ifndef MEDIA_DLL_H
#define MEDIA_DLL_H

#ifdef __cplusplus
extern "C" {
#endif

#include <stdbool.h>
#include <stdint.h>

/*
 * 回调函数类型
 * data: Rust 分配的内存
 * len : 数据长度（包含消息类型字节）
 */
typedef void (*Callback)(unsigned char* data, uint64_t len);

/* ================= 生命周期 ================= */

bool start_media_service(Callback cb);
void stop_media_service(void);

/* 释放 Rust 分配的内存 */
void free_rust_buffer(unsigned char* ptr, uint64_t len);

/* ================= Session 控制 ================= */

bool SelectSession(const char* id);

/* ================= 音频捕获 ================= */

bool StartAudioCapture(void);
bool StopAudioCapture(void);

/* ================= 配置 ================= */

bool SetHighFrequencyProgressUpdates(bool y);
bool SetProgressOffset(int64_t offset);
bool SetAppleMusicOptimization(bool y);

/* ================= 播放控制 ================= */

bool Pause(void);
bool Play(void);
bool SkipNext(void);
bool SkipPrevious(void);
bool SeekTo(uint64_t position);
bool SetVolume(float volume);
bool SetShuffle(bool y);

/*
 * RepeatMode:
 * 0 = Off
 * 1 = One
 * 2 = All
 */
bool SetRepeatMode(int32_t mode);


extern void GoCallback(unsigned char* ptr, int len);
#ifdef __cplusplus
}
#endif

#endif /* MEDIA_DLL_H */