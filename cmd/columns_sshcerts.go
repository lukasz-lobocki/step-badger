package cmd

import (
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

type tSshColumn struct {
	isShown         func(tConfig) bool
	title           func() string
	titleColor      color.Attribute
	contentSource   func(tSshCertificateAndRevocation) string
	contentColor    func(tSshCertificateAndRevocation) color.Attribute
	contentAlignMD  int
	contentEscapeMD bool
}

/*
getX509Columns defines look and content of table's emitted columns
*/
func getSshColumns() []tSshColumn {

	var thisColumns []tSshColumn

	thisColumns = append(thisColumns,

		tSshColumn{
			isShown:    func(_ tConfig) bool { return true },     // Always shown
			title:      func() string { return "Serial number" }, // Static title
			titleColor: color.Bold,

			contentSource:   func(x tSshCertificateAndRevocation) string { return strconv.FormatUint(x.SshCertificate.Serial, 10) },
			contentColor:    func(_ tSshCertificateAndRevocation) color.Attribute { return color.FgWhite }, // Static color
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: false,
		},

		tSshColumn{
			isShown:    func(_ tConfig) bool { return true }, // Always shown
			title:      func() string { return "Type" },      // Static title
			titleColor: color.Bold,

			contentSource: func(x tSshCertificateAndRevocation) string { return getThisCertType()[int(x.SshCertificate.CertType)] },
			contentColor: func(x tSshCertificateAndRevocation) color.Attribute {
				return getThisCertTypeColor()[int(x.SshCertificate.CertType)]
			}, // Dynamic color
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tSshColumn{
			isShown:    func(_ tConfig) bool { return true },        // Always shown
			title:      func() string { return "Valid principals" }, // Static title
			titleColor: color.Bold,

			contentSource: func(x tSshCertificateAndRevocation) string {
				return strings.Join(x.SshCertificate.ValidPrincipals, ",")
			},
			contentColor:    func(_ tSshCertificateAndRevocation) color.Attribute { return color.FgHiWhite }, // Static color
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tSshColumn{
			isShown:    func(tc tConfig) bool { return tc.showKeyId }, // Always shown
			title:      func() string { return "Key ID" },             // Static title
			titleColor: color.Bold,

			contentSource:   func(x tSshCertificateAndRevocation) string { return x.SshCertificate.KeyId },
			contentColor:    func(_ tSshCertificateAndRevocation) color.Attribute { return color.FgHiWhite }, // Static color
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tSshColumn{
			isShown:    func(_ tConfig) bool { return true }, // Always shown
			title:      func() string { return "Start" },     // Static title
			titleColor: color.Bold,

			contentSource: func(x tSshCertificateAndRevocation) string {
				return time.Unix(int64(x.SshCertificate.ValidAfter), 0).UTC().Format(time.DateOnly)
			},
			contentColor:    func(_ tSshCertificateAndRevocation) color.Attribute { return color.FgHiBlack }, // Static color
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tSshColumn{
			isShown:    func(_ tConfig) bool { return true }, // Always shown
			title:      func() string { return "Finish" },    // Static title
			titleColor: color.Bold,

			contentSource: func(x tSshCertificateAndRevocation) string {
				return time.Unix(int64(x.SshCertificate.ValidBefore), 0).UTC().Format(time.DateOnly)
			},
			contentColor:    func(_ tSshCertificateAndRevocation) color.Attribute { return color.FgHiBlack }, // Static color
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tSshColumn{
			isShown:    func(_ tConfig) bool { return true },  // Always shown
			title:      func() string { return "Revoked at" }, // Static title
			titleColor: color.Bold,

			contentSource: func(x tSshCertificateAndRevocation) string {
				if len(x.SshRevocation.ProvisionerID) > 0 {
					return x.SshRevocation.RevokedAt.UTC().Format(time.DateOnly)
				} else {
					return ""
				}
			},
			contentColor:    func(_ tSshCertificateAndRevocation) color.Attribute { return color.FgHiBlack }, // Static color
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tSshColumn{
			isShown:    func(_ tConfig) bool { return true }, // Always shown
			title:      func() string { return "Validity" },  // Static title
			titleColor: color.Bold,

			contentSource: func(x tSshCertificateAndRevocation) string {
				return x.Validity
			},
			contentColor: func(x tSshCertificateAndRevocation) color.Attribute {
				return getThisValidityColor()[x.Validity]
			}, // Dynamic color
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},
	)

	return thisColumns
}

/*
getThisCertType maps given CertType to string to be displayed.
*/
func getThisCertType() map[int]string {
	return map[int]string{
		1: `User`,
		2: `Host`,
	}
}

/*
getThisCertTypeColor maps given CertType to color to be used.
*/
func getThisCertTypeColor() map[int]color.Attribute {
	return map[int]color.Attribute{
		1: color.FgCyan,
		2: color.FgMagenta,
	}
}
