package logger

import (
	"fmt"
	"imapsync-grpc/config"
	"imapsync-grpc/internal/util"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/metadata"
	"gopkg.in/natefinch/lumberjack.v2"
)

var log *zap.SugaredLogger

type DT map[string]any

type DataLogger struct {
	Meta          any `json:"meta"`
	Data          DT  `json:"data"`
	TransactionId any `json:"transactionId,omitempty"`
}

func InitLog(logPath, level string) {

	lumberjackLogger := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    100,
		MaxBackups: 7,
		MaxAge:     28,
		Compress:   true,
		LocalTime:  true,
	}

	consoleConfig := zap.NewDevelopmentEncoderConfig()
	consoleConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(consoleConfig)

	jsonConfig := zap.NewProductionEncoderConfig()
	jsonConfig.TimeKey = "time"
	jsonConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	jsonEncoder := zapcore.NewJSONEncoder(jsonConfig)

	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), GetLogLevel(level))
	fileCore := zapcore.NewCore(jsonEncoder, zapcore.AddSync(lumberjackLogger), GetLogLevel(level))

	core := zapcore.NewTee(consoleCore, fileCore)

	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	baseLogger := zap.New(core).With(
		zap.String("app", config.AppName),
		zap.String("env", config.AppEnv),
		zap.String("hostname", hostname),
	)
	log = baseLogger.Sugar()
}

func Sync() {
	if log != nil {
		_ = log.Sync()
	}
}

func Info(msg string, keysAndValues ...interface{}) {
	keysAndValues = append(FileLog(), keysAndValues...)
	log.Infow(msg, keysAndValues...)
}

func Error(msg string, keysAndValues ...interface{}) {
	keysAndValues = append(FileLog(), keysAndValues...)
	log.Errorw(msg, keysAndValues...)
}

func Debug(msg string, keysAndValues ...interface{}) {
	if log.Level() > zap.DebugLevel {
		return
	}

	keysAndValues = append(FileLog(), keysAndValues...)
	log.Debugw(msg, keysAndValues...)
}

func Fatal(msg string, keysAndValues ...interface{}) {
	keysAndValues = append(FileLog(), keysAndValues...)
	log.Fatalw(msg, keysAndValues...)
}

func InfoL(msg string, meta metadata.MD, data ...DT) {
	dt := map[string]any{}
	if len(data) > 0 {
		dt = data[0]
	}

	keysAndValues := append(FileLog(), "data", DataLogger{Meta: meta, Data: dt, TransactionId: util.GetMetaTransactionId(meta)})
	log.Infow(msg, keysAndValues...)
}

func ErrorL(msg string, meta metadata.MD, data ...DT) {
	dt := map[string]any{}
	if len(data) > 0 {
		dt = data[0]
	}
	keysAndValues := append(FileLog(), "data", DataLogger{Meta: meta, Data: dt, TransactionId: util.GetMetaTransactionId(meta)})
	log.Errorw(msg, keysAndValues...)
}

func DebugL(msg string, meta metadata.MD, data ...DT) {
	if log.Level() > zap.DebugLevel {
		return
	}

	dt := map[string]any{}
	if len(data) > 0 {
		dt = data[0]
	}
	keysAndValues := append(FileLog(), "data", DataLogger{Meta: meta, Data: dt, TransactionId: util.GetMetaTransactionId(meta)})
	log.Debugw(msg, keysAndValues...)
}

func InfoF(msg string, args ...any) {
	log.Infow(fmt.Sprintf(msg, args...), FileLog()...)
}

func FatalF(msg string, args ...any) {
	log.Fatalw(fmt.Sprintf(msg, args...), FileLog()...)
}

func ErrorF(msg string, args ...any) {
	log.Errorw(fmt.Sprintf(msg, args...), FileLog()...)
}

func DebugF(msg string, args ...any) {
	if log.Level() > zap.DebugLevel {
		return
	}
	log.Debugw(fmt.Sprintf(msg, args...), FileLog()...)
}

func FileLog() []any {
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		return []any{}
	}
	fileName := filepath.Base(file)
	fn := runtime.FuncForPC(pc)
	funcName := fn.Name()
	funcName = funcName[strings.LastIndex(funcName, ".")+1:]
	parts := strings.Split(fn.Name(), ".")
	structAndMethod := getStructAndMethodName(fileName, parts)

	msg := fmt.Sprintf("%s:%d - %s()", fileName, line, structAndMethod)

	return []any{"caller", msg}
}

func getStructAndMethodName(fileName string, parts []string) string {
	if len(parts) == 0 {
		return ""
	}

	if len(parts) == 1 {
		return parts[0]
	}

	if fmt.Sprintf("%s.go", parts[0]) == fileName {
		return parts[1]
	}

	structAndMethod := strings.Join(parts[len(parts)-2:], ".")
	structAndMethod = strings.ReplaceAll(structAndMethod, "(", "")
	structAndMethod = strings.ReplaceAll(structAndMethod, "*", "")
	structAndMethod = strings.ReplaceAll(structAndMethod, ")", "")
	return structAndMethod
}

func GetLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zap.DebugLevel
	case "error":
		return zap.ErrorLevel
	default:
		return zap.InfoLevel
	}
}
