package consts

import "fmt"

const (
	loggerCtxName = "logger"
	LanguagesLen = 2
	WaitForContainer = 2
	PeriodForContainerCheck = 30
)

// TODO make images in second option
// Name of language and it image in docker
var languages = map[string]string{
	"js": "",
	"golang": "",
}


var (
	ErrorChannelFull = fmt.Errorf("channel is full")
)