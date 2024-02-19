package consts

import "fmt"

const (
	LoggerCtxName = "logger"
	LanguagesLen = 2
	WaitForContainer = 2
	PeriodForContainerCheck = 30
)

// TODO make images in second option
// Name of language and it image in docker
var Languages = map[string]string{
	"js": "",
	"golang": "golang:1.21",
}


var (
	ErrorChannelFull = fmt.Errorf("channel is full")
	ErrorWaitingContainer = fmt.Errorf("can't get container in timeout")
	ErrorTimer = fmt.Errorf("timer occur")
)

const (
	Port = 3000
)