package cmd

import (
	"bytes"
	"crypto/x509"
	"encoding/binary"
	"encoding/json"
	"encoding/pem"
	"fmt"
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

var config ConfigInfo // Holds status' configuration

// Cobra initiation
func init() {
	rootCmd.AddCommand(exportCmd)

	initChoices()

	exportCmd.Flags().VarP(config.emitFormat, "emit", "e", "emit format: table|json|markdown") // Choice
}

/*
Export main function

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
		panic(err)
	}
	defer db.Close()

	retrieveCerts(db)
	// retrieveTableData(db, []byte("x509_certs_data"), "/dev/stdout")
	// retrieveTableData(db, []byte("admins"), "/dev/stdout")
	// retrieveTableData(db, []byte("provisioners"), "/dev/stdout")
	// retrieveTableData(db, []byte("authority_policies"), "/dev/stdout")
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

func retrieveCerts(db *badger.DB) {
	var allCertsAndRevos []CertificateAndRevocationInfo

	prefix, err := badgerEncode([]byte("x509_certs"))
	if err != nil {
		panic(err)
	}

	txn := db.NewTransaction(false)
	defer txn.Discard()

	iter := txn.NewIterator(badger.DefaultIteratorOptions)
	defer iter.Close()

	for iter.Seek(prefix); iter.ValidForPrefix(prefix); iter.Next() {
		var oneCertAndRevo CertificateAndRevocationInfo = CertificateAndRevocationInfo{}
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

		// read data to object
		marshaledValue, err := json.Marshal(valCopy)
		if err != nil {
			panic(err)
		}

		// make cert-data from db decodable pem
		// json contains ""
		strippedMarshaledValue := strings.ReplaceAll(string(marshaledValue), "\"", "")
		base64 := fmt.Sprintf("-----BEGIN CERTIFICATE-----\n%s\n-----END CERTIFICATE-----", strippedMarshaledValue)
		decodedPEMBlock, _ := pem.Decode([]byte(base64))
		//fmt.Printf("type: %s\n", decodedPEMBlock.Type)
		if decodedPEMBlock == nil {
			panic("failed to parse certificate PEM")
		}
		cert, err := x509.ParseCertificate(decodedPEMBlock.Bytes)

		if err != nil {
			panic("failed to parse certificate: " + err.Error())
		}

		// see if cert is revoked
		revocationData := getRevocationData(db, cert)

		/* 		fmt.Printf("Subject: %s\n", cert.Subject)
		   		fmt.Printf("CertSN: %s\n", cert.SerialNumber)
		   		fmt.Printf("RevoSN: %s\n", revocationData.Serial)
		*/
		oneCertAndRevo.Certificate = *cert
		oneCertAndRevo.Revocation = revocationData

		allCertsAndRevos = append(allCertsAndRevos, oneCertAndRevo)
	}
	jsonInfo, err := json.MarshalIndent(allCertsAndRevos, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonInfo))

}

func getRevocationData(db *badger.DB, cert *x509.Certificate) RevokedCertificateInfo {
	var revocationItem *badger.Item
	var revocationOptions RevokedCertificateInfo = RevokedCertificateInfo{}

	revocationItem, err := getRevocationItem(db, []byte(cert.SerialNumber.String()))
	if err != nil {
		// we skip errors on revoke (like not found)
	} else {
		// we have found a revoked cert
		var valCopy []byte
		valCopy, err = revocationItem.ValueCopy(nil)
		if err != nil {
			panic(err)
		}

		if len(strings.TrimSpace(string(valCopy))) > 0 {
			if err := json.Unmarshal(valCopy, &revocationOptions); err != nil {
				panic(err)
			}
		}
	}
	return revocationOptions
}

/*
getRevocationItem function returns revocation data (if exists) for the certificate of given serial number.

	'db' badger database.
	'key' certificate serial number.
*/
func getRevocationItem(db *badger.DB, key []byte) (*badger.Item, error) {
	prefix := []byte("revoked_x509_certs")
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
