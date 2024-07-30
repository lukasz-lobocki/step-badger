package cmd

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

const (
	MAX_LOGGING_LEVEL int = 3 // Maximum allowed logging level.
)

/*
initLoggers creates colorful loggers.
*/
func initLoggers() {
	thisHiCyan := color.New(color.FgHiCyan).SprintFunc()
	thisHiYellow := color.New(color.FgHiYellow).SprintFunc()
	thisHiRed := color.New(color.FgHiRed).SprintFunc()

	logInfo = log.New(os.Stderr, thisHiCyan("╭info\n╰"), 0)
	logWarning = log.New(os.Stderr, thisHiYellow("╭warning\n╰"), log.Lshortfile)
	logError = log.New(os.Stderr, thisHiRed("╭error\n╰"), log.Lshortfile)

}

/*
initChoices sets up Config struct for 'limited choice' flag.
*/
func initChoices() {
	config.emitSshFormat = newChoice([]string{"t", "j", "m"}, "t")
	config.emitX509Format = newChoice([]string{"t", "j", "m", "o"}, "t")
	config.sortOrder = newChoice([]string{"s", "f"}, "f")
	config.timeFormat = newChoice([]string{"i", "s"}, "i")
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
func escapeMarkdown(text string) string {

	// These characters need to be escaped in Markdown in order to appear as literal characters instead of performing some markdown functions
	needEscape := []string{
		`\`, "`", "*", "_",
		"{", "}",
		"[", "]",
		"(", ")",
		"#", ".", "!",
		"+", "-",
	}

	for _, thisNeed := range needEscape {
		text = strings.Replace(text, thisNeed, `\`+thisNeed, -1)
	}

	return text
}
