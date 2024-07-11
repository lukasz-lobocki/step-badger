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

type X509CertificateAndRevocationInfo struct {
	Certificate x509.Certificate           `json:"Certificate"`
	Revocation  RevokedX509CertificateInfo `json:"Revocation,omitempty"`
}

type RevokedX509CertificateInfo struct {
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
