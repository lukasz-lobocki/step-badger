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
	Long: `
Export x509 certificates' data out of the badger database of step-ca.`,

	Short:                 "Export x509 certificates.",
	DisableFlagsInUseLine: true,
	Use: `x509Certs <PATH> [flags]

Arguments:
  PATH   location of the source database`,

	Aliases: []string{"x509certs"},
	Example: `  step-badger x509certs ./db
  step-badger x509Certs ./db --revoked --valid=false --emit=openssl`,

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

	// Records selection criteria.
	x509certsCmd.Flags().BoolVarP(&config.showValid, "valid", "v", true, "valid certificates shown")
	x509certsCmd.Flags().BoolVarP(&config.showRevoked, "revoked", "r", false, "revoked certificates shown")
	x509certsCmd.Flags().BoolVarP(&config.showExpired, "expired", "e", false, "expired certificates shown")

	// Format choice
	x509certsCmd.Flags().Var(config.emitX509Format, "emit", "emit format: "+FORMAT_TABLE+"|"+FORMAT_JSON+"|"+FORMAT_MARKDOWN+
		"|"+FORMAT_OPENSSL+"|"+FORMAT_PLAIN)
	x509certsCmd.Flags().Var(config.timeFormat, "time", "time format: "+TIME_ISO+"|"+TIME_SHORT)
	x509certsCmd.Flags().Var(config.sortOrder, "sort", "sort order: "+SORT_START+"|"+SORT_FINISH)

	// Columns selection criteria.
	x509certsCmd.Flags().BoolVar(&config.showSerial, "serial", true, "serial number column shown")
	x509certsCmd.Flags().BoolVar(&config.showDNSNames, "dnsnames", false, "dns names column shown")
	x509certsCmd.Flags().BoolVar(&config.showEmailAddresses, "emailaddresses", false, "email addresses column shown")
	x509certsCmd.Flags().BoolVar(&config.showIPAddresses, "ipaddresses", false, "ip addresses column shown")
	x509certsCmd.Flags().BoolVar(&config.showURIs, "uris", false, "uris column shown")
	x509certsCmd.Flags().BoolVar(&config.showIssuer, "issuer", false, "issuer column shown")
	x509certsCmd.Flags().BoolVar(&config.showCrl, "crl", false, "crl column shown")
	x509certsCmd.Flags().BoolVar(&config.showProvisioner, "provisioner", false, "provisioner column shown")
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
	if err = db.Close(); err != nil {
		logError.Fatalln(err)
	}

	// Sort.
	switch thisSort := config.sortOrder.Value; thisSort {
	case SORT_FINISH:
		sort.SliceStable(x509CertificatesProvisionersRevocations, func(i, j int) bool {
			return x509CertificatesProvisionersRevocations[i].X509Certificate.NotAfter.
				Before(x509CertificatesProvisionersRevocations[j].X509Certificate.NotAfter)
		})
	case SORT_START:
		sort.SliceStable(x509CertificatesProvisionersRevocations, func(i, j int) bool {
			return x509CertificatesProvisionersRevocations[i].X509Certificate.NotBefore.
				Before(x509CertificatesProvisionersRevocations[j].X509Certificate.NotBefore)
		})
	}

	// Output.
	switch format := config.emitX509Format.Value; format {
	case FORMAT_JSON:
		emitX509CertsWithRevocationsJson(x509CertificatesProvisionersRevocations)
	case FORMAT_TABLE:
		emitX509Table(x509CertificatesProvisionersRevocations)
	case FORMAT_MARKDOWN:
		emitX509Markdown(x509CertificatesProvisionersRevocations)
	case FORMAT_OPENSSL:
		emitX509OpenSsl(x509CertificatesProvisionersRevocations)
	case FORMAT_PLAIN:
		emitX509Plain(x509CertificatesProvisionersRevocations)
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
