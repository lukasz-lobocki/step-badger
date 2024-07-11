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

type tColumn struct {
	isShown         func() bool
	title           func() string
	titleColor      color.Attribute
	contentSource   func(X509CertificateAndRevocationInfo) string
	contentColor    func(X509CertificateAndRevocationInfo) color.Attribute
	contentAlignMD  int
	contentEscapeMD bool
}

/*
getColumns defines look and content of table's emitted columns
*/
func getColumns() []tColumn {

	var thisColumns []tColumn

	thisColumns = append(thisColumns,

		tColumn{
			isShown:    func() bool { return true },              // Always shown
			title:      func() string { return "Serial number" }, // Static title
			titleColor: color.Bold,

			contentSource:   func(x X509CertificateAndRevocationInfo) string { return x.Certificate.SerialNumber.String() },
			contentColor:    func(_ X509CertificateAndRevocationInfo) color.Attribute { return color.FgWhite }, // Static color
			contentAlignMD:  ALIGN_RIGHT,
			contentEscapeMD: false,
		},

		tColumn{
			isShown:    func() bool { return true },        // Always shown
			title:      func() string { return "Subject" }, // Static title
			titleColor: color.Bold,

			contentSource:   func(x X509CertificateAndRevocationInfo) string { return x.Certificate.Subject.String() },
			contentColor:    func(_ X509CertificateAndRevocationInfo) color.Attribute { return color.FgHiWhite }, // Static color
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tColumn{
			isShown:    func() bool { return true },        // Always shown
			title:      func() string { return "Subject" }, // Static title
			titleColor: color.Bold,

			contentSource: func(x X509CertificateAndRevocationInfo) string {
				return strings.Join(x.Certificate.CRLDistributionPoints, ", ")
			},
			contentColor:    func(_ X509CertificateAndRevocationInfo) color.Attribute { return color.FgHiWhite }, // Static color
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tColumn{
			isShown:    func() bool { return true },           // Always shown
			title:      func() string { return "Not before" }, // Static title
			titleColor: color.Bold,

			contentSource: func(x X509CertificateAndRevocationInfo) string {
				return x.Certificate.NotBefore.UTC().Format(time.DateOnly)
			},
			contentColor:    func(_ X509CertificateAndRevocationInfo) color.Attribute { return color.FgHiBlack }, // Static color
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tColumn{
			isShown:    func() bool { return true },          // Always shown
			title:      func() string { return "Not after" }, // Static title
			titleColor: color.Bold,

			contentSource: func(x X509CertificateAndRevocationInfo) string {
				return x.Certificate.NotAfter.UTC().Format(time.DateOnly)
			},
			contentColor:    func(_ X509CertificateAndRevocationInfo) color.Attribute { return color.FgHiBlack }, // Static color
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tColumn{
			isShown:    func() bool { return true },           // Always shown
			title:      func() string { return "Revoked at" }, // Static title
			titleColor: color.Bold,

			contentSource: func(x X509CertificateAndRevocationInfo) string {
				if len(x.Revocation.ProvisionerID) > 0 {
					return x.Revocation.RevokedAt.UTC().Format(time.DateOnly)
				} else {
					return ""
				}
			},
			contentColor:    func(_ X509CertificateAndRevocationInfo) color.Attribute { return color.FgHiBlack }, // Static color
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tColumn{
			isShown:    func() bool { return true },         // Always shown
			title:      func() string { return "Validity" }, // Static title
			titleColor: color.Bold,

			contentSource: func(x X509CertificateAndRevocationInfo) string {
				if len(x.Revocation.ProvisionerID) > 0 && time.Now().After(x.Revocation.RevokedAt) {
					return "Revoked"
				} else {
					if time.Now().After(x.Certificate.NotAfter) {
						return "Expired"
					} else {
						return "Valid"
					}
				}
			},
			contentColor: func(x X509CertificateAndRevocationInfo) color.Attribute {
				if len(x.Revocation.ProvisionerID) > 0 && time.Now().After(x.Revocation.RevokedAt) {
					return color.FgHiYellow
				} else {
					if time.Now().After(x.Certificate.NotAfter) {
						return color.FgHiBlack
					} else {
						return color.FgGreen
					}
				}
			}, // Static color
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},
	)

	return thisColumns
}
