package cmd

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strings"

	"github.com/dgraph-io/badger/v2"
	"github.com/fatih/color"
	"github.com/lukasz-lobocki/tabby"
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

	x509certsCmd.Flags().SortFlags = false
	x509certsCmd.Flags().VarP(config.sortOrder, "order", "o", "order: validity|before")           // Choice
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
	var (
		allCertsWithRevocations []X509CertificateAndRevocationInfo = []X509CertificateAndRevocationInfo{}
	)

	prefix, err := badgerEncode([]byte("x509_certs"))
	if err != nil {
		panic(err)
	}

	txn := db.NewTransaction(false)
	defer txn.Discard()

	iter := txn.NewIterator(badger.DefaultIteratorOptions)
	defer iter.Close()

	for iter.Seek(prefix); iter.ValidForPrefix(prefix); iter.Next() {
		var (
			oneCertAndRevocation X509CertificateAndRevocationInfo = X509CertificateAndRevocationInfo{}
		)

		oneCert, err := getCertificate(iter)
		if err != nil {
			continue
		}

		// Populate main info of the certificate.
		oneCertAndRevocation.Certificate = oneCert

		// Populate revocation info of the certificate.
		oneCertAndRevocation.Revocation = getRevocationData(db, &oneCert)

		// Populate provisioner sub-info of the certificate.
		oneCertAndRevocation.Provisioner = getX509CertificateProvisionerData(db, &oneCert).Provisioner

		allCertsWithRevocations = append(allCertsWithRevocations, oneCertAndRevocation)
	}

	table := new(tabby.Table)

	thisColumns := getColumns()

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
		panic(err) //"emitTable: setting header failed. %w", err)
	}

	if loggingLevel >= 1 {
		logInfo.Println("header set.")
	}

	/* Populate the table */

	for _, oneCertAndRevocation := range allCertsWithRevocations {

		var thisRow []string

		/* Building slice of columns within a single row*/

		for _, thisColumn := range thisColumns {

			if thisColumn.isShown() {
				thisRow = append(thisRow,
					color.New(thisColumn.contentColor(oneCertAndRevocation)).SprintFunc()(
						thisColumn.contentSource(oneCertAndRevocation),
					),
				)
			}
		}

		if err := table.AppendRow(thisRow); err != nil {
			panic(err) //return fmt.Errorf("emitTable: appending row failed. %w", err)
		}
		// if loggingLevel >= 3 {
		// 	logInfo.Printf("row [%s] appended.", thisCertWithRevocation.ShortName)
		// }

	}

	if loggingLevel >= 2 {
		logInfo.Printf("%d rows appended.\n", len(allCertsWithRevocations))
	}

	/* Emit the table */

	if loggingLevel >= 3 {
		table.Print(&tabby.Config{Spacing: "|", Padding: "."})
	} else {
		table.Print(nil)
	}

	// =========================

	// jsonInfo, err := json.MarshalIndent(allCertsWithRevocations, "", "  ")
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(jsonInfo))
}

func getCertificate(iter *badger.Iterator) (x509.Certificate, error) {
	item := iter.Item()

	var (
		valCopy []byte
		cert    *x509.Certificate
	)

	valCopy, err := item.ValueCopy(nil)
	if err != nil {
		panic(err)
	}

	if len(strings.TrimSpace(string(valCopy))) == 0 {
		// Item is empty
		return x509.Certificate{}, fmt.Errorf("empty")
	} else {
		// read data to object
		marshaledValue, err := json.Marshal(valCopy)
		if err != nil {
			panic(err)
		}

		// make oneCert-data from db decodable pem
		// json contains ""
		base64cert := fmt.Sprintf("-----BEGIN CERTIFICATE-----\n%s\n-----END CERTIFICATE-----",
			strings.ReplaceAll(string(marshaledValue), "\"", ""))
		decodedPEMBlock, _ := pem.Decode([]byte(base64cert))

		if decodedPEMBlock == nil {
			panic("failed to parse certificate PEM")
		}

		cert, err = x509.ParseCertificate(decodedPEMBlock.Bytes)
		if err != nil {
			panic("failed to parse certificate: " + err.Error())
		}

		return *cert, nil
	}

}

func getRevocationData(db *badger.DB, cert *x509.Certificate) RevokedX509CertificateInfo {
	var item *badger.Item
	var data RevokedX509CertificateInfo = RevokedX509CertificateInfo{}

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

func getX509CertificateProvisionerData(db *badger.DB, cert *x509.Certificate) X509CertificateInfo {
	var item *badger.Item
	var info X509CertificateInfo = X509CertificateInfo{}

	item, err := getItem(db, []byte("x509_certs_data"), []byte(cert.SerialNumber.String()))
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
			if err := json.Unmarshal(valCopy, &info); err != nil {
				panic(err)
			}
		}
	}
	return info
}
