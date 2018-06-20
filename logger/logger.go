package logger

import (
	"github.com/sirupsen/logrus"
	"os"
)

var Logging = logrus.New()

func init() {
	// 以JSON格式为输出，代替默认的ASCII格式
	//Logging.Formatter = new(logrus.JSONFormatter)
	Logging.Formatter = new(logrus.TextFormatter)
	// 以Stdout为输出，代替默认的stderr
	Logging.Out = os.Stdout
	// 设置日志等级
	Logging.Level = logrus.InfoLevel
	// 删除时间戳
	//Logging.Formatter.(*logrus.TextFormatter).DisableTimestamp = true
	//Logging.Formatter.(*logrus.JSONFormatter).DisableTimestamp = true
}
