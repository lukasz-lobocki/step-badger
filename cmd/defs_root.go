package cmd

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
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
)

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
}

/*
Configuration structure.
*/
type tConfig struct {
	emitSshFormat      *tChoice
	emitX509Format     *tChoice
	showCrl            bool
	showKeyId          bool
	sortOrder          *tChoice
	showValid          bool
	showExpired        bool
	showRevoked        bool
	showProvisioner    bool
	timeFormat         *tChoice
	showDNSNames       bool
	showEmailAddresses bool
	showIPAddresses    bool
	showURIs           bool
	showIssuer         bool
	showSerial         bool
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

const (
	VALID_STR   string = "Valid"
	EXPIRED_STR string = "Expired"
	REVOKED_STR string = "Revoked"
)

/*
getValidityColor maps given status string to appropriate color.
*/
func getValidityColor() map[string]color.Attribute {
	return map[string]color.Attribute{
		VALID_STR:   color.FgGreen,
		EXPIRED_STR: color.FgHiBlack,
		REVOKED_STR: color.FgHiYellow,
	}
}

/*
getAlignChar amps given alignment to appropriate markdown string to be used in header separator.
*/
func getAlignChar() map[int]string {
	return map[int]string{
		ALIGN_LEFT:   `:-`,
		ALIGN_CENTER: `:-:`,
		ALIGN_RIGHT:  `-:`,
	}
}

/*
escapeMarkdown returns same string but safeguarded against markdown interpretation.

	'text' Text to be safeguarded.
*/
func escapeMarkdown(thisText string) string {

	// These characters need to be escaped in Markdown in order to appear as literal characters instead of performing some markdown functions
	needEscape := []string{
		`\`, "`", "*", "_",
		"{", "}",
		"[", "]",
		"(", ")",
		"#", ".", "!",
		"+", "-",
	}

	for _, need := range needEscape {
		thisText = strings.Replace(thisText, need, `\`+need, -1)
	}

	return thisText
}
