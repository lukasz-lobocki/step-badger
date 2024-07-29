package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dgraph-io/badger/v2"
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

	db, err := badger.Open(badger.DefaultOptions(args[0]).WithLogger(nil))
	if err != nil {
		logError.Panic(err)
	}
	defer db.Close()

	dbRecords := retrieveDbTableData(db, []byte(args[1]))
	emitDbRecordsJson(dbRecords)
}

/*
retrieveDbTableData returns the structure containing data table from Badger database.

	'thisDb' Source database.
	'thisBucket' Name of the bucket.
*/
func retrieveDbTableData(thisDb *badger.DB, thisBucket []byte) []tDbRecord {
	var (
		dbRecords []tDbRecord = []tDbRecord{}
	)
	txn := thisDb.NewTransaction(false)
	defer txn.Discard()

	thisPrefix, err := badgerEncode(thisBucket)
	if err != nil {
		logError.Panic(err)
	}

	if loggingLevel >= 2 {
		logInfo.Printf("Encoded table prefix: %s", string(thisPrefix))
	}

	opts := badger.DefaultIteratorOptions
	iter := txn.NewIterator(opts)
	defer iter.Close()

	for iter.Seek(thisPrefix); iter.ValidForPrefix(thisPrefix); iter.Next() {
		var (
			dbRecord tDbRecord = tDbRecord{}
		)
		item := iter.Item()

		var valCopy []byte
		valCopy, err = item.ValueCopy(nil)
		if err != nil {
			logError.Panic(err)
		}

		if len(strings.TrimSpace(string(valCopy))) == 0 {
			// Item is empty
			continue
		}

		// Construct child key and value.
		dbRecord.Key = string(item.Key())
		dbRecord.Value = valCopy

		// Append child to the collection.
		dbRecords = append(dbRecords, dbRecord)

		if loggingLevel >= 3 {
			logInfo.Printf("[key=%s] [value=%s]", strings.TrimSpace(string(item.Key())), strings.TrimSpace(string(valCopy)))
		}
	}
	return dbRecords
}

/*
emitDbRecordsJson prints result in the form of a json.

	'thisDbRecords' Slice of structures describing the records.
*/
func emitDbRecordsJson(thisDbRecords []tDbRecord) {
	jsonInfo, err := json.MarshalIndent(thisDbRecords, "", "  ")
	if err != nil {
		logError.Panic(err)
	}
	fmt.Println(string(jsonInfo))
	if loggingLevel >= 2 {
		logInfo.Printf("%d records marshalled.\n", len(thisDbRecords))
	}
}
