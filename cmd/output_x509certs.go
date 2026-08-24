package cmd

import (
	"fmt"
	"regexp"
	"time"
)

/*
opensslTimeRegex matches the non-alphanumeric separators in RFC3339 timestamps.
Hoisted so it is compiled once and shared by both usages below.
*/
var opensslTimeRegex = regexp.MustCompile(`[-T:]+`)

/*
emitX509OpenSsl prints result in the form of markdown table.

	'thisX509CertsWithRevocations' Slice of certs.
*/
func emitX509OpenSsl(thisX509CertsWithRevocations []tX509CertificateProvisionerRevocation) {
	for _, x509CertWithRevocation := range thisX509CertsWithRevocations {

		var revokedAt string

		// Construct RevokedAt string in compliance with specification.
		if len(x509CertWithRevocation.X509Revocation.ProvisionerID) > 0 {
			revokedAt = opensslTimeRegex.
				ReplaceAllString(x509CertWithRevocation.X509Revocation.RevokedAt.UTC().
					Format(time.RFC3339), "")[2:]
		} else {
			revokedAt = ""
		}

		fmt.Printf("%s\t%s\t%s\t%040X\t%s\t%s\n",
			x509CertWithRevocation.Validity[0:1],
			opensslTimeRegex.
				ReplaceAllString(x509CertWithRevocation.X509Certificate.NotAfter.UTC().
					Format(time.RFC3339), "")[2:], // Construct NotAfter string in compliance with specification.
			revokedAt,
			x509CertWithRevocation.X509Certificate.SerialNumber, // len(HEX(20bytes))=40chars [FF/byte]
			"unknown", // As per specification.
			x509CertWithRevocation.X509Certificate.Subject)
	}
}
