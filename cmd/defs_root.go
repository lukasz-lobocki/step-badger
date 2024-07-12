package cmd

const (
	MAX_LOGGING_LEVEL int = 3 // Maximum allowed logging level
)

/*
initChoices sets up Config struct for 'limited choice' flag
*/
func initChoices() {
	config.emitFormat = newChoice([]string{"t", "j", "m"}, "t")
	config.sortOrder = newChoice([]string{"v", "b"}, "v")
}

/*
Configuration structure
*/
type tConfig struct {
	emitFormat *tChoice
	sortOrder  *tChoice
}
