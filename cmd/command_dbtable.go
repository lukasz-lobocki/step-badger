package cmd

import (
	"strings"

	"github.com/dgraph-io/badger/v2"
	"github.com/spf13/cobra"
)

// dbTableCmd represents the shell command
var dbTableCmd = &cobra.Command{
	Use:   "dbTable [PATH] \"command\"",
	Short: "Export table.",
	Long:  `Export table data out of the badger database of step-ca.`,

	Example: "  gitas table XXXXX -duda",

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

	db, err := badger.Open(badger.DefaultOptions(args[0]).WithLogger(nil))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	retrieveDbTableData(db, []byte("ssh_hosts"))
	retrieveDbTableData(db, []byte("ssh_host_principals"))
	retrieveDbTableData(db, []byte("ssh_users"))
	retrieveDbTableData(db, []byte("ssh_certs"))

}

func retrieveDbTableData(db *badger.DB, prefix []byte) {
	txn := db.NewTransaction(false)
	defer txn.Discard()

	prefix, err := badgerEncode(prefix)
	if err != nil {
		panic(err)
	}
	logInfo.Printf("Encoded table prefix: %s", string(prefix))

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

		logInfo.Printf("key=%s ::\nvalue=%s", strings.TrimSpace(string(item.Key())), strings.TrimSpace(string(valCopy)))
	}
}
