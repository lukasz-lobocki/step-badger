package cmd

import (
	"log"
	"os"
	"time"

	"github.com/fatih/color"
)

const (
	MAX_LOGGING_LEVEL int = 3 // Maximum allowed logging level
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
initChoices sets up Config struct for 'limited choice' flag
*/
func initChoices() {
	config.emitFormat = newChoice([]string{"t", "j"}, "t")
	config.sortOrder = newChoice([]string{"s", "f"}, "f")
}

/*
Configuration structure
*/
type tConfig struct {
	emitFormat      *tChoice
	showCrl         bool
	showKeyId       bool
	sortOrder       *tChoice
	showValid       bool
	showExpired     bool
	showRevoked     bool
	showProvisioner bool
}

/*
Alignment of markdown table
*/
const (
	ALIGN_LEFT = iota
	ALIGN_CENTER
	ALIGN_RIGHT
)

/*
Certificate revocation information. Bots ssh & x509.
*/
type tRevokedCertificate struct {
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
getThisABColor maps given ahead / behind status string to appropriate color
*/
func getThisValidityColor() map[string]color.Attribute {
	return map[string]color.Attribute{
		VALID_STR:   color.FgGreen,
		EXPIRED_STR: color.FgHiBlack,
		REVOKED_STR: color.FgHiYellow,
	}
}
