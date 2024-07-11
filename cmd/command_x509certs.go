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

// x509certsCmd represents the shell command
var x509certsCmd = &cobra.Command{
	Use:   "x509certs [PATH] \"command\"",
	Short: "Export certificates.",
	Long:  `Export certificates' data out of the badger database of step-ca.`,

	Example: "  gitas x509certsCmd XXXXX -duda",

	Args: cobra.RangeArgs(1, 2),

	Run: func(cmd *cobra.Command, args []string) {
		exportMain(args)
	},
}

// Cobra initiation
func init() {
	rootCmd.AddCommand(x509certsCmd)

	initChoices()

	x509certsCmd.Flags().VarP(config.emitFormat, "emit", "e", "emit format: table|json|markdown") // Choice
}

/*
Export main function

	'args' given command line arguments, that contain the command to be run by shell
*/
func exportMain(args []string) {

	checkLogginglevel(args)

	logInfo.Println(args[0])

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
}

func retrieveCerts(db *badger.DB) {
	var allCertsAndRevos []X509CertificateAndRevocationInfo

	prefix, err := badgerEncode([]byte("x509_certs"))
	if err != nil {
		panic(err)
	}

	txn := db.NewTransaction(false)
	defer txn.Discard()

	iter := txn.NewIterator(badger.DefaultIteratorOptions)
	defer iter.Close()

	for iter.Seek(prefix); iter.ValidForPrefix(prefix); iter.Next() {
		var oneCertAndRevo X509CertificateAndRevocationInfo = X509CertificateAndRevocationInfo{}
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

func getRevocationData(db *badger.DB, cert *x509.Certificate) RevokedX509CertificateInfo {
	var revocationItem *badger.Item
	var revocationOptions RevokedX509CertificateInfo = RevokedX509CertificateInfo{}

	revocationItem, err := getItem(db, []byte("revoked_x509_certs"), []byte(cert.SerialNumber.String()))
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
getItem function returns revocation data (if exists) for the certificate of given serial number.

	'db' badger database.
	'key' certificate serial number.
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
