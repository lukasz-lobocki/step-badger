package cmd

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/smallstep/nosql"
	"github.com/smallstep/nosql/database"
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

	var (
		err error
		db  database.DB

		x509CertificateProvisionerRevocation    tX509CertificateProvisionerRevocation
		x509CertificatesProvisionersRevocations []tX509CertificateProvisionerRevocation
	)

	// Open the database.
	db, err = nosql.New("badgerv2", args[0], database.WithValueDir(args[0]))
	if err != nil {
		logError.Fatalln(err)
	}

	// Get records from the x509_certs bucket.
	records, err := db.List([]byte("x509_certs"))
	if err != nil {
		logError.Fatalln(err)
	}
	if records == nil {
		logError.Fatalln("no records found")
	}

	for _, record := range records {
		if loggingLevel >= 2 { // Show info.
			logInfo.Printf("Bucket: %s", record.Bucket)
			logInfo.Printf("Key: %s", record.Key)
			logInfo.Printf("Value: %q", record.Value)
		}

		// Get certificate.
		x509Certificate := parseValueToX509Certificate(record.Value)
		if loggingLevel >= 2 { // Show info.
			logInfo.Printf("Serial: %s", x509Certificate.SerialNumber.String())
			logInfo.Printf("Subject: %s", x509Certificate.Subject)
		}

		// Get revocation.
		x509CertificateRevocation := getX509Revocation(db, x509Certificate)
		if loggingLevel >= 2 { // Show info.
			logInfo.Printf("RevocationProvisionerID: %s", x509CertificateRevocation.ProvisionerID)
		}

		// Get provisioner.
		x509CertificateData := getX509CertificateData(db, x509Certificate)
		if loggingLevel >= 2 { // Show info.
			logInfo.Printf("Provisioner: %s", x509CertificateData.Provisioner.Type)
		}

		// Populate the child.
		x509CertificateProvisionerRevocation = tX509CertificateProvisionerRevocation{
			X509Certificate: x509Certificate,
			X509Revocation:  x509CertificateRevocation,
			X509Provisioner: x509CertificateData.Provisioner,
		}

		// Populate child validity info of the certificate.
		if len(x509CertificateRevocation.ProvisionerID) > 0 && time.Now().After(x509CertificateRevocation.RevokedAt) {
			x509CertificateProvisionerRevocation.Validity = REVOKED_STR
		} else {
			if time.Now().After(x509Certificate.NotAfter) {
				x509CertificateProvisionerRevocation.Validity = EXPIRED_STR
			} else {
				x509CertificateProvisionerRevocation.Validity = VALID_STR
			}
		}

		// Append child into collection, if record selection criteria are met.
		if (config.showExpired && x509CertificateProvisionerRevocation.Validity == EXPIRED_STR) ||
			(config.showRevoked && x509CertificateProvisionerRevocation.Validity == REVOKED_STR) ||
			(config.showValid && x509CertificateProvisionerRevocation.Validity == VALID_STR) {
			x509CertificatesProvisionersRevocations = append(x509CertificatesProvisionersRevocations,
				x509CertificateProvisionerRevocation)
		}

	}

	// Close the database.
	err = db.Close()
	if err != nil {
		logError.Fatalln(err)
	}

	// Sort.
	switch thisSort := config.sortOrder.Value; thisSort {
	case "f":
		sort.SliceStable(x509CertificatesProvisionersRevocations, func(i, j int) bool {
			return x509CertificatesProvisionersRevocations[i].X509Certificate.NotAfter.
				Before(x509CertificatesProvisionersRevocations[j].X509Certificate.NotAfter)
		})
	case "s":
		sort.SliceStable(x509CertificatesProvisionersRevocations, func(i, j int) bool {
			return x509CertificatesProvisionersRevocations[i].X509Certificate.NotBefore.
				Before(x509CertificatesProvisionersRevocations[j].X509Certificate.NotBefore)
		})
	}

	// Output.
	switch format := config.emitX509Format.Value; format {
	case "j":
		emitX509CertsWithRevocationsJson(x509CertificatesProvisionersRevocations)
	case "t":
		emitX509Table(x509CertificatesProvisionersRevocations)
	case "m":
		emitX509Markdown(x509CertificatesProvisionersRevocations)
	case "o":
		emitOpenSsl(x509CertificatesProvisionersRevocations)
	}
}

func getX509Revocation(thisDB database.DB, thisX509Certificate x509.Certificate) tCertificateRevocation {

	revocationValue, err := thisDB.Get([]byte("revoked_x509_certs"), []byte(thisX509Certificate.SerialNumber.String()))

	switch {
	case errors.Is(err, database.ErrNotFound):
		if loggingLevel >= 2 { // Show info.
			logInfo.Printf("key for revocation not found")
		}
	case err != nil:
		logInfo.Panic(err)
	}

	if loggingLevel >= 2 { // Show info.
		logInfo.Printf("revocationValue: %s", revocationValue)
	}

	return parseValueToCertificateRevocation(revocationValue)
}

func getX509CertificateData(thisDB database.DB, thisX509Certificate x509.Certificate) tX509CertificateData {

	certsDataValue, err := thisDB.Get([]byte("x509_certs_data"), []byte(thisX509Certificate.SerialNumber.String()))

	switch {
	case errors.Is(err, database.ErrNotFound):
		if loggingLevel >= 2 { // Show info.
			logInfo.Printf("key for certificate data not found")
		}
	case err != nil:
		logInfo.Panic(err)
	}

	if loggingLevel >= 2 { // Show info.
		logInfo.Printf("certsDataValue: %s", certsDataValue)
	}

	return parseValueToX509CertificateData(certsDataValue)
}

func parseValueToX509CertificateData(thisValue []byte) tX509CertificateData {

	var (
		certificateData tX509CertificateData
	)

	if len(strings.TrimSpace(string(thisValue))) > 0 {
		if err := json.Unmarshal(thisValue, &certificateData); err != nil {
			logError.Panic(err)
		}
	}
	return certificateData
}

func parseValueToCertificateRevocation(thisValue []byte) tCertificateRevocation {

	var (
		certificateRevocation tCertificateRevocation
	)

	if len(strings.TrimSpace(string(thisValue))) > 0 {
		if err := json.Unmarshal(thisValue, &certificateRevocation); err != nil {
			logError.Panic(err)
		}
	}
	return certificateRevocation
}

func parseValueToX509Certificate(thisValue []byte) x509.Certificate {

	var (
		x509Certificate *x509.Certificate
	)

	marshaledValue, err := json.Marshal(thisValue)
	if err != nil {
		logError.Panic(err)
	}

	if loggingLevel >= 2 { // Show info.
		logInfo.Printf("marshaledValue: %s", marshaledValue)
	}

	// Adding header and footer.
	pemBlockValue := makePEM(marshaledValue)

	if loggingLevel >= 2 { // Show info.
		logInfo.Printf("pemBlockValue: %s", pemBlockValue)
	}

	// Decode the PEM block.
	decodedPEMBlock, _ := pem.Decode([]byte(pemBlockValue))
	if decodedPEMBlock == nil {
		logError.Panicf("failed to parse certificate PEM")
	}

	// Parse PEM block into certificate.
	x509Certificate, err = x509.ParseCertificate(decodedPEMBlock.Bytes)
	if err != nil {
		logError.Panicf("failed to parse certificate: " + err.Error())
	}

	return *x509Certificate
}

func makePEM(thisValue []byte) string {
	return fmt.Sprintf("-----BEGIN CERTIFICATE-----\n%s\n-----END CERTIFICATE-----",
		strings.ReplaceAll(string(thisValue), "\"", ""))
}
