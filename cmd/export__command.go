package cmd

import (
	"github.com/spf13/cobra"
)

// shellCmd represents the shell command
var shellCmd = &cobra.Command{
	Use:   "export [PATH] \"command\"",
	Short: "Export certificates.",
	Long:  `Export certificates' data out of the badger database of step-ca.`,

	Example: "  gitas EXPORT XXXXX -duda",

	Args: cobra.RangeArgs(1, 2),

	Run: func(cmd *cobra.Command, args []string) {
		exportMain(args)
	},
}

var config tConfig // Holds status' configuration

// Cobra initiation
func init() {
	rootCmd.AddCommand(shellCmd)
}

/*
Shell main function

	'args' given command line arguments, that contain the command to be run by shell
*/
func exportMain(args []string) {
	/* 	var (
	   		cmdArgs  []string // Args of the command to execute
	   		givenDir string
	   		err      error
	   	)
	*/
	checkLogginglevel(args)

	logInfo.Println(args[0])

	/* Construct arguments */

	/*
		 	switch lenArgs := len(args); lenArgs {
			case 1:
				givenDir = "."
				cmdArgs = append([]string{"-c"}, args[0])
			case 2:
				givenDir = args[0]
				cmdArgs = append([]string{"-c"}, args[1])
			}
	*/
}
