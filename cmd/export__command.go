package cmd

import (
	"bytes"
	"encoding/binary"
	"strings"

	"github.com/dgraph-io/badger/v2"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// exportCmd represents the shell command
var exportCmd = &cobra.Command{
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
	rootCmd.AddCommand(exportCmd)

	initChoices()

	exportCmd.Flags().VarP(config.emitFormat, "emit", "e", "emit format: table|json|markdown") // Choice
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

	db, err := badger.Open(badger.DefaultOptions(args[0]).WithLogger(nil))
	if err != nil {
		logInfo.Println("BULWA")
		panic(err)
	}
	defer db.Close()

	// retrieveCerts(db, []byte("x509_certs"), "/dev/stdout")
	// retrieveTableData(db, []byte("x509_certs_data"), "/dev/stdout")
	// retrieveTableData(db, []byte("admins"), "/dev/stdout")
	// retrieveTableData(db, []byte("provisioners"), "/dev/stdout")
	// retrieveTableData(db, []byte("authority_policies"), "/dev/stdout")
	retrieveTableData(db, []byte("revoked_x509_certs"), "/dev/stdout")

}

func retrieveTableData(db *badger.DB, prefix []byte, filename string) {
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

	/* f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close() */

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

		logInfo.Printf("\nkey=%s ::\nvalue=%s\n", strings.TrimSpace(string(item.Key())), strings.TrimSpace(string(valCopy)))

	}
}

// badgerEncode encodes a byte slice into a section of a BadgerKey.
// See documentation for toBadgerKey.
func badgerEncode(val []byte) ([]byte, error) {
	l := len(val)
	switch {
	case l == 0:
		return nil, errors.Errorf("input cannot be empty")
	case l > 65535:
		return nil, errors.Errorf("length of input cannot be greater than 65535")
	default:
		lb := new(bytes.Buffer)
		if err := binary.Write(lb, binary.LittleEndian, uint16(l)); err != nil {
			return nil, errors.Wrap(err, "error doing binary Write")
		}
		return append(lb.Bytes(), val...), nil
	}
}

func toBadgerKey(bucket, key []byte) ([]byte, error) {
	first, err := badgerEncode(bucket)
	if err != nil {
		return nil, err
	}
	second, err := badgerEncode(key)
	if err != nil {
		return nil, err
	}
	return append(first, second...), nil
}
