package cmd

import (
	"strings"
	"time"

	"github.com/fatih/color"
)

const (
	ALIGN_LEFT = iota
	ALIGN_CENTER
	ALIGN_RIGHT
)

type tX509Column struct {
	isShown         func() bool
	title           func() string
	titleColor      color.Attribute
	contentSource   func(tX509CertificateAndRevocation) string
	contentColor    func(tX509CertificateAndRevocation) color.Attribute
	contentAlignMD  int
	contentEscapeMD bool
}

/*
getX509Columns defines look and content of table's emitted columns
*/
func getX509Columns() []tX509Column {

	var thisColumns []tX509Column

	thisColumns = append(thisColumns,

		tX509Column{
			isShown:    func() bool { return true },              // Always shown
			title:      func() string { return "Serial number" }, // Static title
			titleColor: color.Bold,

			contentSource:   func(x tX509CertificateAndRevocation) string { return x.X509Certificate.SerialNumber.String() },
			contentColor:    func(_ tX509CertificateAndRevocation) color.Attribute { return color.FgWhite }, // Static color
			contentAlignMD:  ALIGN_RIGHT,
			contentEscapeMD: false,
		},

		tX509Column{
			isShown:    func() bool { return true },        // Always shown
			title:      func() string { return "Subject" }, // Static title
			titleColor: color.Bold,

			contentSource:   func(x tX509CertificateAndRevocation) string { return x.X509Certificate.Subject.String() },
			contentColor:    func(_ tX509CertificateAndRevocation) color.Attribute { return color.FgHiWhite }, // Static color
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tX509Column{
			isShown:    func() bool { return true },                      // Always shown
			title:      func() string { return "CRLDistributionPoints" }, // Static title
			titleColor: color.Bold,

			contentSource: func(x tX509CertificateAndRevocation) string {
				return strings.Join(x.X509Certificate.CRLDistributionPoints, ", ")
			},
			contentColor:    func(_ tX509CertificateAndRevocation) color.Attribute { return color.FgHiWhite }, // Static color
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tX509Column{
			isShown:    func() bool { return true },           // Always shown
			title:      func() string { return "Not before" }, // Static title
			titleColor: color.Bold,

			contentSource: func(x tX509CertificateAndRevocation) string {
				return x.X509Certificate.NotBefore.UTC().Format(time.DateOnly)
			},
			contentColor:    func(_ tX509CertificateAndRevocation) color.Attribute { return color.FgHiBlack }, // Static color
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tX509Column{
			isShown:    func() bool { return true },          // Always shown
			title:      func() string { return "Not after" }, // Static title
			titleColor: color.Bold,

			contentSource: func(x tX509CertificateAndRevocation) string {
				return x.X509Certificate.NotAfter.UTC().Format(time.DateOnly)
			},
			contentColor:    func(_ tX509CertificateAndRevocation) color.Attribute { return color.FgHiBlack }, // Static color
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tX509Column{
			isShown:    func() bool { return true },           // Always shown
			title:      func() string { return "Revoked at" }, // Static title
			titleColor: color.Bold,

			contentSource: func(x tX509CertificateAndRevocation) string {
				if len(x.X509Revocation.ProvisionerID) > 0 {
					return x.X509Revocation.RevokedAt.UTC().Format(time.DateOnly)
				} else {
					return ""
				}
			},
			contentColor:    func(_ tX509CertificateAndRevocation) color.Attribute { return color.FgHiBlack }, // Static color
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tX509Column{
			isShown:    func() bool { return true },         // Always shown
			title:      func() string { return "Validity" }, // Static title
			titleColor: color.Bold,

			contentSource: func(x tX509CertificateAndRevocation) string {
				if len(x.X509Revocation.ProvisionerID) > 0 && time.Now().After(x.X509Revocation.RevokedAt) {
					return "Revoked"
				} else {
					if time.Now().After(x.X509Certificate.NotAfter) {
						return "Expired"
					} else {
						return "Valid"
					}
				}
			},
			contentColor: func(x tX509CertificateAndRevocation) color.Attribute {
				if len(x.X509Revocation.ProvisionerID) > 0 && time.Now().After(x.X509Revocation.RevokedAt) { // Dynamic color
					return color.FgHiYellow // Revoked
				} else {
					if time.Now().After(x.X509Certificate.NotAfter) {
						return color.FgHiBlack // Expired
					} else {
						return color.FgGreen // Valid
					}
				}
			}, // Static color
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},
	)

	return thisColumns
}
