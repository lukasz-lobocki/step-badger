package cmd

import (
	"bytes"
	"crypto/x509"
	"encoding/binary"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	"time"

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

	retrieveCerts(db, []byte("x509_certs"), "/dev/stdout")
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

		var bigos tCertificateRevocation

		err := json.Unmarshal(valCopy, &bigos)
		if err != nil {
			fmt.Println("error:", err)
		}

		logInfo.Printf("DUPA=%s", bigos.RevokedAt)
		logInfo.Printf("key=%s ::\nvalue=%s", strings.TrimSpace(string(item.Key())), strings.TrimSpace(string(valCopy)))

	}
}

func retrieveCerts(db *badger.DB, prefix []byte, filename string) {
	txn := db.NewTransaction(false)
	defer txn.Discard()
	opts := badger.DefaultIteratorOptions
	prefix, err := badgerEncode(prefix)
	if err != nil {
		panic(err)
	}

	iter := txn.NewIterator(opts)
	defer iter.Close()

	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()

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
		revokedAt, flag := GetRevocationDate(db, cert)

		today := time.Now()
		if revokedAt == "" {
			// Cert is still valid?
			if today.Before(cert.NotAfter) {
				flag = "V"
			} else {
				flag = "E"
			}
		}

		//	0) Entry type. May be "V" (valid), "R" (revoked) or "E" (expired).
		//     Note that an expired may have the type "V" because the type has
		//     not been updated. 'openssl ca updatedb' does such an update.
		//  2) Revokation datetime. This is set for any entry of the type "R".
		//  3) Serial number.

		fmt.Fprintf(f, "\n\n******\n")

		fmt.Printf("Subject: %s\n", cert.Subject)
		fmt.Printf("Flag: %s\n", flag)
		fmt.Printf("revokedAt: %s\n", revokedAt)
		fmt.Printf("Not before: %s\n", cert.NotBefore)
		fmt.Printf("Not after: %s\n", cert.NotAfter)
		fmt.Printf("DNSNames: %s\n", cert.DNSNames)
		fmt.Printf("EmailAddresses: %s\n", cert.EmailAddresses)
		fmt.Printf("IPAddresses: %s\n", cert.IPAddresses)
		fmt.Printf("URIs: %s\n", cert.URIs)
		fmt.Printf("SerialNumber: %s\n", cert.SerialNumber)
		fmt.Printf("SerialNumberHex: %X\n", cert.SerialNumber)
		fmt.Printf("Issuer.CommonName: %s\n", cert.Issuer.CommonName)
		fmt.Printf("CRLDistributionPoints: %s\n", cert.CRLDistributionPoints)
		// fmt.Printf("IssuingCertificateURL: %s\n", cert.IssuingCertificateURL)
		//fmt.Printf("Extensions: %v\n", cert.Extensions)

		/* for _, ext := range cert.Extensions {
			fmt.Printf("Extension: %s value:%v\n", ext.Id.String(), []byte(ext.Value))
		} */

		jsonInfo, err := json.MarshalIndent(cert, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(jsonInfo))
	}
}

func GetRevocationDate(db *badger.DB, cert *x509.Certificate) (string, string) {
	flag := "X"
	var revokedAt string
	var revocationItem *badger.Item

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
			var revocationOptions RevokeOptions
			if err := json.Unmarshal(valCopy, &revocationOptions); err != nil {
				panic(err)
			}
			revokedAt = fmt.Sprintf(revocationOptions.RevokedAt.UTC().Format(time.RFC3339))
			flag = "R"
		}
	}
	return revokedAt, flag
}

type RevokeOptions struct {
	Serial      string
	Reason      string
	ReasonCode  int
	PassiveOnly bool
	RevokedAt   time.Time
	MTLS        bool
}

/*
getRevocationItem function returns revocation data (if exists) for the certificate of given serial number.

	'val' given byte slice, that contains the key data.
*/
func getRevocationItem(db *badger.DB, key []byte) (*badger.Item, error) {
	prefix := []byte("revoked_x509_certs")
	badgerKey, _ := toBadgerKey(prefix, key)
	txn := db.NewTransaction(false)
	defer txn.Discard()
	opts := badger.DefaultIteratorOptions
	_ = opts

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
