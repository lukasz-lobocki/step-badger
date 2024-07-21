package cmd

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"sort"
	"strings"

	"github.com/dgraph-io/badger/v2"
	"github.com/fatih/color"
	"github.com/lukasz-lobocki/tabby"
	"github.com/spf13/cobra"
)

// x509certsCmd represents the shell command
var x509certsCmd = &cobra.Command{
	Use:   "x509Certs BADGERPATH",
	Short: "Export x509 certificates.",
	Long:  `Export x509 certificates' data out of the badger database of step-ca.`,

	Example: "  step-badger x509certs ./db",

	Args: cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		exportX509Main(args)
	},
}

// Cobra initiation
func init() {
	rootCmd.AddCommand(x509certsCmd)

	x509certsCmd.Flags().VarP(config.emitFormat, "emit", "e", "emit format: table|json") // Choice
}

/*
Export main function

	'args' given command line arguments, that contain the command to be run by shell
*/
func exportX509Main(args []string) {

	checkLogginglevel(args)

	db, err := badger.Open(badger.DefaultOptions(args[0]).WithLogger(nil))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	x509CertsWithRevocations := retrieveX509Certs(db)

	sort.SliceStable(x509CertsWithRevocations, func(i, j int) bool {
		return x509CertsWithRevocations[i].X509Certificate.NotAfter.Before(x509CertsWithRevocations[j].X509Certificate.NotAfter)
	})

	switch thisFormat := config.emitFormat.Value; thisFormat {
	case "j":
		emitX509CertsWithRevocationsJson(x509CertsWithRevocations)
	case "t":
		emitX509Table(x509CertsWithRevocations)
	}

}

func retrieveX509Certs(db *badger.DB) []tX509CertificateAndRevocation {
	var (
		x509CertsWithRevocations []tX509CertificateAndRevocation = []tX509CertificateAndRevocation{}
	)

	prefix, err := badgerEncode([]byte("x509_certs"))
	if err != nil {
		logError.Panic(err)
	}

	txn := db.NewTransaction(false)
	defer txn.Discard()

	iter := txn.NewIterator(badger.DefaultIteratorOptions)
	defer iter.Close()

	for iter.Seek(prefix); iter.ValidForPrefix(prefix); iter.Next() {
		var (
			x509CertAndRevocation tX509CertificateAndRevocation = tX509CertificateAndRevocation{}
		)

		x509Cert, err := getX509Certificate(iter)
		if err != nil {
			continue
		}

		// Populate main info of the certificate.
		x509CertAndRevocation.X509Certificate = x509Cert

		// Populate revocation info of the certificate.
		x509CertAndRevocation.X509Revocation = getX509RevocationData(db, &x509Cert)

		// Populate provisioner sub-info of the certificate.
		x509CertAndRevocation.X509Provisioner = getX509CertificateProvisionerData(db, &x509Cert).Provisioner

		x509CertsWithRevocations = append(x509CertsWithRevocations, x509CertAndRevocation)
	}
	return x509CertsWithRevocations
}

func emitX509Table(x509CertsWithRevocations []tX509CertificateAndRevocation) {
	table := new(tabby.Table)

	thisColumns := getX509Columns()

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

	for _, x509CertAndRevocation := range x509CertsWithRevocations {

		var thisRow []string
		/* Building slice of columns within a single row*/
		for _, thisColumn := range thisColumns {

			if thisColumn.isShown() {
				thisRow = append(thisRow,
					color.New(thisColumn.contentColor(x509CertAndRevocation)).SprintFunc()(
						thisColumn.contentSource(x509CertAndRevocation),
					),
				)
			}
		}

		if err := table.AppendRow(thisRow); err != nil {
			logError.Panic(err)
		}
		if loggingLevel >= 3 {
			logInfo.Printf("row [%s] appended.", x509CertAndRevocation.X509Certificate.SerialNumber.String())
		}

	}

	if loggingLevel >= 2 {
		logInfo.Printf("%d rows appended.\n", len(x509CertsWithRevocations))
	}

	/* Emit the table */

	if loggingLevel >= 3 {
		table.Print(&tabby.Config{Spacing: "|", Padding: "."})
	} else {
		table.Print(nil)
	}
}

func getX509Certificate(iter *badger.Iterator) (x509.Certificate, error) {
	item := iter.Item()

	var (
		valCopy  []byte
		x509cert *x509.Certificate
	)

	valCopy, err := item.ValueCopy(nil)
	if err != nil {
		logError.Fatalf("Error parsing item value: %v", err)
	}

	if len(strings.TrimSpace(string(valCopy))) == 0 {
		// Item is empty
		return x509.Certificate{}, fmt.Errorf("empty")
	} else {
		// read data to object
		marshaledValue, err := json.Marshal(valCopy)
		if err != nil {
			logError.Panic(err)
		}

		// make x509Cert-data from db decodable pem
		// json contains ""
		base64cert := fmt.Sprintf("-----BEGIN CERTIFICATE-----\n%s\n-----END CERTIFICATE-----",
			strings.ReplaceAll(string(marshaledValue), "\"", ""))
		decodedPEMBlock, _ := pem.Decode([]byte(base64cert))

		if decodedPEMBlock == nil {
			logError.Fatalf("failed to parse certificate PEM")
		}

		x509cert, err = x509.ParseCertificate(decodedPEMBlock.Bytes)
		if err != nil {
			logError.Fatalf("failed to parse certificate: " + err.Error())
		}

		return *x509cert, nil
	}

}

func getX509RevocationData(db *badger.DB, cert *x509.Certificate) tX509RevokedCertificate {
	var item *badger.Item
	var data tX509RevokedCertificate = tX509RevokedCertificate{}

	item, err := getItem(db, []byte("revoked_x509_certs"), []byte(cert.SerialNumber.String()))
	if err != nil {
		// we skip errors (like not found)
	} else {
		// we have found a revoked cert
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

func getX509CertificateProvisionerData(db *badger.DB, cert *x509.Certificate) tX509Certificate {
	var item *badger.Item
	var info tX509Certificate = tX509Certificate{}

	item, err := getItem(db, []byte("x509_certs_data"), []byte(cert.SerialNumber.String()))
	if err != nil {
		// we skip errors (like not found)
	} else {
		// we have found a revoked cert
		var valCopy []byte
		valCopy, err = item.ValueCopy(nil)
		if err != nil {
			logError.Panic(err)
		}

		if len(strings.TrimSpace(string(valCopy))) > 0 {
			if err := json.Unmarshal(valCopy, &info); err != nil {
				logError.Panic(err)
			}
		}
	}
	return info
}

func emitX509CertsWithRevocationsJson(x509CertsWithRevocations []tX509CertificateAndRevocation) {
	jsonInfo, err := json.MarshalIndent(x509CertsWithRevocations, "", "  ")
	if err != nil {
		logError.Panic(err)
	}
	fmt.Println(string(jsonInfo))
	if loggingLevel >= 2 {
		logInfo.Printf("%d records marshalled.\n", len(x509CertsWithRevocations))
	}
}
