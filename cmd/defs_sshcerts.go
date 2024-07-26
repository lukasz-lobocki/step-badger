package cmd

import "golang.org/x/crypto/ssh"

/*
Combined information of certificate, revocation and provisioner.
*/
type tSshCertificateWithRevocation struct {
	SshCertificate ssh.Certificate     `json:"Certificate"`
	Validity       string              `json:"Validity"`
	SshRevocation  tRevokedCertificate `json:"Revocation,omitempty"`
}
