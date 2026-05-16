package configs

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Logger *zap.Logger
	Sugar  *zap.SugaredLogger
)

func NewCustomLogger(logFile string) (*zap.Logger, error) {
	// 1. 配置编码器（重点是时间格式）
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",

		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
		},

		// 级别编码器 - 大写显示 INFO、ERROR 等
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}

	// 2. 创建文件输出
	file, err := os.OpenFile(
		logFile,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0666, // 文件权限
	)
	if err != nil {
		return nil, err
	}

	// 3. 创建多个输出（可以同时输出到文件和终端）
	writeSyncer := zapcore.NewMultiWriteSyncer(
		zapcore.AddSync(file),      // 输出到文件
		zapcore.AddSync(os.Stdout), // 同时输出到终端
	)

	// 4. 创建 Core
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig), // 使用 JSON 格式
		writeSyncer,
		zap.InfoLevel, // 设置日志级别
	)

	// 5. 创建 Logger，添加额外选项
	logger := zap.New(core,
		zap.AddCaller(),                   // 添加调用者信息
		zap.AddCallerSkip(1),              // 跳过一层调用栈
		zap.AddStacktrace(zap.ErrorLevel), // 错误级别记录堆栈
	)

	return logger, nil
}

func InitLogger() {
	Logger, _ = NewCustomLogger("Storage/Log/ZhiHu.log")
	Sugar = Logger.Sugar()
}
