package cmd

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/spf13/cobra"
)

// x509certsCmd represents the shell command.
var x509certsCmd = &cobra.Command{
	Use:   "x509Certs PATH",
	Short: "Export x509 certificates.",
	Long:  `Export x509 certificates' data out of the badger database of step-ca.`,

	Example: "  step-badger x509certs ./db",

	Args: cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		exportX509Main(args)
	},
}

/*
Cobra initiation.
*/
func init() {
	rootCmd.AddCommand(x509certsCmd)

	// Hide help command.
	x509certsCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	//Do not sort flags.
	x509certsCmd.Flags().SortFlags = false

	x509certsCmd.Flags().VarP(config.emitX509Format, "emit", "e", "emit format: table|json|markdown|openssl") // Choice
	x509certsCmd.Flags().VarP(config.timeFormat, "time", "t", "time format: iso|short")                       // Choice
	x509certsCmd.Flags().VarP(config.sortOrder, "sort", "s", "sort order: start|finish")                      // Choice

	// Columns selection criteria.
	x509certsCmd.Flags().BoolVarP(&config.showDNSNames, "dnsnames", "d", false, "DNSNames column shown")
	x509certsCmd.Flags().BoolVarP(&config.showEmailAddresses, "emailaddresses", "m", false, "EmailAddresses column shown")
	x509certsCmd.Flags().BoolVarP(&config.showIPAddresses, "ipaddresses", "i", false, "IPAddresses column shown")
	x509certsCmd.Flags().BoolVarP(&config.showURIs, "uris", "u", false, "URIs column shown")
	x509certsCmd.Flags().BoolVarP(&config.showCrl, "crl", "c", false, "crl column shown")
	x509certsCmd.Flags().BoolVarP(&config.showProvisioner, "provisioner", "p", false, "provisioner column shown")

	// Records selection criteria.
	x509certsCmd.Flags().BoolVarP(&config.showValid, "valid", "v", true, "valid certificates shown")
	x509certsCmd.Flags().BoolVarP(&config.showRevoked, "revoked", "r", true, "revoked certificates shown")
	x509certsCmd.Flags().BoolVarP(&config.showExpired, "expired", "x", false, "expired certificates shown")
}

/*
Export x509 main function.

	'args' Given command line arguments, that contain the command to be run by shell.
*/
func exportX509Main(args []string) {

	checkLogginglevel(args)

	db, err := badger.Open(badger.DefaultOptions(args[0]).WithLogger(nil))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Get.
	x509CertsWithRevocations := getX509Certs(db)

	// Sort.
	switch thisSort := config.sortOrder.Value; thisSort {
	case "f":
		sort.SliceStable(x509CertsWithRevocations, func(i, j int) bool {
			return x509CertsWithRevocations[i].X509Certificate.NotAfter.Before(x509CertsWithRevocations[j].X509Certificate.NotAfter)
		})
	case "s":
		sort.SliceStable(x509CertsWithRevocations, func(i, j int) bool {
			return x509CertsWithRevocations[i].X509Certificate.NotBefore.Before(x509CertsWithRevocations[j].X509Certificate.NotBefore)
		})
	}

	// Output.
	switch thisFormat := config.emitX509Format.Value; thisFormat {
	case "j":
		emitX509CertsWithRevocationsJson(x509CertsWithRevocations)
	case "t":
		emitX509Table(x509CertsWithRevocations)
	case "m":
		emitX509Markdown(x509CertsWithRevocations)
	case "o":
		emitOpenSsl(x509CertsWithRevocations)
	}

}

