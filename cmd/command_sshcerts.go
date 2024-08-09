package cmd

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/smallstep/nosql"
	"github.com/smallstep/nosql/database"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

// sshCertsCmd represents the shell command.
var sshCertsCmd = &cobra.Command{
	Long:  `Export ssh certificates' data out of the badger database of step-ca.`,
	Short: "Export ssh certificates.",
	Use:   "sshCerts <PATH>",

	Example: "  step-badger sshCerts ./db",

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

	// Records selection criteria.
	sshCertsCmd.Flags().BoolVarP(&config.showValid, "valid", "v", true, "valid certificates shown")
	sshCertsCmd.Flags().BoolVarP(&config.showRevoked, "revoked", "r", true, "revoked certificates shown")
	sshCertsCmd.Flags().BoolVarP(&config.showExpired, "expired", "e", false, "expired certificates shown")

	// Format choice
	sshCertsCmd.Flags().Var(config.emitSshFormat, "emit", "emit format: "+FORMAT_TABLE+"|"+FORMAT_JSON+"|"+FORMAT_MARKDOWN+"|"+FORMAT_PLAIN)
	sshCertsCmd.Flags().Var(config.timeFormat, "time", "time format: "+TIME_ISO+"|"+TIME_SHORT)
	sshCertsCmd.Flags().Var(config.sortOrder, "sort", "sort order: "+SORT_START+"|"+SORT_FINISH)

	// Columns selection criteria.
	sshCertsCmd.Flags().BoolVar(&config.showSerial, "serial", true, "serial column shown")
	sshCertsCmd.Flags().BoolVar(&config.showKeyId, "keyid", false, "key id column shown")
}

/*
ExportSsh main function.

	'args' Given command line arguments, that contain the command to be run by shell.
*/
func exportSshMain(args []string) {

	checkLogginglevel(args)

	var (
		err error
		db  database.DB

		sshCertificateWithRevocation   tSshCertificateWithRevocation
		sshCertificatesWithRevocations []tSshCertificateWithRevocation
	)

	// Open the database.
	db, err = nosql.New("badgerv2", args[0], database.WithValueDir(args[0]))
	if err != nil {
		logError.Fatalln(err)
	}

	// Get records from the ssh_certs bucket.
	records, err := db.List([]byte("ssh_certs"))
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
		sshCertificate := parseValueToSshCertificate(record.Value)
		if loggingLevel >= 2 { // Show info.
			logInfo.Printf("Serial: %s", strconv.FormatUint(sshCertificate.Serial, 10))
			logInfo.Printf("Subject: %s", strings.Join(sshCertificate.ValidPrincipals, ","))
		}

		// Get revocation.
		sshCertificateRevocation := getSshRevocation(db, sshCertificate)
		if loggingLevel >= 2 { // Show info.
			logInfo.Printf("RevocationProvisionerID: %s", sshCertificateRevocation.ProvisionerID)
		}

		// Populate the child.
		sshCertificateWithRevocation = tSshCertificateWithRevocation{
			SshCertificate:           sshCertificate,
			SshCertificateRevocation: sshCertificateRevocation,
		}

		// Populate child validity info of the certificate.
		if len(sshCertificateRevocation.ProvisionerID) > 0 && time.Now().After(sshCertificateRevocation.RevokedAt) {
			sshCertificateWithRevocation.Validity = REVOKED_STR
		} else {
			if time.Now().After(time.Unix(int64(sshCertificate.ValidBefore), 0)) {
				sshCertificateWithRevocation.Validity = EXPIRED_STR
			} else {
				sshCertificateWithRevocation.Validity = VALID_STR
			}
		}

		// Append child into collection, if record selection criteria are met.
		if (config.showExpired && sshCertificateWithRevocation.Validity == EXPIRED_STR) ||
			(config.showRevoked && sshCertificateWithRevocation.Validity == REVOKED_STR) ||
			(config.showValid && sshCertificateWithRevocation.Validity == VALID_STR) {
			sshCertificatesWithRevocations = append(sshCertificatesWithRevocations, sshCertificateWithRevocation)
		}
	}

	// Close the database.
	if err = db.Close(); err != nil {
		logError.Fatalln(err)
	}

	// Sort.
	switch thisSort := config.sortOrder.Value; thisSort {
	case SORT_FINISH:
		sort.SliceStable(sshCertificatesWithRevocations, func(i, j int) bool {
			return sshCertificatesWithRevocations[i].SshCertificate.ValidBefore < sshCertificatesWithRevocations[j].SshCertificate.ValidBefore
		})
	case SORT_START:
		sort.SliceStable(sshCertificatesWithRevocations, func(i, j int) bool {
			return sshCertificatesWithRevocations[i].SshCertificate.ValidAfter < sshCertificatesWithRevocations[j].SshCertificate.ValidAfter
		})
	}

	// Output.
	switch format := config.emitSshFormat.Value; format {
	case FORMAT_JSON:
		emitSshCertsJson(sshCertificatesWithRevocations)
	case FORMAT_TABLE:
		emitSshCertsTable(sshCertificatesWithRevocations)
	case FORMAT_MARKDOWN:
		emitSshCertsMarkdown(sshCertificatesWithRevocations)
	case FORMAT_PLAIN:
		emitSshCertsPlain(sshCertificatesWithRevocations)
	}
}

func getSshRevocation(thisDB database.DB, thisSshCertificate ssh.Certificate) tCertificateRevocation {

	revocationValue, err := thisDB.Get([]byte("revoked_ssh_certs"), []byte(strconv.FormatUint(thisSshCertificate.Serial, 10)))

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

func parseValueToSshCertificate(thisValue []byte) ssh.Certificate {

	var (
		sshCertificate *ssh.Certificate
	)

	// Parse the SSH certificate.
	pubKey, err := ssh.ParsePublicKey(thisValue)
	if err != nil {
		logError.Panicf("Error parsing SSH certificate: %v", err)
	}

	sshCertificate, ok := pubKey.(*ssh.Certificate)
	if !ok {
		logError.Panicf("Key is not an SSH certificate")
	}

	return *sshCertificate
}
