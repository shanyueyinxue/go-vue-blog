package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LogConfig struct {
	Level          string `mapstructure:"level"`          // 日志级别: debug, info, warn, error
	Format         string `mapstructure:"format"`         // 日志格式: text, json
	Output         string `mapstructure:"output"`         // 日志输出目录
	OutputFileName string `mapstructure:"outputFileName"` // 日志文件名
	Rotate         bool   `mapstructure:"rotate"`         // 是否启用日志轮转
	RotateType     string `mapstructure:"rotateType"`     // 轮转类型: size, time
	MaxSize        int    `mapstructure:"maxSize"`        // 单个日志文件最大大小(字节)，用于size轮转
	MaxAge         int    `mapstructure:"maxAge"`         // 日志文件保留天数
	RotateInterval int    `mapstructure:"rotateInterval"` // 轮转间隔(天)，用于time轮转
	Stdout         bool   `mapstructure:"stdout"`         // 是否同时输出到标准输出
}

// NewLogger 根据配置初始化日志记录器; IsReplaceGlobals 是否替换全局的zap.Logger，默认 true。
func NewLogger(cfg LogConfig, IsReplaceGlobals ...bool) (*zap.Logger, error) {
	if len(IsReplaceGlobals) == 0 {
		IsReplaceGlobals = []bool{true}
	}
	isReplaceGlobal := IsReplaceGlobals[0]

	// 解析日志级别
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("解析日志级别失败: %v", err)
	}

	// 创建编码器
	encoder := createEncoder(cfg.Format)

	// 创建输出写入器
	writeSyncer, err := createWriteSyncer(cfg)
	if err != nil {
		return nil, fmt.Errorf("创建日志写入器失败: %v", err)
	}

	// 创建核心
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// 创建logger
	logger := zap.New(
		core,
		zap.AddCaller(),                   // 显示调用者信息
		zap.AddCallerSkip(1),              // 跳过当前调用层
		zap.AddStacktrace(zap.ErrorLevel), // 错误级别以上显示堆栈跟踪
	)
	if isReplaceGlobal {
		zap.ReplaceGlobals(logger)
	}
	return logger, nil
}

// parseLevel 解析日志级别字符串为zapcore.Level
func parseLevel(levelStr string) (zapcore.Level, error) {
	var level zapcore.Level
	switch levelStr {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		return level, fmt.Errorf("不支持的日志级别: %s", levelStr)
	}
	return level, nil
}

// createEncoder 根据格式创建日志编码器
func createEncoder(format string) zapcore.Encoder {
	// 基础配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder, // 级别大写，如 INFO, ERROR
		EncodeTime:     zapcore.ISO8601TimeEncoder,  // 时间格式 ISO8601
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, // 短路径调用者，如 pkg/file.go:23
	}

	if format == "json" {
		return zapcore.NewJSONEncoder(encoderConfig)
	}
	// 默认使用文本格式
	return zapcore.NewConsoleEncoder(encoderConfig)
}

// createWriteSyncer 创建日志写入器
func createWriteSyncer(cfg LogConfig) (zapcore.WriteSyncer, error) {
	// 确保输出目录存在
	if cfg.Output != "" {
		if err := os.MkdirAll(cfg.Output, 0755); err != nil {
			return nil, fmt.Errorf("创建日志目录失败: %v", err)
		}
	}

	// 日志文件路径
	logFilePath := filepath.Join(cfg.Output, cfg.OutputFileName)

	var writers []zapcore.WriteSyncer

	// 添加文件写入器（支持轮转）
	if cfg.Rotate {
		switch cfg.RotateType {
		case "size":
			// 使用lumberjack实现按大小轮转
			fileWriter := &lumberjack.Logger{
				Filename:   logFilePath,
				MaxSize:    cfg.MaxSize, // lumberjack的MaxSize单位是MB
				MaxAge:     cfg.MaxAge,
				MaxBackups: 0,     // 不限制备份数量
				Compress:   false, // 不压缩
			}
			writers = append(writers, zapcore.AddSync(fileWriter))

		case "time":
			// 使用rotatelogs实现按时间轮转
			// 轮转后的日志文件格式: app.log.20240520150405
			rotateLogs, err := rotatelogs.New(
				logFilePath+".%Y%m%d%H%M%S",
				rotatelogs.WithLinkName(logFilePath),                                        // 创建软链接指向最新日志文件
				rotatelogs.WithMaxAge(time.Duration(cfg.MaxAge)*24*time.Hour),               // 日志保留时间
				rotatelogs.WithRotationTime(time.Duration(cfg.RotateInterval)*24*time.Hour), // 轮转间隔
			)
			if err != nil {
				return nil, fmt.Errorf("初始化时间轮转失败: %v", err)
			}
			writers = append(writers, zapcore.AddSync(rotateLogs))

		default:
			return nil, fmt.Errorf("不支持的轮转类型: %s", cfg.RotateType)
		}
	} else {
		// 不启用轮转，直接写入文件
		file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, err
		}
		writers = append(writers, zapcore.AddSync(file))
	}

	// 如果需要同时输出到标准输出
	if cfg.Stdout {
		writers = append(writers, zapcore.AddSync(os.Stdout))
	}

	// 组合多个写入器
	return zapcore.NewMultiWriteSyncer(writers...), nil
}