/*
getX509Certs returns struct with x509 certificates.

	'thisDb' Badger database.
*/
func getX509Certs(thisDb *badger.DB) []tX509CertificateWithRevocation {
	var (
		x509CertsWithRevocations []tX509CertificateWithRevocation = []tX509CertificateWithRevocation{}
	)

	thisPrefix, err := badgerEncode([]byte("x509_certs"))
	if err != nil {
		logError.Panic(err)
	}

	txn := thisDb.NewTransaction(false)
	defer txn.Discard()

	iter := txn.NewIterator(badger.DefaultIteratorOptions)
	defer iter.Close()

	for iter.Seek(thisPrefix); iter.ValidForPrefix(thisPrefix); iter.Next() {
		var (
			x509CertWithRevocation tX509CertificateWithRevocation = tX509CertificateWithRevocation{}
		)

		x509Cert, err := getX509Certificate(iter)
		if err != nil {
			continue
		}

		// Populate child main info of the certificate.
		x509CertWithRevocation.X509Certificate = x509Cert

		// Populate child revocation info of the certificate.
		x509CertWithRevocation.X509Revocation = getX509RevocationData(thisDb, &x509Cert)

		// Populate child provisioner sub-info of the certificate.
		x509CertWithRevocation.X509Provisioner = getX509CertificateProvisionerData(thisDb, &x509Cert).Provisioner

		// Populate child validity info of the certificate.
		if len(x509CertWithRevocation.X509Revocation.ProvisionerID) > 0 && time.Now().After(x509CertWithRevocation.X509Revocation.RevokedAt) {
			x509CertWithRevocation.Validity = REVOKED_STR
		} else {
			if time.Now().After(x509CertWithRevocation.X509Certificate.NotAfter) {
				x509CertWithRevocation.Validity = EXPIRED_STR
			} else {
				x509CertWithRevocation.Validity = VALID_STR
			}
		}

		// Append child to collection, if record selection criteria are met.
		if (config.showExpired && x509CertWithRevocation.Validity == EXPIRED_STR) ||
			(config.showRevoked && x509CertWithRevocation.Validity == REVOKED_STR) ||
			(config.showValid && x509CertWithRevocation.Validity == VALID_STR) {
			x509CertsWithRevocations = append(x509CertsWithRevocations, x509CertWithRevocation)
		}
	}
	return x509CertsWithRevocations
}

/*
getX509Certificate returns x509 certificate.
*/
func getX509Certificate(thisIter *badger.Iterator) (x509.Certificate, error) {
	item := thisIter.Item()

	var (
		valCopy  []byte
		x509cert *x509.Certificate
	)

	valCopy, err := item.ValueCopy(nil)
	if err != nil {
		logError.Panicf("Error parsing item value: %v", err)
	}

	if len(strings.TrimSpace(string(valCopy))) == 0 {
		// Item is empty.
		return x509.Certificate{}, fmt.Errorf("empty")
	} else {
		// Read data to object.
		marshaledValue, err := json.Marshal(valCopy)
		if err != nil {
			logError.Panic(err)
		}

		// Make x509Cert-data from db decodable pem.
		base64cert := fmt.Sprintf("-----BEGIN CERTIFICATE-----\n%s\n-----END CERTIFICATE-----",
			strings.ReplaceAll(string(marshaledValue), "\"", ""))
		decodedPEMBlock, _ := pem.Decode([]byte(base64cert))

		if decodedPEMBlock == nil {
			logError.Panicf("failed to parse certificate PEM")
		}

		x509cert, err = x509.ParseCertificate(decodedPEMBlock.Bytes)
		if err != nil {
			logError.Panicf("failed to parse certificate: " + err.Error())
		}

		return *x509cert, nil
	}

}

/*
getX509RevocationData returns revocation information for a given certificate, if exists.

	'thisDb' Badger database
	'thisCert' Certificate to get revocation information for.
*/
func getX509RevocationData(thisDb *badger.DB, thisCert *x509.Certificate) tRevokedCertificate {
	var item *badger.Item
	var data tRevokedCertificate = tRevokedCertificate{}

	item, err := getItem(thisDb, []byte("revoked_x509_certs"), []byte(thisCert.SerialNumber.String()))
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

/*
getX509CertificateProvisionerData returns provisioner information for a given certificate, if exists.

	'thisDb' Badger database.
	'thisCert' Certificate to get provisioner information for.
*/
func getX509CertificateProvisionerData(thisDb *badger.DB, thisCert *x509.Certificate) tX509Certificate {
	var item *badger.Item
	var info tX509Certificate = tX509Certificate{}

	item, err := getItem(thisDb, []byte("x509_certs_data"), []byte(thisCert.SerialNumber.String()))
	if err != nil {
		// Skip errors (like not found).
	} else {
		// Found a provisioner info.
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
