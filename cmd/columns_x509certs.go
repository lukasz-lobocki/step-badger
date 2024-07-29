package cmd

import (
	"strings"
	"time"

	"github.com/fatih/color"
)

type tX509Column struct {
	isShown         func(tConfig) bool
	title           func() string
	titleColor      color.Attribute
	contentSource   func(tX509CertificateWithRevocation, tConfig) string
	contentColor    func(tX509CertificateWithRevocation) color.Attribute
	contentAlignMD  int
	contentEscapeMD bool
}

/*
getX509Columns defines look and content of table's emitted columns.
*/
func getX509Columns() []tX509Column {

	var thisColumns []tX509Column

	thisColumns = append(thisColumns,

		tX509Column{
			isShown:    func(_ tConfig) bool { return true },     // Always shown.
			title:      func() string { return "Serial number" }, // Static title.
			titleColor: color.Bold,

			contentSource: func(x tX509CertificateWithRevocation, _ tConfig) string {
				return x.X509Certificate.SerialNumber.String()
			},
			contentColor:    func(_ tX509CertificateWithRevocation) color.Attribute { return color.FgWhite }, // Static color.
			contentAlignMD:  ALIGN_RIGHT,
			contentEscapeMD: false,
		},

		tX509Column{
			isShown:    func(_ tConfig) bool { return true }, // Always shown.
			title:      func() string { return "Subject" },   // Static title.
			titleColor: color.Bold,

			contentSource:   func(x tX509CertificateWithRevocation, _ tConfig) string { return x.X509Certificate.Subject.String() },
			contentColor:    func(_ tX509CertificateWithRevocation) color.Attribute { return color.FgHiWhite }, // Static color.
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tX509Column{
			isShown:    func(tc tConfig) bool { return tc.showDNSNames },
			title:      func() string { return "DNSNames" }, // Static title.
			titleColor: color.Bold,

			contentSource: func(x tX509CertificateWithRevocation, _ tConfig) string {
				return strings.Join(x.X509Certificate.DNSNames, ", ")
			},
			contentColor:    func(_ tX509CertificateWithRevocation) color.Attribute { return color.FgWhite }, // Static color.
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tX509Column{
			isShown:    func(tc tConfig) bool { return tc.showEmailAddresses },
			title:      func() string { return "EmailAddresses" }, // Static title.
			titleColor: color.Bold,

			contentSource: func(x tX509CertificateWithRevocation, _ tConfig) string {
				return strings.Join(x.X509Certificate.EmailAddresses, ", ")
			},
			contentColor:    func(_ tX509CertificateWithRevocation) color.Attribute { return color.FgWhite }, // Static color.
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tX509Column{
			isShown:    func(tc tConfig) bool { return tc.showIPAddresses },
			title:      func() string { return "IPAddresses" }, // Static title.
			titleColor: color.Bold,

			contentSource: func(x tX509CertificateWithRevocation, _ tConfig) string {
				var thisIPAddresses []string
				for _, thisIPAddress := range x.X509Certificate.IPAddresses {
					thisIPAddresses = append(thisIPAddresses, thisIPAddress.String())
				}
				return strings.Join(thisIPAddresses, ", ")
			},
			contentColor:    func(_ tX509CertificateWithRevocation) color.Attribute { return color.FgWhite }, // Static color.
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tX509Column{
			isShown:    func(tc tConfig) bool { return tc.showURIs },
			title:      func() string { return "URIs" }, // Static title.
			titleColor: color.Bold,

			contentSource: func(x tX509CertificateWithRevocation, _ tConfig) string {
				var thisUris []string
				for _, thisUri := range x.X509Certificate.URIs {
					thisUris = append(thisUris, thisUri.String())
				}
				return strings.Join(thisUris, ", ")
			},
			contentColor:    func(_ tX509CertificateWithRevocation) color.Attribute { return color.FgWhite }, // Static color.
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tX509Column{
			isShown:    func(tc tConfig) bool { return tc.showCrl },
			title:      func() string { return "CRLDistributionPoints" }, // Static title.
			titleColor: color.Bold,

			contentSource: func(x tX509CertificateWithRevocation, _ tConfig) string {
				return strings.Join(x.X509Certificate.CRLDistributionPoints, ", ")
			},
			contentColor:    func(_ tX509CertificateWithRevocation) color.Attribute { return color.FgWhite }, // Static color.
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tX509Column{
			isShown:    func(tc tConfig) bool { return tc.showProvisioner },
			title:      func() string { return "Provisioner" }, // Static title.
			titleColor: color.Bold,

			contentSource: func(x tX509CertificateWithRevocation, _ tConfig) string {
				return (x.X509Provisioner.Type + " " + x.X509Provisioner.Name[:min(len(x.X509Provisioner.Name), 6)])
			},
			contentColor:    func(_ tX509CertificateWithRevocation) color.Attribute { return color.FgWhite }, // Static color.
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tX509Column{
			isShown:    func(_ tConfig) bool { return true }, // Always shown.
			title:      func() string { return "Start" },     // Static title.
			titleColor: color.Bold,

			contentSource: func(x tX509CertificateWithRevocation, tc tConfig) string {
				if tc.timeFormat.Value == "s" {
					return x.X509Certificate.NotBefore.UTC().Format(time.DateOnly)
				} else {
					return x.X509Certificate.NotBefore.UTC().Format(time.RFC3339)
				}
			},
			contentColor:    func(_ tX509CertificateWithRevocation) color.Attribute { return color.FgHiBlack }, // Static color.
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tX509Column{
			isShown:    func(_ tConfig) bool { return true }, // Always shown.
			title:      func() string { return "Finish" },    // Static title.
			titleColor: color.Bold,

			contentSource: func(x tX509CertificateWithRevocation, tc tConfig) string {
				if tc.timeFormat.Value == "s" {
					return x.X509Certificate.NotAfter.UTC().Format(time.DateOnly)
				} else {
					return x.X509Certificate.NotAfter.UTC().Format(time.RFC3339)
				}
			},
			contentColor:    func(_ tX509CertificateWithRevocation) color.Attribute { return color.FgHiBlack }, // Static color.
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tX509Column{
			isShown:    func(_ tConfig) bool { return true },  // Always shown.
			title:      func() string { return "Revoked at" }, // Static title.
			titleColor: color.Bold,

			contentSource: func(x tX509CertificateWithRevocation, tc tConfig) string {
				if len(x.X509Revocation.ProvisionerID) > 0 {
					if tc.timeFormat.Value == "s" {
						return x.X509Revocation.RevokedAt.UTC().Format(time.DateOnly)
					} else {
						return x.X509Revocation.RevokedAt.UTC().Format(time.RFC3339)
					}
				} else {
					return ""
				}
			},
			contentColor:    func(_ tX509CertificateWithRevocation) color.Attribute { return color.FgHiBlack }, // Static color.
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tX509Column{
			isShown:    func(_ tConfig) bool { return true }, // Always shown.
			title:      func() string { return "Validity" },  // Static title.
			titleColor: color.Bold,

			contentSource: func(x tX509CertificateWithRevocation, _ tConfig) string {
				return x.Validity
			},
			contentColor: func(x tX509CertificateWithRevocation) color.Attribute {
				return getThisValidityColor()[x.Validity]
			}, // Dynamic color
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},
	)

	return thisColumns
}
