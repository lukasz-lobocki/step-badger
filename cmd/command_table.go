package cmd

import (
	"strings"

	"github.com/dgraph-io/badger/v2"
	"github.com/spf13/cobra"
)

// tableCmd represents the shell command
var tableCmd = &cobra.Command{
	Use:   "table [PATH] \"command\"",
	Short: "Export table.",
	Long:  `Export table data out of the badger database of step-ca.`,

	Example: "  gitas table XXXXX -duda",

	Args: cobra.RangeArgs(1, 2),

	Run: func(cmd *cobra.Command, args []string) {
		tableMain(args)
	},
}

// Cobra initiation
func init() {
	rootCmd.AddCommand(tableCmd)

	initChoices()

	tableCmd.Flags().VarP(config.emitFormat, "emit", "e", "emit format: table|json|markdown") // Choice
}

/*
table main function

	'args' given command line arguments, that contain the command to be run by shell
*/
func tableMain(args []string) {
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

	// retrieveCerts(db)
	// retrieveTableData(db, []byte("x509_certs_data"), "/dev/stdout")
	// retrieveTableData(db, []byte("admins"), "/dev/stdout")
	// retrieveTableData(db, []byte("provisioners"), "/dev/stdout")
	// retrieveTableData(db, []byte("authority_policies"), "/dev/stdout")
	retrieveTableData(db, []byte("x509_certs_data"))
	retrieveTableData(db, []byte("revoked_x509_certs"))

}

func retrieveTableData(db *badger.DB, prefix []byte) {
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
