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
	contentSource   func(tSshCertificateWithRevocation, tConfig) string
	contentColor    func(tSshCertificateWithRevocation) color.Attribute
	contentAlignMD  int
	contentEscapeMD bool
}

/*
getX509Columns defines look and content of table's emitted columns.
*/
func getSshColumns() []tSshColumn {

	var columns []tSshColumn

	columns = append(columns,

		tSshColumn{
			isShown:    func(_ tConfig) bool { return true },     // Always shown.
			title:      func() string { return "Serial number" }, // Static title.
			titleColor: color.Bold,

			contentSource: func(x tSshCertificateWithRevocation, _ tConfig) string {
				return strconv.FormatUint(x.SshCertificate.Serial, 10)
			},

			contentColor:    func(_ tSshCertificateWithRevocation) color.Attribute { return color.FgWhite }, // Static color.
			contentAlignMD:  ALIGN_RIGHT,
			contentEscapeMD: false,
		},

		tSshColumn{
			isShown:    func(_ tConfig) bool { return true }, // Always shown.
			title:      func() string { return "Type" },      // Static title.
			titleColor: color.Bold,

			contentSource: func(x tSshCertificateWithRevocation, _ tConfig) string {
				return getCertType()[int(x.SshCertificate.CertType)]
			},

			contentColor: func(x tSshCertificateWithRevocation) color.Attribute {
				return getCertTypeColor()[int(x.SshCertificate.CertType)]
			}, // Dynamic color
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tSshColumn{
			isShown:    func(_ tConfig) bool { return true },        // Always shown.
			title:      func() string { return "Valid principals" }, // Static title.
			titleColor: color.Bold,

			contentSource: func(x tSshCertificateWithRevocation, _ tConfig) string {
				return strings.Join(x.SshCertificate.ValidPrincipals, ",")
			},

			contentColor:    func(_ tSshCertificateWithRevocation) color.Attribute { return color.FgHiWhite }, // Static color.
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tSshColumn{
			isShown:    func(tc tConfig) bool { return tc.showKeyId },
			title:      func() string { return "Key ID" }, // Static title.
			titleColor: color.Bold,

			contentSource: func(x tSshCertificateWithRevocation, _ tConfig) string { return x.SshCertificate.KeyId },

			contentColor:    func(_ tSshCertificateWithRevocation) color.Attribute { return color.FgHiWhite }, // Static color.
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tSshColumn{
			isShown:    func(_ tConfig) bool { return true }, // Always shown.
			title:      func() string { return "Start" },     // Static title.
			titleColor: color.Bold,

			contentSource: func(x tSshCertificateWithRevocation, tc tConfig) string {
				if tc.timeFormat.Value == "s" {
					return time.Unix(int64(x.SshCertificate.ValidAfter), 0).UTC().Format(time.DateOnly)
				} else {
					return time.Unix(int64(x.SshCertificate.ValidAfter), 0).UTC().Format(time.RFC3339)
				}
			},

			contentColor:    func(_ tSshCertificateWithRevocation) color.Attribute { return color.FgHiBlack }, // Static color.
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tSshColumn{
			isShown:    func(_ tConfig) bool { return true }, // Always shown.
			title:      func() string { return "Finish" },    // Static title.
			titleColor: color.Bold,

			contentSource: func(x tSshCertificateWithRevocation, tc tConfig) string {
				if tc.timeFormat.Value == "s" {
					return time.Unix(int64(x.SshCertificate.ValidBefore), 0).UTC().Format(time.DateOnly)
				} else {
					return time.Unix(int64(x.SshCertificate.ValidBefore), 0).UTC().Format(time.RFC3339)
				}
			},

			contentColor:    func(_ tSshCertificateWithRevocation) color.Attribute { return color.FgHiBlack }, // Static color.
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tSshColumn{
			isShown:    func(tc tConfig) bool { return tc.showRevoked }, // Always shown.
			title:      func() string { return "Revoked at" },           // Static title.
			titleColor: color.Bold,

			contentSource: func(x tSshCertificateWithRevocation, tc tConfig) string {
				if len(x.SshCertificateRevocation.ProvisionerID) > 0 {
					if tc.timeFormat.Value == "s" {
						return x.SshCertificateRevocation.RevokedAt.UTC().Format(time.DateOnly)
					} else {
						return x.SshCertificateRevocation.RevokedAt.UTC().Format(time.RFC3339)
					}
				} else {
					return ""
				}
			},

			contentColor:    func(_ tSshCertificateWithRevocation) color.Attribute { return color.FgHiBlack }, // Static color.
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tSshColumn{
			isShown:    func(_ tConfig) bool { return true }, // Always shown.
			title:      func() string { return "Validity" },  // Static title.
			titleColor: color.Bold,

			contentSource: func(x tSshCertificateWithRevocation, _ tConfig) string {
				return x.Validity
			},

			contentColor: func(x tSshCertificateWithRevocation) color.Attribute {
				return getValidityColor()[x.Validity]
			}, // Dynamic color
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},
	)

	return columns
}

/*
getCertType maps given CertType to string to be displayed.
*/
func getCertType() map[int]string {
	return map[int]string{
		1: "User",
		2: "Host",
	}
}

/*
getCertTypeColor maps given CertType to color to be used.
*/
func getCertTypeColor() map[int]color.Attribute {
	return map[int]color.Attribute{
		1: color.FgCyan,
		2: color.FgMagenta,
	}
}
