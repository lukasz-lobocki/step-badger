package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

/*
Version number shown in help message. `version` is updated with `-ldflags` during compilation.

	sem_release_ver+architecture.short_git_hash[.dirty.build_date]
*/
var (
	semVer     string
	commitHash string
	goArch     string

	semReleaseVersion string = semVer +
		func(pre string, txt string) string {
			if len(txt) > 0 {
				return pre + txt
			} else {
				return ""
			}
		}("+", goArch) +
		func(pre string, txt string) string {
			if len(txt) > 0 {
				return pre + txt
			} else {
				return ""
			}
		}(".", commitHash)
)

var config tConfig // Holds configuration.

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:     "step-badger",
	Short:   "Export step-ca data from badger.",
	Long:    `Export certificate or table data from the badger database of step-ca. Requires off-line database directory.`,
	Version: semReleaseVersion,

	CompletionOptions: cobra.CompletionOptions{HiddenDefaultCmd: true},

	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

var (
	loggingLevel int         // Global logging level, see MAX_LOGGING_LEVEL.
	logInfo      *log.Logger // Blue logger, for info.
	logWarning   *log.Logger // Yellow logger, for warning.
	logError     *log.Logger // Red logger, for error.
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	initLoggers()
	initChoices()

	// Hide help command.
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	//Do not sort flags.
	rootCmd.Flags().SortFlags = false

	// Adding global ie. persistent logging level flag.
	rootCmd.PersistentFlags().IntVar(&loggingLevel, "logging", 0,
		fmt.Sprintf("logging level [0...%d] (default 0)", MAX_LOGGING_LEVEL))
}

/*
checkLogginglevel confirms if logging level does not exceed maximum level.

	loggingLevel = 1 : often
	loggingLevel = 2 : average
	loggingLevel = 3 : seldom

For convenience it also emits some log if loggingLevel >= 1.

	'thisArgs' Values emitted to log.
*/
func checkLogginglevel(thisArgs []string) {
	if loggingLevel > MAX_LOGGING_LEVEL {
		logError.Fatalln(fmt.Errorf("%s", rootCmd.Flag("logging").Usage))
	}

	if loggingLevel >= 1 { // Show info.
		logInfo.Printf("len(args): %d. args: %#v\n", len(thisArgs), thisArgs)
		logInfo.Printf("loggingLevel: %d. config: %#v\n", loggingLevel, config)
	}
}
