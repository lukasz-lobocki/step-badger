package cmd

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"

	"github.com/dgraph-io/badger/v2"
	"github.com/fatih/color"
	"github.com/pkg/errors"
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

var config tConfig // Holds configuration

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "step-badger",
	Short:   "Export step-ca data from badger.",
	Long:    `Export certificate or table data from the badger database of step-ca. Requires off-line database directory.`,
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
	initChoices()

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

/*
getItem function returns data item (if exists for the prefix) for the given key.

	'db' badger database.
	'prefix' badger prefix.
	'key' badger key.
*/
func getItem(db *badger.DB, prefix []byte, key []byte) (*badger.Item, error) {
	badgerKey, _ := toBadgerKey(prefix, key)

	txn := db.NewTransaction(false)
	defer txn.Discard()

	item, err := txn.Get(badgerKey)
	if err != nil {
		return nil, err
	}
	return item, nil
}

/*
badgerEncode function encodes a byte slice into a section of a BadgerKey.

	'val' given byte slice, that contains the key data.
*/
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

/*
toBadgerKey function encodes bucket and key into the BadgerKey.

	'bucket' given byte slice, that bucket name.
	'key' given byte slice, that key value.
*/
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

/*
checkLogginglevel confirms if logging level does not exceed maximum level.

For convenience it also emits some log

	'args' values emitted to log
*/
func checkLogginglevel(args []string) {
	if loggingLevel > MAX_LOGGING_LEVEL {
		logError.Fatalln(fmt.Errorf("%s", rootCmd.Flag("logging").Usage))
	}

	if loggingLevel >= 1 {
		logInfo.Printf("len(args): %d. args: %+v\n", len(args), args)
		logInfo.Printf("loggingLevel: %d. config: %+v\n", loggingLevel, config)
	}

	/*
		loggingLevel = 1 : often
		loggingLevel = 2 : average
		loggingLevel = 3 : seldom
	*/

}
