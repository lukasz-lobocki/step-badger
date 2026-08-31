package cmd

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/smallstep/nosql"
	"github.com/smallstep/nosql/database"
)

const (
	MAX_LOGGING_LEVEL int    = 3 // Maximum allowed logging level.
	TIME_SHORT        string = "short"
	TIME_ISO          string = "iso"
	SORT_START        string = "start"
	SORT_FINISH       string = "finish"
	FORMAT_TABLE      string = "table"
	FORMAT_JSON       string = "json"
	FORMAT_MARKDOWN   string = "markdown"
	FORMAT_OPENSSL    string = "openssl"
	FORMAT_PLAIN      string = "plain"
	SERIAL_DEC        string = "dec"
	SERIAL_HEX        string = "hex"
)

/*
classifyValidity returns the validity status of a certificate, shared by the ssh and
x509 export paths. A certificate is Revoked when it carries a revocation record whose
RevokedAt is in the past; otherwise Expired when its expiry time is in the past; else
Valid. It mirrors the original two-step check (revocation first, then expiry) exactly.

	'revokedProvisionerID' non-empty when a revocation record exists for the cert.
	'revokedAt'          the revocation timestamp (only meaningful if revoked).
	'expiry'             the certificate's end-of-validity time.
*/
func classifyValidity(revokedProvisionerID string, revokedAt, expiry time.Time) string {
	if len(revokedProvisionerID) > 0 && time.Now().After(revokedAt) {
		return REVOKED_STR
	}
	if time.Now().After(expiry) {
		return EXPIRED_STR
	}
	return VALID_STR
}

/*
openDB opens the badger database at path, logging its open at level >= 1 and exiting on
error. Shared by both export paths so DB-lifecycle error handling lives in one place.
*/
func openDB(path string) database.DB {
	db, err := nosql.New("badgerv2", path, database.WithValueDir(path))
	if err != nil {
		logError.Fatalln(err)
	}
	if loggingLevel >= 1 { // Show info.
		logInfo.Printf("Database opened: %s", path)
	}
	return db
}

/*
listBucket returns all entries in bucket, exiting on error or when the bucket is empty.
The "no records found" message matches the original handlers exactly.
*/
func listBucket(db database.DB, bucket string) []*database.Entry {
	records, err := db.List([]byte(bucket))
	if err != nil {
		logError.Fatalln(err)
	}
	if records == nil {
		logError.Fatalln("no records found")
	}
	return records
}

/*
closeDB closes the database, logging its close at level >= 1 and exiting on error.
*/
func closeDB(db database.DB, path string) {
	if err := db.Close(); err != nil {
		logError.Fatalln(err)
	}
	if loggingLevel >= 1 { // Show info.
		logInfo.Printf("Database closed: %s", path)
	}
}

/*
initLoggers creates colorful loggers.
*/
func initLoggers() {
	hiCyan := color.New(color.FgHiCyan).SprintFunc()
	hiYellow := color.New(color.FgHiYellow).SprintFunc()
	hiRed := color.New(color.FgHiRed).SprintFunc()

	logInfo = log.New(os.Stderr, hiCyan("╭info\n╰"), 0)
	logWarning = log.New(os.Stderr, hiYellow("╭warning\n╰"), log.Lshortfile)
	logError = log.New(os.Stderr, hiRed("╭error\n╰"), log.Lshortfile)
}

/*
initChoices sets up Config struct for 'limited choice' flag.
*/
func initChoices() {
	config.emitSshFormat = newChoice([]string{FORMAT_TABLE, FORMAT_JSON, FORMAT_MARKDOWN, FORMAT_PLAIN}, FORMAT_TABLE)
	config.emitX509Format = newChoice([]string{FORMAT_TABLE, FORMAT_JSON, FORMAT_MARKDOWN, FORMAT_OPENSSL, FORMAT_PLAIN}, FORMAT_TABLE)
	config.sortOrder = newChoice([]string{SORT_START, SORT_FINISH}, SORT_FINISH)
	config.timeFormat = newChoice([]string{TIME_ISO, TIME_SHORT}, TIME_ISO)
	config.serialFormat = newChoice([]string{SERIAL_DEC, SERIAL_HEX}, SERIAL_DEC)
}

/*
Configuration structure.
*/
type tConfig struct {
	emitSshFormat          *tChoice
	emitX509Format         *tChoice
	showCrl                bool
	showKeyId              bool
	sortOrder              *tChoice
	showValid              bool
	showExpired            bool
	showRevoked            bool
	showProvisioner        bool
	timeFormat             *tChoice
	showDNSNames           bool
	showEmailAddresses     bool
	showIPAddresses        bool
	showURIs               bool
	showIssuer             bool
	showHostType           bool
	showSignatureAlgorithm bool
	serialFormat           *tChoice
}

/*
Alignment of markdown table.
*/
const (
	ALIGN_LEFT = iota
	ALIGN_CENTER
	ALIGN_RIGHT
)

/*
Certificate revocation information. Both ssh & x509.
*/
type tCertificateRevocation struct {
	Serial        string    `json:"-"`
	ProvisionerID string    `json:"ProvisionerID"`
	ReasonCode    int       `json:"ReasonCode"`
	Reason        string    `json:"Reason"`
	RevokedAt     time.Time `json:"RevokedAt"`
	ExpiresAt     time.Time `json:"ExpiresAt"`
	TokenID       string    `json:"TokenID"`
	MTLS          bool      `json:"MTLS"`
	ACME          bool      `json:"ACME"`
}

type tCertificateStringSerials struct {
	SerialDec string `json:"SerialDec"`
	SerialHex string `json:"SerialHex"`
}

const (
	VALID_STR   string = "Valid"
	EXPIRED_STR string = "Expired"
	REVOKED_STR string = "Revoked"
)

/*
validityColors maps a validity status string to its display color. Package-level so
per-cell colouring does not reallocate the map for every rendered cell.
*/
var validityColors = map[string]color.Attribute{
	VALID_STR:   color.FgGreen,
	EXPIRED_STR: color.FgHiBlack,
	REVOKED_STR: color.FgHiYellow,
}

/*
alignChars maps an ALIGN_* constant to its markdown header-separator marker.
*/
var alignChars = map[int]string{
	ALIGN_LEFT:   `:-`,
	ALIGN_CENTER: `:-:`,
	ALIGN_RIGHT:  `-:`,
}

/*
markdownEscaper escapes the characters that would otherwise be interpreted by
Markdown. The backslash rule comes first, so pre-existing backslashes are escaped
before any new escape is introduced — the same result as the old sequential
strings.Replace loop, in a single pass.
*/
var markdownEscaper = strings.NewReplacer(
	`\`, `\\`,
	"`", "\\`",
	"*", `\*`,
	"_", `\_`,
	"{", `\{`,
	"}", `\}`,
	"[", `\[`,
	"]", `\]`,
	"(", `\(`,
	")", `\)`,
	"#", `\#`,
	".", `\.`,
	"!", `\!`,
	"+", `\+`,
	"-", `\-`,
)

/*
escapeMarkdown returns same string but safeguarded against markdown interpretation.

	'text' Text to be safeguarded.
*/
func escapeMarkdown(thisText string) string {
	return markdownEscaper.Replace(thisText)
}
