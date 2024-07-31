package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/smallstep/nosql"
	"github.com/smallstep/nosql/database"
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

	var (
		err error
		db  database.DB
	)

	checkLogginglevel(args)

	// Open the database.
	db, err = nosql.New("badgerv2", args[0], database.WithValueDir(args[0]))
	if err != nil {
		logError.Fatalln(err)
	}

	// Get records from the bucket.
	records, err := db.List([]byte(args[1]))
	if err != nil {
		logError.Fatalln(err)
	}
	if records == nil {
		logError.Fatalln("no records found")
	}

	// Close the database.
	if err = db.Close(); err != nil {
		logError.Fatalln(err)
	}

	if loggingLevel >= 2 { // Show info.
		for _, record := range records {
			logInfo.Printf("Bucket: %s", record.Bucket)
			logInfo.Printf("Key: %s", record.Key)
			logInfo.Printf("Value: %q", record.Value)
		}
	}

	// Marshal into json.
	json, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		logError.Panic(err)
	}

	// Emit.
	fmt.Println(string(json))

	if loggingLevel >= 2 { // Show info.
		logInfo.Printf("%d records marshalled.\n", len(records))
	}

}
