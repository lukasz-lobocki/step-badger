package cmd

import (
	"crypto/x509"
	"time"
)

type X509CertificateAndRevocationInfo struct {
	X509Certificate x509.Certificate               `json:"Certificate"`
	X509Revocation  X509RevokedCertificateInfo     `json:"Revocation,omitempty"`
	X509Provisioner X509CertificateProvisionerInfo `json:"Provisioner,omitempty"`
}

type X509RevokedCertificateInfo struct {
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

type X509CertificateInfo struct {
	Provisioner X509CertificateProvisionerInfo `json:"Provisioner,omitempty"`
	RaInfo      *string                        `json:"-"`
}

type X509CertificateProvisionerInfo struct {
	ID   string `json:"ID"`
	Name string `json:"Name"`
	Type string `json:"Type"`
}
