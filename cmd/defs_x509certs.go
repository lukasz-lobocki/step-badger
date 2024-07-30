package cmd

import (
	"crypto/x509"
)

/*
Combined information of certificate, revocation and provisioner.
*/
type tX509CertificateProvisionerRevocation struct {
	X509Certificate x509.Certificate            `json:"Certificate"`
	Validity        string                      `json:"Validity"`
	X509Revocation  tCertificateRevocation      `json:"Revocation,omitempty"`
	X509Provisioner tX509CertificateProvisioner `json:"Provisioner,omitempty"`
}

/*
Intermediate structure to store certificate provisioner information.
*/
type tX509CertificateData struct {
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
