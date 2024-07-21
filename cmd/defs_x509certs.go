package cmd

import (
	"crypto/x509"
	"time"
)

/*
Combined information of certificate, revocation and provisioner.
*/
type tX509CertificateAndRevocation struct {
	X509Certificate x509.Certificate            `json:"Certificate"`
	X509Revocation  tX509RevokedCertificate     `json:"Revocation,omitempty"`
	X509Provisioner tX509CertificateProvisioner `json:"Provisioner,omitempty"`
}

/*
Certificate revocation information.
*/
type tX509RevokedCertificate struct {
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

type tX509Certificate struct {
	Provisioner tX509CertificateProvisioner `json:"Provisioner,omitempty"`
	RaInfo      *string                     `json:"-"`
}

/*
Certificate provisioner information.
*/
type tX509CertificateProvisioner struct {
	ID   string `json:"ID"`
	Name string `json:"Name"`
	Type string `json:"Type"`
}
