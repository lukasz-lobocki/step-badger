package cmd

import (
	"crypto/x509"
	"time"
)

/*
Configuration structure for exportCmd
*/
type tConfig struct {
	emitFormat *tChoice
}

type tCertificate struct {
	Cert       x509.Certificate       `json:"topLevelPath"` // Full path
	Revocation tCertificateRevocation `json:"uniqueName"`   // Shortest unique path
}

type tCertificateRevocation struct {
	Serial        string    `json:"Serial"`
	ProvisionerID string    `json:"ProvisionerID"`
	ReasonCode    int       `json:"ReasonCode"`
	Reason        string    `json:"Reason"`
	RevokedAt     time.Time `json:"RevokedAt"`
	ExpiresAt     time.Time `json:"ExpiresAt"`
	TokenID       string    `json:"TokenID"`
	MTLS          bool      `json:"MTLS"`
	ACME          bool      `json:"ACME"`
}
