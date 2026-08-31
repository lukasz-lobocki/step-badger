package cmd

import (
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/smallstep/nosql/database"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

// sshCertsCmd represents the shell command.
var sshCertsCmd = &cobra.Command{
	Short:   "Export ssh certificates.",
	Run:     func(cmd *cobra.Command, args []string) { exportSshMain(args) },
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"sshcerts", "ssh"},

	DisableFlagsInUseLine: true,

	Example: "  step-badger sshCerts ./db",
	Long: `
Export ssh certificates' data out of the badger database of step-ca.`,
	Use: `sshCerts <PATH> [flags]

Arguments:
  PATH   location of the source database`,
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
	sshCertsCmd.Flags().BoolVarP(&config.showRevoked, "revoked", "r", false, "revoked certificates shown")
	sshCertsCmd.Flags().BoolVarP(&config.showExpired, "expired", "e", false, "expired certificates shown")

	// Format choice
	sshCertsCmd.Flags().Var(config.emitSshFormat, "emit", "emit format: "+config.emitSshFormat.Type())
	sshCertsCmd.Flags().Var(config.timeFormat, "time", "time format: "+config.timeFormat.Type())
	sshCertsCmd.Flags().Var(config.sortOrder, "sort", "sort order: "+config.sortOrder.Type())
	sshCertsCmd.Flags().Var(config.serialFormat, "serial", "serial format: "+config.serialFormat.Type())

	// Columns selection criteria.
	sshCertsCmd.Flags().BoolVar(&config.showHostType, "type", true, "host type column shown")
	sshCertsCmd.Flags().BoolVar(&config.showKeyId, "keyid", false, "key id column shown")
	sshCertsCmd.Flags().BoolVar(&config.showSignatureAlgorithm, "algorithm", false, "signature algorithm column shown")
}

/*
ExportSsh main function.

	'args' Given command line arguments, that contain the command to be run by shell.
*/
func exportSshMain(args []string) {

	checkLogginglevel(args)

	var (
		sshCertificateWithRevocation   tSshCertificateWithRevocation
		sshCertificatesWithRevocations = make([]tSshCertificateWithRevocation, 0)
		sshCertificateStringSerials    tCertificateStringSerials
	)

	// Open the database.
	db := openDB(args[0])

	// Get records from the ssh_certs bucket.
	records := listBucket(db, "ssh_certs")

	for _, record := range records {
		if loggingLevel >= 3 { // Show info.
			logInfo.Printf("Bucket: %s", record.Bucket)
			logInfo.Printf("Key: %s", record.Key)
			logInfo.Printf("Value: %q", record.Value)
		}

		// Get certificate.
		sshCertificate := parseValueToSshCertificate(record.Value)
		if loggingLevel >= 3 { // Show info.
			logInfo.Printf("Serial: %s", strconv.FormatUint(sshCertificate.Serial, 10))
			logInfo.Printf("Subject: %s", strings.Join(sshCertificate.ValidPrincipals, ","))
		}

		// Get revocation.
		sshCertificateRevocation := getSshRevocation(db, sshCertificate)
		if loggingLevel >= 3 { // Show info.
			logInfo.Printf("RevocationProvisionerID: %s", sshCertificateRevocation.ProvisionerID)
		}

		// Get serials and embed them as strings. This is to handle uint64 compatibility issues.
		sshCertificateStringSerials.SerialDec = strconv.FormatUint(sshCertificate.Serial, 10)
		sshCertificateStringSerials.SerialHex = strconv.FormatUint(sshCertificate.Serial, 16)

		// Populate the child.
		sshCertificateWithRevocation = tSshCertificateWithRevocation{
			SshCertificate:              sshCertificate,
			SshCertificateRevocation:    sshCertificateRevocation,
			SshCertificateStringSerials: sshCertificateStringSerials,
		}

		// Populate child validity info of the certificate.
		sshCertificateWithRevocation.Validity = classifyValidity(
			sshCertificateRevocation.ProvisionerID,
			sshCertificateRevocation.RevokedAt,
			time.Unix(int64(sshCertificate.ValidBefore), 0),
		)

		// Append child into collection, if record selection criteria are met.
		if selectValid(sshCertificateWithRevocation.Validity) {
			sshCertificatesWithRevocations = append(sshCertificatesWithRevocations, sshCertificateWithRevocation)
		}
	}

	// Close the database.
	closeDB(db, args[0])

	// Sort + emit via the shared driver; only the ordering predicates and columns are ssh-specific.
	spec := exportSpec[tSshCertificateWithRevocation]{
		columns:  getSshColumns(),
		rowLabel: func(x tSshCertificateWithRevocation) string { return x.SshCertificateStringSerials.SerialDec },
		finishLess: func(a, b tSshCertificateWithRevocation) bool {
			return a.SshCertificate.ValidBefore < b.SshCertificate.ValidBefore
		},
		startLess: func(a, b tSshCertificateWithRevocation) bool {
			return a.SshCertificate.ValidAfter < b.SshCertificate.ValidAfter
		},
	}
	sortRecords(sshCertificatesWithRevocations, spec)
	emitRecords(sshCertificatesWithRevocations, config.emitSshFormat, spec)
}

func getSshRevocation(thisDB database.DB, thisSshCertificate ssh.Certificate) tCertificateRevocation {

	serial := strconv.FormatUint(thisSshCertificate.Serial, 10)

	revocationValue, err := thisDB.Get([]byte("revoked_ssh_certs"), []byte(serial))

	switch {
	case errors.Is(err, database.ErrNotFound):
		if loggingLevel >= 3 { // Show info.
			logInfo.Printf("key for revocation not found")
		}
	case err != nil:
		logError.Panicf("lookup revocation for ssh serial %s: %v", serial, err)
	}

	if loggingLevel >= 3 { // Show info.
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
