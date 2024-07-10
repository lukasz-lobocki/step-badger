package cmd

import (
	"crypto/x509"
	"time"
)

/*
Configuration structure for exportCmd
*/
type ConfigInfo struct {
	emitFormat *tChoice
}

type CertificateAndRevocationInfo struct {
	Certificate x509.Certificate       `json:"Certificate"`
	Revocation  RevokedCertificateInfo `json:"Revocation,omitempty"`
}

type RevokedCertificateInfo struct {
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
