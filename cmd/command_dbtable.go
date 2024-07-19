package cmd

import (
	"strings"

	"github.com/dgraph-io/badger/v2"
	"github.com/spf13/cobra"
)

// dbTableCmd represents the shell command
var dbTableCmd = &cobra.Command{
	Use:   "dbTable [PATH]",
	Short: "Export table.",
	Long:  `Export data table out of the badger database of step-ca.`,

	Example: "  step-badger dbTable ./db",

	Args: cobra.RangeArgs(1, 2),

	Run: func(cmd *cobra.Command, args []string) {
		dbTableMain(args)
	},
}

// Cobra initiation
func init() {
	rootCmd.AddCommand(dbTableCmd)

	initChoices()

	dbTableCmd.Flags().SortFlags = false
	dbTableCmd.Flags().VarP(config.emitFormat, "emit", "e", "emit format: table|json|markdown") // Choice
}

/*
table main function

	'args' given command line arguments, that contain the command to be run by shell
*/
func dbTableMain(args []string) {

	checkLogginglevel(args)

	db, err := badger.Open(badger.DefaultOptions(args[0]).WithLogger(nil))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	retrieveDbTableData(db, []byte(args[1]))

}

func retrieveDbTableData(db *badger.DB, prefix []byte) {
	txn := db.NewTransaction(false)
	defer txn.Discard()

	prefix, err := badgerEncode(prefix)
	if err != nil {
		panic(err)
	}

	if loggingLevel >= 2 {
		logInfo.Printf("Encoded table prefix: %s", string(prefix))
	}

	opts := badger.DefaultIteratorOptions
	iter := txn.NewIterator(opts)
	defer iter.Close()

	for iter.Seek(prefix); iter.ValidForPrefix(prefix); iter.Next() {
		item := iter.Item()

		var valCopy []byte
		valCopy, err = item.ValueCopy(nil)
		if err != nil {
			panic(err)
		}

		if len(strings.TrimSpace(string(valCopy))) == 0 {
			// Item is empty
			continue
		}
		if loggingLevel >= 3 {
			logInfo.Printf("[key=%s] [value=%s]", strings.TrimSpace(string(item.Key())), strings.TrimSpace(string(valCopy)))
		}
	}
}
