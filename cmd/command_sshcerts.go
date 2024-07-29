package cmd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

// sshCertsCmd represents the shell command.
var sshCertsCmd = &cobra.Command{
	Use:   "sshCerts PATH",
	Short: "Export ssh certificates.",
	Long:  `Export ssh certificates' data out of the badger database of step-ca.`,

	Example: "  step-badger ssCerts ./db",

	Args: cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		exportSshMain(args)
	},
}

/*
Cobra initiation.
*/
func init() {
	rootCmd.AddCommand(sshCertsCmd)

	// Hide help command.
	sshCertsCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	//Do not sort flags.
	sshCertsCmd.Flags().SortFlags = false

	sshCertsCmd.Flags().VarP(config.emitSshFormat, "emit", "e", "emit format: table|json|markdown") // Choice
	sshCertsCmd.Flags().VarP(config.timeFormat, "time", "t", "time format: iso|short")              // Choice
	sshCertsCmd.Flags().VarP(config.sortOrder, "sort", "s", "sort order: start|finish")             // Choice
	sshCertsCmd.Flags().BoolVarP(&config.showKeyId, "kid", "k", false, "Key ID column shown")
	sshCertsCmd.Flags().BoolVarP(&config.showValid, "valid", "v", true, "valid certificates shown")
	sshCertsCmd.Flags().BoolVarP(&config.showRevoked, "revoked", "r", true, "revoked certificates shown")
	sshCertsCmd.Flags().BoolVarP(&config.showExpired, "expired", "x", false, "expired certificates shown")
}

/*
ExportSsh main function.

	'args' Given command line arguments, that contain the command to be run by shell.
*/
func exportSshMain(args []string) {

	checkLogginglevel(args)

	db, err := badger.Open(badger.DefaultOptions(args[0]).WithLogger(nil))
	if err != nil {
		logError.Panic(err)
	}
	defer db.Close()

	// Get.
	sshCerts := getSshCerts(db)

	// Sort.
	switch thisSort := config.sortOrder.Value; thisSort {
	case "f":
		sort.SliceStable(sshCerts, func(i, j int) bool {
			return sshCerts[i].SshCertificate.ValidBefore < sshCerts[j].SshCertificate.ValidBefore
		})
	case "s":
		sort.SliceStable(sshCerts, func(i, j int) bool {
			return sshCerts[i].SshCertificate.ValidAfter < sshCerts[j].SshCertificate.ValidAfter
		})
	}

	// Output.
	switch thisFormat := config.emitSshFormat.Value; thisFormat {
	case "j":
		emitSshCertsJson(sshCerts)
	case "t":
		emitSshCertsTable(sshCerts)
	case "m":
		emitSshCertsMarkdown(sshCerts)
	}
}

/*
getSshCerts returns struct with ssh certificates.

	'thisDb' Badger database.
*/
func getSshCerts(thisDb *badger.DB) []tSshCertificateWithRevocation {
	var (
		sshCertsWithRevocations []tSshCertificateWithRevocation = []tSshCertificateWithRevocation{}
	)

	prefix, err := badgerEncode([]byte("ssh_certs"))
	if err != nil {
		logError.Panic(err)
	}

	txn := thisDb.NewTransaction(false)
	defer txn.Discard()

	iter := txn.NewIterator(badger.DefaultIteratorOptions)
	defer iter.Close()

	for iter.Seek(prefix); iter.ValidForPrefix(prefix); iter.Next() {
		var (
			sshCertsWithRevocation tSshCertificateWithRevocation = tSshCertificateWithRevocation{}
		)

		sshCert, err := getSshCertificate(iter)
		if err != nil {
			continue
		}

		// Populate child main info of the certificate.
		sshCertsWithRevocation.SshCertificate = sshCert

		// Populate child revocation info of the certificate.
		sshCertsWithRevocation.SshRevocation = getSshRevocationData(thisDb, &sshCert)

		// Populate child validity info of the certificate.
		if len(sshCertsWithRevocation.SshRevocation.ProvisionerID) > 0 && time.Now().After(sshCertsWithRevocation.SshRevocation.RevokedAt) {
			sshCertsWithRevocation.Validity = REVOKED_STR
		} else {
			if time.Now().After(time.Unix(int64(sshCertsWithRevocation.SshCertificate.ValidBefore), 0)) {
				sshCertsWithRevocation.Validity = EXPIRED_STR
			} else {
				sshCertsWithRevocation.Validity = VALID_STR
			}
		}

		// Append child into collection, if record selection criteria are met.
		if (config.showExpired && sshCertsWithRevocation.Validity == EXPIRED_STR) ||
			(config.showRevoked && sshCertsWithRevocation.Validity == REVOKED_STR) ||
			(config.showValid && sshCertsWithRevocation.Validity == VALID_STR) {
			sshCertsWithRevocations = append(sshCertsWithRevocations, sshCertsWithRevocation)
		}

	}

	return sshCertsWithRevocations

}

/*
getSshCertificate returns ssh certificate.
*/
func getSshCertificate(iter *badger.Iterator) (ssh.Certificate, error) {
	item := iter.Item()

	var (
		valCopy []byte
	)

	valCopy, err := item.ValueCopy(nil)
	if err != nil {
		logError.Panicf("Error parsing item value: %v", err)
	}

	if len(strings.TrimSpace(string(valCopy))) == 0 {
		// Item is empty.
		return ssh.Certificate{}, fmt.Errorf("empty")
	} else {

		// Parse the SSH certificate.
		pubKey, err := ssh.ParsePublicKey(valCopy)
		if err != nil {
			logError.Panicf("Error parsing SSH certificate: %v", err)
		}

		cert, ok := pubKey.(*ssh.Certificate)
		if !ok {
			logError.Panicf("Key is not an SSH certificate")
		}

		return *cert, nil
	}

}

/*
getSshRevocationData returns revocation information for a given certificate, if exists.

	'thisDb' Badger database.
	'thisCert' Certificate to get the revocation information for.
*/
func getSshRevocationData(thisDb *badger.DB, thisCert *ssh.Certificate) tRevokedCertificate {
	var item *badger.Item
	var data tRevokedCertificate = tRevokedCertificate{}

	item, err := getItem(thisDb, []byte("revoked_ssh_certs"), []byte(strconv.FormatUint(thisCert.Serial, 10)))
	if err != nil {
		// Skip errors (like not found).
	} else {
		// Found a revoked cert.
		var valCopy []byte
		valCopy, err = item.ValueCopy(nil)
		if err != nil {
			logError.Panic(err)
		}

		if len(strings.TrimSpace(string(valCopy))) > 0 {
			if err := json.Unmarshal(valCopy, &data); err != nil {
				logError.Panic(err)
			}
		}
	}
	return data
}
