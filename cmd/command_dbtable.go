package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// dbTableCmd represents the shell command.
var dbTableCmd = &cobra.Command{
	Use:   "dbTable PATH TABLE",
	Short: "Export badger table.",
	Long:  `Export data table out of the badger database of step-ca. For list of tables see: https://raw.githubusercontent.com/smallstep/certificates/master/db/db.go`,

	Example: "  step-badger dbTable ./db ssh_host_principals",

	Args: cobra.ExactArgs(2),

	Run: func(cmd *cobra.Command, args []string) {
		dbTableMain(args)
	},
}

// Cobra initiation.
func init() {
	rootCmd.AddCommand(dbTableCmd)

	// Hide help command.
	dbTableCmd.SetHelpCommand(&cobra.Command{Hidden: true})
}

/*
dbTable main function.

	'args' Given command line arguments, that contain the command to be run by shell.
*/
func dbTableMain(args []string) {

	checkLogginglevel(args)

	var thisDB DB

	// Open the database.
	err := thisDB.Open(args[0])
	if err != nil {
		logError.Panic(err)
	}

	// Get records from the bucket.
	thisRecords, err := thisDB.List([]byte(args[1]))
	if err != nil {
		logError.Panic(err)
	}

	// Show info.
	if loggingLevel >= 2 {
		for _, thisRecord := range thisRecords {
			logInfo.Printf("Bucket: %s", thisRecord.Bucket)
			logInfo.Printf("Key: %s", thisRecord.Key)
			logInfo.Printf("Value: %s", thisRecord.Value)
		}
	}

	// Marshal into json.
	thisJson, err := json.MarshalIndent(thisRecords, "", "  ")
	if err != nil {
		logError.Panic(err)
	}

	// Emit.
	fmt.Println(string(thisJson))

	// Show info.
	if loggingLevel >= 2 {
		logInfo.Printf("%d records marshalled.\n", len(thisRecords))
	}

}
