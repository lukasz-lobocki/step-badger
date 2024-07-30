package cmd

import "golang.org/x/crypto/ssh"

/*
Combined information of certificate and revocation.
*/
type tSshCertificateWithRevocation struct {
	SshCertificate ssh.Certificate        `json:"Certificate"`
	Validity       string                 `json:"Validity"`
	SshRevocation  tCertificateRevocation `json:"Revocation,omitempty"`
}
