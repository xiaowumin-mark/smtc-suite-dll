package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xiaowumin-mark/go-smtc-suite"
)

// SimpleEventHandler 简单事件处理器
type SimpleEventHandler struct{}

func (h *SimpleEventHandler) OnTrackChanged(info *smtc.TrackInfo) {
	fmt.Printf("🎵 轨道更新: %s - %s\n", info.Artist, info.Title)
}

func (h *SimpleEventHandler) OnSessionsChanged(sessions []smtc.SessionInfo) {
	fmt.Printf("📱 会话列表更新: %d 个会话\n", len(sessions))
	for _, session := range sessions {
		fmt.Printf("   - %s (%s)\n", session.DisplayName, session.SessionID)
	}
}

func (h *SimpleEventHandler) OnVolumeChanged(info *smtc.VolumeInfo) {
	fmt.Printf("🔊 音量变化: %.0f%% (静音: %v)\n", info.Volume*100, info.IsMuted)
}

func (h *SimpleEventHandler) OnSelectedSessionVanished(info *smtc.SelectedSessionVanishedInfo) {
	fmt.Printf("❌ 选中会话消失: %s\n", info.Info)
}

func (h *SimpleEventHandler) OnError(info *smtc.ErrorInfo) {
	fmt.Printf("⚠️ 错误: %s\n", info.Info)
}

func (h *SimpleEventHandler) OnAudioData(data []byte) {
	fmt.Printf("🎧 音频数据: %d 字节\n", len(data))
}

func (h *SimpleEventHandler) OnCoverData(data []byte) {
	fmt.Printf("🖼️ 封面数据: %d 字节\n", len(data))
}

func main() {
	fmt.Println("=== SMTC套件简单测试 ===")
	fmt.Println("这个示例演示了SMTC套件的基本功能")
	fmt.Println()

	// 创建SMTC包装器
	wrapper := smtc.NewSMTCWrapper()

	// 设置事件处理器
	handler := &SimpleEventHandler{}
	wrapper.SetEventHandler(handler)

	// 启动媒体服务
	fmt.Println("正在启动媒体服务...")
	if err := wrapper.Start(); err != nil {
		fmt.Printf("❌ 启动失败: %v\n", err)
		return
	}
	fmt.Println("✅ 媒体服务启动成功!")
	fmt.Println()

	// 设置一些配置
	fmt.Println("正在配置服务...")
	wrapper.SetHighFrequencyProgressUpdates(true)
	wrapper.SetAppleMusicOptimization(true)
	fmt.Println("✅ 配置完成")
	fmt.Println()

	fmt.Println("🎯 服务正在运行中...")
	fmt.Println("   现在可以播放音乐或视频，观察事件输出")
	fmt.Println("   按 Ctrl+C 退出程序")
	fmt.Println()

	// 设置信号处理，优雅退出
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("\n🛑 收到退出信号，正在停止服务...")

		// 停止音频捕获
		if err := wrapper.StopAudioCapture(); err != nil {
			fmt.Printf("⚠️ 停止音频捕获失败: %v\n", err)
		} else {
			fmt.Println("✅ 音频捕获已停止")
		}

		// 停止媒体服务
		wrapper.Stop()
		fmt.Println("✅ 媒体服务已停止")
		fmt.Println("👋 程序退出")
		os.Exit(0)
	}()

	// 保持程序运行，等待事件
	for {
		time.Sleep(1 * time.Second)

		// 每10秒检查一次服务状态
		if !wrapper.IsRunning() {
			fmt.Println("❌ 服务意外停止")
			break
		}
	}
}

