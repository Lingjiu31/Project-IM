package logger

import (
	"go.uber.org/zap"
)

// Init 初始化全局 logger，程序启动时调用一次
// 开发环境用 NewDevelopment：彩色输出、对人友好
// 生产环境用 NewProduction：JSON 格式、适合日志系统采集
func Init() {
	logger, _ := zap.NewDevelopment()
	// 替换全局 logger，之后所有地方都可以用 zap.L() 调用
	zap.ReplaceGlobals(logger)
}
