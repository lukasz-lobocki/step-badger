/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
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
)

var semReleaseVersion string = semVer +
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

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "step-badger",
	Short:   "A brief description of your application",
	Long:    `A longer description that spans multiple lines and likely contains`,
	Version: semReleaseVersion,

	CompletionOptions: cobra.CompletionOptions{HiddenDefaultCmd: true},

	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

var (
	loggingLevel int         // Global logging level, see MAX_LOGGING_LEVEL
	logInfo      *log.Logger // Blue logger, for info
	logWarning   *log.Logger // Yellow logger, for warning
	logError     *log.Logger // Red logger, for error
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
	// Hiding help command
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	rootCmd.Flags().SortFlags = false

	// Adding global ie. persistent logging level flag
	rootCmd.PersistentFlags().IntVar(&loggingLevel, "logging", 0,
		fmt.Sprintf("logging level [0...%d] (default 0)", MAX_LOGGING_LEVEL))

	/* Init loggers */

	thisHiCyan := color.New(color.FgHiCyan).SprintFunc()
	thisHiYellow := color.New(color.FgHiYellow).SprintFunc()
	thisHiRed := color.New(color.FgHiRed).SprintFunc()

	logInfo = log.New(os.Stderr, thisHiCyan("╭info\n╰"), 0)
	logWarning = log.New(os.Stderr, thisHiYellow("╭warning\n╰"), log.Lshortfile)
	logError = log.New(os.Stderr, thisHiRed("╭error\n╰"), log.Lshortfile)

}
