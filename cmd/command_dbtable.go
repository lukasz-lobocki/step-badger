package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// dbTableCmd represents the shell command.
var dbTableCmd = &cobra.Command{
	Short:   "Export badger table.",
	Run:     func(cmd *cobra.Command, args []string) { dbTableMain(args) },
	Args:    cobra.ExactArgs(2),
	Aliases: []string{"dbtable"},

	DisableFlagsInUseLine: true,

	Example: "  step-badger dbTable ./db ssh_host_principals",
	Long: `
Export data table out of the badger database of step-ca.`,
	Use: `dbTable <PATH> <TABLE> [flags]

Arguments:
  PATH    location of the source database
  TABLE   name of Badger table to export

Note:
  For list of tables see: https://raw.githubusercontent.com/smallstep/certificates/master/db/db.go`,
}

// Cobra initiation.
func init() {
	rootCmd.AddCommand(dbTableCmd)

	// Hide help command.
	dbTableCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	//Do not sort flags.
	dbTableCmd.Flags().SortFlags = false
}

/*
dbTable main function.

	'args' Given command line arguments, that contain the command to be run by shell.
*/
func dbTableMain(args []string) {

	checkLogginglevel(args)

	// Open the database.
	db := openDB(args[0])

	// Get records from the bucket.
	records := listBucket(db, args[1])

	// Close the database.
	closeDB(db, args[0])

	if loggingLevel >= 3 { // Show info.
		for _, record := range records {
			logInfo.Printf("Bucket: %s", record.Bucket)
			logInfo.Printf("Key: %s", record.Key)
			logInfo.Printf("Value: %q", record.Value)
		}
	}

	// Marshal into json.
	jsonInfo, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		logError.Panic(err)
	}

	// Emit.
	fmt.Println(string(jsonInfo))

	if loggingLevel >= 2 { // Show info.
		logInfo.Printf("%d records marshalled.\n", len(records))
	}

}
