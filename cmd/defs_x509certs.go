package cmd

import (
	"crypto/x509"
)

/*
Combined information of certificate, revocation and provisioner.
*/
type tX509CertificateWithRevocation struct {
	X509Certificate x509.Certificate            `json:"Certificate"`
	Validity        string                      `json:"Validity"`
	X509Revocation  tRevokedCertificate         `json:"Revocation,omitempty"`
	X509Provisioner tX509CertificateProvisioner `json:"Provisioner,omitempty"`
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
