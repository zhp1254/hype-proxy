package logger

import (
	"fmt"
	"go.uber.org/zap"
)

var (
	logInstance *zap.Logger
)

func init(){
	logInstance = zapLogger()
}

func zapLogger() *zap.Logger {
	cfg := zap.NewDevelopmentConfig()
	cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)

	logger, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic(fmt.Sprintf("failed to initialize zap logger: %v", err))
	}
	return logger
}

func GetLogger() *zap.Logger{
	return logInstance
}

func Infof(format string, args ...interface{}){
	logInstance.Info(fmt.Sprintf(format, args...))
}


func Debugf(format string, args ...interface{}){
	logInstance.Debug(fmt.Sprintf(format, args...))
}

func Errorf(format string, args ...interface{}){
	logInstance.Error(fmt.Sprintf(format, args...))
}

func Warnf(format string, args ...interface{}){
	logInstance.Warn(fmt.Sprintf(format, args...))
}

func Info(msg string){
	logInstance.Info(msg)
}


func Debug(msg string){
	logInstance.Debug(msg)
}

func Error(msg string){
	logInstance.Error(msg)
}

func Warn(msg string){
	logInstance.Warn(msg)
}

func Fatalln(format string, args ...interface{}) {
	logInstance.Fatal(fmt.Sprintf(format, args...))
}

func Sync(){
	logInstance.Sync()
}