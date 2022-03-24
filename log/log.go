package log

import "go.uber.org/zap"

var Logger *zap.SugaredLogger

func init() {
	prod, _ := zap.NewProduction()
	Logger = prod.Sugar()
}