/*package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xiaowumin-mark/go-smtc-suite"
)

// MyEventHandler 自定义事件处理器
type MyEventHandler struct{}

func (h *MyEventHandler) OnTrackChanged(info *smtc.TrackInfo) {
	fmt.Printf("=== 轨道信息更新 ===\n")
	fmt.Printf("标题: %s\n", info.Title)
	fmt.Printf("艺术家: %s\n", info.Artist)
	fmt.Printf("专辑: %s\n", info.AlbumTitle)
	fmt.Printf("专辑艺术家: %s\n", info.AlbumArtist)
	fmt.Printf("流派: %s\n", info.Genres)
	fmt.Printf("轨道号: %d/%d\n", info.TrackNumber, info.AlbumTrackCount)
	fmt.Printf("时长: %d ms\n", info.Duration)
	fmt.Printf("位置: %d ms\n", info.Position)
	fmt.Printf("SMTC位置: %d ms\n", info.SMTCPostionMS)
	fmt.Printf("播放状态: %s\n", info.PlaybackStatus)
	fmt.Printf("重复模式: %s\n", info.RepeatMode)
	fmt.Println()
}

func (h *MyEventHandler) OnSessionsChanged(sessions []smtc.SessionInfo) {
	fmt.Printf("=== 会话列表更新 ===\n")
	fmt.Printf("发现 %d 个会话:\n", len(sessions))
	for i, session := range sessions {
		fmt.Printf("  %d. %s (ID: %s)\n", i+1, session.DisplayName, session.SessionID)
		fmt.Printf("     应用ID: %s\n", session.SourceAppUserModelID)
	}
	fmt.Println()
}

func (h *MyEventHandler) OnVolumeChanged(info *smtc.VolumeInfo) {
	fmt.Printf("=== 音量变化 ===\n")
	fmt.Printf("会话ID: %s\n", info.SessionID)
	fmt.Printf("音量: %.2f\n", info.Volume)
	fmt.Printf("静音: %v\n", info.IsMuted)
	fmt.Println()
}

func (h *MyEventHandler) OnSelectedSessionVanished(info *smtc.SelectedSessionVanishedInfo) {
	fmt.Printf("=== 选中会话消失 ===\n")
	fmt.Printf("信息: %s\n", info.Info)
	fmt.Println()
}

func (h *MyEventHandler) OnError(info *smtc.ErrorInfo) {
	fmt.Printf("=== 错误信息 ===\n")
	fmt.Printf("错误: %s\n", info.Info)
	fmt.Println()
}

func (h *MyEventHandler) OnAudioData(data []byte) {
	fmt.Printf("=== 音频数据 ===\n")
	fmt.Printf("收到 %d 字节音频数据\n", len(data))
	// 这里可以处理音频数据，比如保存到文件或进行音频分析
	fmt.Println()
}

func (h *MyEventHandler) OnCoverData(data []byte) {
	fmt.Printf("=== 封面数据 ===\n")
	fmt.Printf("收到 %d 字节封面数据\n", len(data))
	// 这里可以处理封面数据，比如保存到文件或进行图片处理
	fmt.Println()
}

func main() {
	fmt.Println("=== SMTC套件Go封装示例 ===")
	fmt.Println()

	// 创建SMTC包装器实例
	wrapper := smtc.NewSMTCWrapper()

	// 设置事件处理器
	handler := &MyEventHandler{}
	wrapper.SetEventHandler(handler)

	// 启动媒体服务
	fmt.Println("正在启动媒体服务...")
	if err := wrapper.Start(); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
	fmt.Println("媒体服务启动成功!")
	fmt.Println()

	// 等待一段时间让服务初始化
	time.Sleep(2 * time.Second)

	// 演示各种控制功能
	demoControlFunctions(wrapper)

	// 设置信号处理，优雅退出
	setupSignalHandler(wrapper)

	// 保持程序运行
	fmt.Println("程序运行中，按 Ctrl+C 退出...")
	select {}
}

func demoControlFunctions(wrapper *smtc.SMTCWrapper) {
	// 演示配置设置
	fmt.Println("=== 配置演示 ===")

	// 设置高频进度更新
	if err := wrapper.SetHighFrequencyProgressUpdates(true); err != nil {
		fmt.Printf("设置高频进度更新失败: %v\n", err)
	} else {
		fmt.Println("✓ 高频进度更新已启用")
	}

	// 设置进度偏移
	if err := wrapper.SetProgressOffset(1000); err != nil {
		fmt.Printf("设置进度偏移失败: %v\n", err)
	} else {
		fmt.Println("✓ 进度偏移已设置为 1000ms")
	}

	// 设置Apple Music优化
	if err := wrapper.SetAppleMusicOptimization(true); err != nil {
		fmt.Printf("设置Apple Music优化失败: %v\n", err)
	} else {
		fmt.Println("✓ Apple Music优化已启用")
	}

	fmt.Println()

	// 演示播放控制（需要先有活跃的会话）
	fmt.Println("=== 播放控制演示 ===")
	fmt.Println("注意: 这些操作需要先有活跃的媒体会话")
	fmt.Println()

	// 等待一段时间让会话信息到达
	time.Sleep(3 * time.Second)

	// 尝试播放控制（这些操作在实际使用时需要根据会话状态决定）
	demoPlaybackControls(wrapper)
}

func demoPlaybackControls(wrapper *smtc.SMTCWrapper) {
	// 演示播放控制功能
	// 注意：这些操作在实际应用中应该根据当前会话状态来决定是否执行

	fmt.Println("尝试播放控制操作...")

	// 尝试播放
	if err := wrapper.Play(); err != nil {
		fmt.Printf("播放操作失败: %v (可能没有活跃会话)\n", err)
	} else {
		fmt.Println("✓ 播放命令已发送")
	}

	time.Sleep(1 * time.Second)

	// 尝试暂停
	if err := wrapper.Pause(); err != nil {
		fmt.Printf("暂停操作失败: %v\n", err)
	} else {
		fmt.Println("✓ 暂停命令已发送")
	}

	time.Sleep(1 * time.Second)

	// 设置音量
	if err := wrapper.SetVolume(0.8); err != nil {
		fmt.Printf("设置音量失败: %v\n", err)
	} else {
		fmt.Println("✓ 音量已设置为 80%")
	}

	time.Sleep(1 * time.Second)

	// 设置随机播放
	if err := wrapper.SetShuffle(true); err != nil {
		fmt.Printf("设置随机播放失败: %v\n", err)
	} else {
		fmt.Println("✓ 随机播放已启用")
	}

	time.Sleep(1 * time.Second)

	// 设置重复模式
	if err := wrapper.SetRepeatMode(2); err != nil {
		fmt.Printf("设置重复模式失败: %v\n", err)
	} else {
		fmt.Println("✓ 重复模式已设置为 All")
	}

	fmt.Println()
}

func setupSignalHandler(wrapper *smtc.SMTCWrapper) {
	// 设置信号处理，优雅退出
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("\n收到退出信号，正在停止服务...")

		// 停止音频捕获（如果正在运行）
		if err := wrapper.StopAudioCapture(); err != nil {
			fmt.Printf("停止音频捕获失败: %v\n", err)
		} else {
			fmt.Println("✓ 音频捕获已停止")
		}

		// 停止媒体服务
		wrapper.Stop()
		fmt.Println("✓ 媒体服务已停止")
		fmt.Println("程序退出")
		os.Exit(0)
	}()
}

// 高级使用示例：会话管理
func advancedSessionManagement(wrapper *smtc.SMTCWrapper) {
	// 这个函数展示了如何在实际应用中进行会话管理
	// 需要在实际有会话信息时调用

	fmt.Println("=== 高级会话管理示例 ===")

	// 假设我们收到了会话列表，这里模拟一个会话选择过程
	sessionID := "example-session-id" // 实际使用时应该从OnSessionsChanged事件中获取

	fmt.Printf("尝试选择会话: %s\n", sessionID)
	if err := wrapper.SelectSession(sessionID); err != nil {
		fmt.Printf("选择会话失败: %v\n", err)
	} else {
		fmt.Println("✓ 会话选择命令已发送")
	}

	// 开始音频捕获
	time.Sleep(2 * time.Second)
	fmt.Println("开始音频捕获...")
	if err := wrapper.StartAudioCapture(); err != nil {
		fmt.Printf("开始音频捕获失败: %v\n", err)
	} else {
		fmt.Println("✓ 音频捕获已开始")
	}

	fmt.Println()
}

// 工具函数：将结构体转换为JSON字符串（用于调试）
func toJSON(v interface{}) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("JSON序列化错误: %v", err)
	}
	return string(data)
}
*/
