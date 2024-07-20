package cmd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/dgraph-io/badger/v2"
	"github.com/fatih/color"
	"github.com/lukasz-lobocki/tabby"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

// sshCertsCmd represents the shell command
var sshCertsCmd = &cobra.Command{
	Use:   "sshCerts BADGERPATH",
	Short: "Export ssh certificates.",
	Long:  `Export ssh certificates' data out of the badger database of step-ca.`,

	Example: "  step-badger ssCerts ./db",

	Args: cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		exportSshMain(args)
	},
}

// Cobra initiation
func init() {
	rootCmd.AddCommand(sshCertsCmd)

	//initChoices()

	sshCertsCmd.Flags().VarP(config.emitFormat, "emit", "e", "emit format: table|json") // Choice
	logInfo.Printf(config.emitFormat.Value)
}

/*
Export main function

	'args' given command line arguments, that contain the command to be run by shell
*/
func exportSshMain(args []string) {

	checkLogginglevel(args)

	db, err := badger.Open(badger.DefaultOptions(args[0]).WithLogger(nil))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	sshCerts := retrieveSshCerts(db)

	sort.SliceStable(sshCerts, func(i, j int) bool {
		return sshCerts[i].ValidBefore < sshCerts[j].ValidBefore
	})

	switch thisFormat := config.emitFormat.Value; thisFormat {
	case "j":
		emitSshCertsJson(sshCerts)
	case "t":
		emitSshCertsTable(sshCerts)
	}
}

func retrieveSshCerts(db *badger.DB) []ssh.Certificate {
	var (
		sshCerts []ssh.Certificate = []ssh.Certificate{}
	)

	prefix, err := badgerEncode([]byte("ssh_certs"))
	if err != nil {
		logError.Panic(err)
	}

	txn := db.NewTransaction(false)
	defer txn.Discard()

	iter := txn.NewIterator(badger.DefaultIteratorOptions)
	defer iter.Close()

	for iter.Seek(prefix); iter.ValidForPrefix(prefix); iter.Next() {

		sshCert, err := getSshCertificate(iter)
		if err != nil {
			continue
		}

		sshCerts = append(sshCerts, sshCert)

	}

	return sshCerts

}

func emitSshCertsTable(sshCerts []ssh.Certificate) {
	table := new(tabby.Table)

	thisColumns := getSshColumns()

	var thisHeader []string
	/* Building slice of titles */
	for _, thisColumn := range thisColumns {
		if thisColumn.isShown() {
			thisHeader = append(thisHeader,
				color.New(thisColumn.titleColor).SprintFunc()(
					thisColumn.title(),
				),
			)
		}
	}

	/* Set the header */

	if err := table.SetHeader(thisHeader); err != nil {
		logError.Panic("Setting header failed. %w", err)
	}

	if loggingLevel >= 1 {
		logInfo.Println("header set.")
	}

	/* Populate the table */

	for _, sshCert := range sshCerts {

		var thisRow []string
		/* Building slice of columns within a single row*/

		for _, thisColumn := range thisColumns {

			if thisColumn.isShown() {
				thisRow = append(thisRow,
					color.New(thisColumn.contentColor(sshCert)).SprintFunc()(
						thisColumn.contentSource(sshCert),
					),
				)
			}
		}

		if err := table.AppendRow(thisRow); err != nil {
			logError.Panic(err)
		}
		if loggingLevel >= 3 {
			logInfo.Printf("row [%s] appended.", string(sshCert.Serial))
		}

	}

	if loggingLevel >= 2 {
		logInfo.Printf("%d rows appended.\n", len(sshCerts))
	}

	/* Emit the table */

	if loggingLevel >= 3 {
		table.Print(&tabby.Config{Spacing: "|", Padding: "."})
	} else {
		table.Print(nil)
	}
}

func getSshCertificate(iter *badger.Iterator) (ssh.Certificate, error) {
	item := iter.Item()

	var (
		valCopy []byte
	)

	valCopy, err := item.ValueCopy(nil)
	if err != nil {
		logError.Fatalf("Error parsing item value: %v", err)
	}

	if len(strings.TrimSpace(string(valCopy))) == 0 {
		// Item is empty
		return ssh.Certificate{}, fmt.Errorf("empty")
	} else {

		// Parse the SSH certificate
		pubKey, err := ssh.ParsePublicKey(valCopy)
		if err != nil {
			logError.Fatalf("Error parsing SSH certificate: %v", err)
		}

		cert, ok := pubKey.(*ssh.Certificate)
		if !ok {
			logError.Fatalf("Key is not an SSH certificate")
		}

		return *cert, nil
	}

}

func emitSshCertsJson(sshCerts []ssh.Certificate) {
	jsonInfo, err := json.MarshalIndent(sshCerts, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonInfo))
	if loggingLevel >= 2 {
		logInfo.Printf("%d records marshalled.\n", len(sshCerts))
	}
}
