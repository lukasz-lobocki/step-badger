package cmd

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

// sshCertsCmd represents the shell command
var sshCertsCmd = &cobra.Command{
	Use:   "sshCerts [PATH] \"command\"",
	Short: "Export certificates.",
	Long:  `Export certificates' data out of the badger database of step-ca.`,

	Example: "  gitas x509certs XXXXX -duda",

	Args: cobra.RangeArgs(1, 2),

	Run: func(cmd *cobra.Command, args []string) {
		exportSshMain(args)
	},
}

// Cobra initiation
func init() {
	rootCmd.AddCommand(sshCertsCmd)

	initChoices()

	sshCertsCmd.Flags().SortFlags = false
	sshCertsCmd.Flags().VarP(config.sortOrder, "order", "o", "order: validity|before")           // Choice
	sshCertsCmd.Flags().VarP(config.emitFormat, "emit", "e", "emit format: table|json|markdown") // Choice
}

/*
Export main function

	'args' given command line arguments, that contain the command to be run by shell
*/
func exportSshMain(args []string) {

	checkLogginglevel(args)

	logInfo.Println(args[0])

	db, err := badger.Open(badger.DefaultOptions(args[0]).WithLogger(nil))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	retrieveSshCerts(db)
}

func retrieveSshCerts(db *badger.DB) {
	/* var (
		sshCertsWithRevocations []X509CertificateAndRevocationInfo = []X509CertificateAndRevocationInfo{}
	) */

	prefix, err := badgerEncode([]byte("ssh_certs"))
	if err != nil {
		panic(err)
	}

	txn := db.NewTransaction(false)
	defer txn.Discard()

	iter := txn.NewIterator(badger.DefaultIteratorOptions)
	defer iter.Close()

	for iter.Seek(prefix); iter.ValidForPrefix(prefix); iter.Next() {
		// var (
		// 	sshCertAndRevocation X509CertificateAndRevocationInfo = X509CertificateAndRevocationInfo{}
		// )

		sshCert, err := getSshCertificate(iter)
		if err != nil {
			continue
		}

		logInfo.Printf("sn %v\n", sshCert.Serial)
		logInfo.Printf("keyid %+v\n", sshCert.KeyId)
		logInfo.Printf("valprinc %+v\n", sshCert.ValidPrincipals)
		logInfo.Printf("after %+v\n", time.Unix(int64(sshCert.ValidAfter), 0).UTC())
		logInfo.Printf("before %+v\n", time.Unix(int64(sshCert.ValidBefore), 0).UTC())
		// Populate main info of the certificate.
		// sshCertAndRevocation.X509Certificate = sshCert

		// Populate revocation info of the certificate.
		// sshCertAndRevocation.X509Revocation = getSshRevocationData(db, &sshCert)

		// sshCertsWithRevocations = append(sshCertsWithRevocations, sshCertAndRevocation)
	}

}

func getSshCertificate(iter *badger.Iterator) (ssh.Certificate, error) {
	item := iter.Item()

	var (
		valCopy []byte
	)

	valCopy, err := item.ValueCopy(nil)
	if err != nil {
		panic(err)
	}

	if len(strings.TrimSpace(string(valCopy))) == 0 {
		// Item is empty
		return ssh.Certificate{}, fmt.Errorf("empty")
	} else {

		// Parse the SSH certificate
		pubKey, err := ssh.ParsePublicKey(valCopy)
		if err != nil {
			log.Fatalf("Error parsing SSH certificate: %v", err)
		}

		cert, ok := pubKey.(*ssh.Certificate)
		if !ok {
			log.Fatalf("Key is not an SSH certificate")
		}

		return *cert, nil
	}

}

/* func getSshRevocationData(db *badger.DB, cert *x509.Certificate) X509RevokedCertificateInfo {
	var item *badger.Item
	var data X509RevokedCertificateInfo = X509RevokedCertificateInfo{}

	item, err := getItem(db, []byte("revoked_x509_certs"), []byte(cert.SerialNumber.String()))
	if err != nil {
		// we skip errors (like not found)
	} else {
		// we have found a revoked cert
		var valCopy []byte
		valCopy, err = item.ValueCopy(nil)
		if err != nil {
			panic(err)
		}

		if len(strings.TrimSpace(string(valCopy))) > 0 {
			if err := json.Unmarshal(valCopy, &data); err != nil {
				panic(err)
			}
		}
	}
	return data
}
*/
