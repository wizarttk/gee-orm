/*
这个日志模块实现了一个 简单的、线程安全的、支持级别控制的日志系统，具有以下特性：
  - 支持两种日志等级（info 和 error）
  - 可以通过 SetLevel 统一关闭某类日志（甚至全部关闭）
  - 日志输出加上颜色标签，便于终端阅读
  - 封装了简单的 Println / Printf 方法供调用
  - 线程安全支持（使用 sync.Mutex）package log
*/
package log

import (
	"io"
	"log"
	"os"
	"sync"
)

// 定义两个全局日志记录器：errorLog 和 infoLog。
// 使用 ANSI 转义码添加颜色（红色表示错误，蓝色表示信息）以区分日志类型。
// log.LstdFlags: 标准日志头（时间+日期）；log.Lshortfile: 打印调用日志的文件名和行号。
var (
	errorLog = log.New(os.Stdout, "\033[31m[error]\033[0m ", log.LstdFlags|log.Lshortfile)
	infoLog  = log.New(os.Stdout, "\033[34m[info ]\033[0m ", log.LstdFlags|log.Lshortfile)
	loggers  = []*log.Logger{errorLog, infoLog} // 用于统一管理所有日志记录器
	mu       sync.Mutex                         // 用于线程安全地修改日志输出级别
)

// 将日志方法暴露出去，供外部使用。
// 这样可以直接调用 log.Error("...") 或 log.Infof("...") 来输出日志。
var (
	Error  = errorLog.Println // 输出 error 类型的日志，自动换行
	Errorf = errorLog.Printf  // 输出 error 类型的日志，格式化输出
	Info   = infoLog.Println  // 输出 info 类型的日志，自动换行
	Infof  = infoLog.Printf   // 输出 info 类型的日志，格式化输出
)

// 定义日志级别常量，使用 iota 自动递增
const (
	InfoLevel  = iota // 0，表示显示 info 和 error 日志
	ErrorLevel        // 1，只显示 error 日志
	Disabled          // 2，禁用所有日志输出
)

// SetLevel 设置全局日志等级，控制日志的输出行为
func SetLevel(level int) {
	mu.Lock()         // 加锁，防止并发修改 loggers 导致数据竞争
	defer mu.Unlock() // 解锁，确保函数退出时释放锁

	// 在每次调用 SetLevel 函数时，先将所有的日志输出目标重置回标准输出（os.Stdout）
	for _, logger := range loggers {
		logger.SetOutput(os.Stdout)
	}

	// 如果设置的等级比 ErrorLevel 高，则关闭 errorLog 输出
	if ErrorLevel < level {
		errorLog.SetOutput(io.Discard) // io.Discard 表示丢弃输出
	}
	// 如果设置的等级比 InfoLevel 高，则关闭 infoLog 输出
	if InfoLevel < level {
		infoLog.SetOutput(io.Discard)
	}
}
