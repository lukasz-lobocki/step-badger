package cmd

import (
	"log"
	"os"

	"github.com/fatih/color"
)

const (
	MAX_LOGGING_LEVEL int = 3 // Maximum allowed logging level
)

/*
initLoggers creates colorful loggers.
*/
func initLoggers() {
	thisHiCyan := color.New(color.FgHiCyan).SprintFunc()
	thisHiYellow := color.New(color.FgHiYellow).SprintFunc()
	thisHiRed := color.New(color.FgHiRed).SprintFunc()

	logInfo = log.New(os.Stderr, thisHiCyan("╭info\n╰"), 0)
	logWarning = log.New(os.Stderr, thisHiYellow("╭warning\n╰"), log.Lshortfile)
	logError = log.New(os.Stderr, thisHiRed("╭error\n╰"), log.Lshortfile)

}

/*
initChoices sets up Config struct for 'limited choice' flag
*/
func initChoices() {
	config.emitFormat = newChoice([]string{"t", "j"}, "t")
	config.sortOrder = newChoice([]string{"s", "f"}, "f")
}

/*
Configuration structure
*/
type tConfig struct {
	emitFormat *tChoice
	showCrl    bool
	showKeyId  bool
	sortOrder  *tChoice
}

/*
Alignment of markdown table
*/
const (
	ALIGN_LEFT = iota
	ALIGN_CENTER
	ALIGN_RIGHT
)
