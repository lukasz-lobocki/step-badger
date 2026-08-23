package cmd

import (
	"strings"
	"time"

	"github.com/fatih/color"
)

/*
getX509Columns defines look and content of table's emitted columns.
*/
func getX509Columns() []tColumn[tX509CertificateProvisionerRevocation] {

	var columns []tColumn[tX509CertificateProvisionerRevocation]

	columns = append(columns,

		tColumn[tX509CertificateProvisionerRevocation]{
			isShown:    func(tc tConfig) bool { return true },
			title:      func() string { return "Serial number" }, // Static title.
			titleColor: color.Bold,

			contentSource: func(x tX509CertificateProvisionerRevocation, tc tConfig) string {
				if tc.serialFormat.Value == SERIAL_DEC {
					return x.X509CertificateStringSerials.SerialDec
				} else {
					return x.X509CertificateStringSerials.SerialHex
				}
			},

			contentColor:    func(_ tX509CertificateProvisionerRevocation) color.Attribute { return color.FgWhite }, // Static color.
			contentAlignMD:  ALIGN_RIGHT,
			contentEscapeMD: false,
		},

		tColumn[tX509CertificateProvisionerRevocation]{
			isShown:    func(_ tConfig) bool { return true }, // Always shown.
			title:      func() string { return "Subject" },   // Static title.
			titleColor: color.Bold,

			contentSource: func(x tX509CertificateProvisionerRevocation, _ tConfig) string {
				// return x.X509Certificate.SignatureAlgorithm.String()
				return x.X509Certificate.Subject.String()
			},

			contentColor:    func(_ tX509CertificateProvisionerRevocation) color.Attribute { return color.FgHiYellow }, // Static color.
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tColumn[tX509CertificateProvisionerRevocation]{
			isShown:    func(tc tConfig) bool { return tc.showIssuer },
			title:      func() string { return "Issuer" }, // Static title.
			titleColor: color.Bold,

			contentSource: func(x tX509CertificateProvisionerRevocation, _ tConfig) string {
				return x.X509Certificate.Issuer.CommonName
			},

			contentColor:    func(_ tX509CertificateProvisionerRevocation) color.Attribute { return color.FgWhite }, // Static color.
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tColumn[tX509CertificateProvisionerRevocation]{
			isShown:    func(tc tConfig) bool { return tc.showDNSNames },
			title:      func() string { return "DNS names" }, // Static title.
			titleColor: color.Bold,

			contentSource: func(x tX509CertificateProvisionerRevocation, _ tConfig) string {
				return strings.Join(x.X509Certificate.DNSNames, ", ")
			},

			contentColor:    func(_ tX509CertificateProvisionerRevocation) color.Attribute { return color.FgWhite }, // Static color.
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tColumn[tX509CertificateProvisionerRevocation]{
			isShown:    func(tc tConfig) bool { return tc.showEmailAddresses },
			title:      func() string { return "Email addresses" }, // Static title.
			titleColor: color.Bold,

			contentSource: func(x tX509CertificateProvisionerRevocation, _ tConfig) string {
				return strings.Join(x.X509Certificate.EmailAddresses, ", ")
			},

			contentColor:    func(_ tX509CertificateProvisionerRevocation) color.Attribute { return color.FgWhite }, // Static color.
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tColumn[tX509CertificateProvisionerRevocation]{
			isShown:    func(tc tConfig) bool { return tc.showIPAddresses },
			title:      func() string { return "IP addresses" }, // Static title.
			titleColor: color.Bold,

			contentSource: func(x tX509CertificateProvisionerRevocation, _ tConfig) string {
				var ipAddresses []string
				for _, ipAddress := range x.X509Certificate.IPAddresses {
					ipAddresses = append(ipAddresses, ipAddress.String())
				}
				return strings.Join(ipAddresses, ", ")
			},

			contentColor:    func(_ tX509CertificateProvisionerRevocation) color.Attribute { return color.FgWhite }, // Static color.
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tColumn[tX509CertificateProvisionerRevocation]{
			isShown:    func(tc tConfig) bool { return tc.showURIs },
			title:      func() string { return "URIs" }, // Static title.
			titleColor: color.Bold,

			contentSource: func(x tX509CertificateProvisionerRevocation, _ tConfig) string {
				var uris []string
				for _, uri := range x.X509Certificate.URIs {
					uris = append(uris, uri.String())
				}
				return strings.Join(uris, ", ")
			},

			contentColor:    func(_ tX509CertificateProvisionerRevocation) color.Attribute { return color.FgWhite }, // Static color.
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tColumn[tX509CertificateProvisionerRevocation]{
			isShown:    func(tc tConfig) bool { return tc.showCrl },
			title:      func() string { return "CRL distribution points" }, // Static title.
			titleColor: color.Bold,

			contentSource: func(x tX509CertificateProvisionerRevocation, _ tConfig) string {
				return strings.Join(x.X509Certificate.CRLDistributionPoints, ", ")
			},

			contentColor:    func(_ tX509CertificateProvisionerRevocation) color.Attribute { return color.FgWhite }, // Static color.
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tColumn[tX509CertificateProvisionerRevocation]{
			isShown:    func(tc tConfig) bool { return tc.showProvisioner },
			title:      func() string { return "Provisioner" }, // Static title.
			titleColor: color.Bold,

			contentSource: func(x tX509CertificateProvisionerRevocation, _ tConfig) string {
				return (x.X509Provisioner.Type + " " + x.X509Provisioner.Name[:min(len(x.X509Provisioner.Name), 6)])
			},

			contentColor:    func(_ tX509CertificateProvisionerRevocation) color.Attribute { return color.FgWhite }, // Static color.
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tColumn[tX509CertificateProvisionerRevocation]{
			isShown:    func(tc tConfig) bool { return tc.showSignatureAlgorithm },
			title:      func() string { return "Algorithm" }, // Static title.
			titleColor: color.Bold,

			contentSource: func(x tX509CertificateProvisionerRevocation, _ tConfig) string {
				return x.X509Certificate.SignatureAlgorithm.String()
			},

			contentColor:    func(_ tX509CertificateProvisionerRevocation) color.Attribute { return color.FgWhite }, // Static color.
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tColumn[tX509CertificateProvisionerRevocation]{
			isShown:    func(_ tConfig) bool { return true }, // Always shown.
			title:      func() string { return "Start" },     // Static title.
			titleColor: color.Bold,

			contentSource: func(x tX509CertificateProvisionerRevocation, tc tConfig) string {
				if tc.timeFormat.Value == TIME_SHORT {
					return x.X509Certificate.NotBefore.UTC().Format(time.DateOnly)
				} else {
					return x.X509Certificate.NotBefore.UTC().Format(time.RFC3339)
				}
			},

			contentColor:    func(_ tX509CertificateProvisionerRevocation) color.Attribute { return color.FgHiBlack }, // Static color.
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tColumn[tX509CertificateProvisionerRevocation]{
			isShown:    func(_ tConfig) bool { return true }, // Always shown.
			title:      func() string { return "Finish" },    // Static title.
			titleColor: color.Bold,

			contentSource: func(x tX509CertificateProvisionerRevocation, tc tConfig) string {
				if tc.timeFormat.Value == TIME_SHORT {
					return x.X509Certificate.NotAfter.UTC().Format(time.DateOnly)
				} else {
					return x.X509Certificate.NotAfter.UTC().Format(time.RFC3339)
				}
			},

			contentColor:    func(_ tX509CertificateProvisionerRevocation) color.Attribute { return color.FgHiBlack }, // Static color.
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tColumn[tX509CertificateProvisionerRevocation]{
			isShown:    func(tc tConfig) bool { return tc.showRevoked }, // Always shown.
			title:      func() string { return "Revoked at" },           // Static title.
			titleColor: color.Bold,

			contentSource: func(x tX509CertificateProvisionerRevocation, tc tConfig) string {
				if len(x.X509Revocation.ProvisionerID) > 0 {
					if tc.timeFormat.Value == TIME_SHORT {
						return x.X509Revocation.RevokedAt.UTC().Format(time.DateOnly)
					} else {
						return x.X509Revocation.RevokedAt.UTC().Format(time.RFC3339)
					}
				} else {
					return ""
				}
			},

			contentColor:    func(_ tX509CertificateProvisionerRevocation) color.Attribute { return color.FgHiBlack }, // Static color.
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tColumn[tX509CertificateProvisionerRevocation]{
			isShown:    func(_ tConfig) bool { return true }, // Always shown.
			title:      func() string { return "Validity" },  // Static title.
			titleColor: color.Bold,

			contentSource: func(x tX509CertificateProvisionerRevocation, _ tConfig) string {
				return x.Validity
			},

			contentColor: func(x tX509CertificateProvisionerRevocation) color.Attribute {
				return getValidityColor()[x.Validity]
			}, // Dynamic color
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},
	)

	return columns
}
