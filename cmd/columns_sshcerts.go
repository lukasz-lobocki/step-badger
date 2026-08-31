package cmd

import (
	"strings"
	"time"

	"github.com/fatih/color"
)

/*
getSshColumns defines look and content of table's emitted columns.
*/
func getSshColumns() []tColumn[tSshCertificateWithRevocation] {

	var columns []tColumn[tSshCertificateWithRevocation]

	columns = append(columns,

		tColumn[tSshCertificateWithRevocation]{
			isShown:    func(tc tConfig) bool { return true },
			title:      func() string { return "Serial number" }, // Static title.
			titleColor: color.Bold,

			contentSource: func(x tSshCertificateWithRevocation, tc tConfig) string {
				if tc.serialFormat.Value == SERIAL_DEC {
					return x.SshCertificateStringSerials.SerialDec
				} else {
					return x.SshCertificateStringSerials.SerialHex
				}
			},

			contentColor:    func(_ tSshCertificateWithRevocation) color.Attribute { return color.FgWhite }, // Static color.
			contentAlignMD:  ALIGN_RIGHT,
			contentEscapeMD: false,
		},

		tColumn[tSshCertificateWithRevocation]{
			isShown:    func(_ tConfig) bool { return true },        // Always shown.
			title:      func() string { return "Valid principals" }, // Static title.
			titleColor: color.Bold,

			contentSource: func(x tSshCertificateWithRevocation, _ tConfig) string {
				return strings.Join(x.SshCertificate.ValidPrincipals, ",")
			},

			contentColor:    func(_ tSshCertificateWithRevocation) color.Attribute { return color.FgHiYellow }, // Static color.
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tColumn[tSshCertificateWithRevocation]{
			isShown:    func(tc tConfig) bool { return tc.showHostType },
			title:      func() string { return "Type" }, // Static title.
			titleColor: color.Bold,

			contentSource: func(x tSshCertificateWithRevocation, _ tConfig) string {
				return certTypeMap[int(x.SshCertificate.CertType)]
			},

			contentColor: func(x tSshCertificateWithRevocation) color.Attribute {
				return certTypeColors[int(x.SshCertificate.CertType)]
			}, // Dynamic color
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tColumn[tSshCertificateWithRevocation]{
			isShown:    func(tc tConfig) bool { return tc.showKeyId },
			title:      func() string { return "Key ID" }, // Static title.
			titleColor: color.Bold,

			contentSource: func(x tSshCertificateWithRevocation, _ tConfig) string { return x.SshCertificate.KeyId },

			contentColor:    func(_ tSshCertificateWithRevocation) color.Attribute { return color.FgHiWhite }, // Static color.
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tColumn[tSshCertificateWithRevocation]{
			isShown:    func(tc tConfig) bool { return tc.showSignatureAlgorithm },
			title:      func() string { return "Algorithm" }, // Static title.
			titleColor: color.Bold,

			contentSource: func(x tSshCertificateWithRevocation, _ tConfig) string {
				return x.SshCertificate.Signature.Format
			},

			contentColor:    func(_ tSshCertificateWithRevocation) color.Attribute { return color.FgWhite }, // Static color.
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tColumn[tSshCertificateWithRevocation]{
			isShown:    func(_ tConfig) bool { return true }, // Always shown.
			title:      func() string { return "Start" },     // Static title.
			titleColor: color.Bold,

			contentSource: func(x tSshCertificateWithRevocation, tc tConfig) string {
				if tc.timeFormat.Value == TIME_SHORT {
					return time.Unix(int64(x.SshCertificate.ValidAfter), 0).UTC().Format(time.DateOnly)
				} else {
					return time.Unix(int64(x.SshCertificate.ValidAfter), 0).UTC().Format(time.RFC3339)
				}
			},

			contentColor:    func(_ tSshCertificateWithRevocation) color.Attribute { return color.FgHiBlack }, // Static color.
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tColumn[tSshCertificateWithRevocation]{
			isShown:    func(_ tConfig) bool { return true }, // Always shown.
			title:      func() string { return "Finish" },    // Static title.
			titleColor: color.Bold,

			contentSource: func(x tSshCertificateWithRevocation, tc tConfig) string {
				if tc.timeFormat.Value == TIME_SHORT {
					return time.Unix(int64(x.SshCertificate.ValidBefore), 0).UTC().Format(time.DateOnly)
				} else {
					return time.Unix(int64(x.SshCertificate.ValidBefore), 0).UTC().Format(time.RFC3339)
				}
			},

			contentColor:    func(_ tSshCertificateWithRevocation) color.Attribute { return color.FgHiBlack }, // Static color.
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tColumn[tSshCertificateWithRevocation]{
			isShown:    func(tc tConfig) bool { return tc.showRevoked },
			title:      func() string { return "Revoked at" }, // Static title.
			titleColor: color.Bold,

			contentSource: func(x tSshCertificateWithRevocation, tc tConfig) string {
				if len(x.SshCertificateRevocation.ProvisionerID) > 0 {
					if tc.timeFormat.Value == TIME_SHORT {
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

		tColumn[tSshCertificateWithRevocation]{
			isShown:    func(_ tConfig) bool { return true }, // Always shown.
			title:      func() string { return "Validity" },  // Static title.
			titleColor: color.Bold,

			contentSource: func(x tSshCertificateWithRevocation, _ tConfig) string {
				return x.Validity
			},

			contentColor: func(x tSshCertificateWithRevocation) color.Attribute {
				return validityColors[x.Validity]
			}, // Dynamic color
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},
	)

	return columns
}

/*
certTypeMap maps given CertType to string to be displayed.
*/
var certTypeMap = map[int]string{
	1: "User",
	2: "Host",
}

/*
certTypeColors maps given CertType to color to be used.
*/
var certTypeColors = map[int]color.Attribute{
	1: color.FgCyan,
	2: color.FgMagenta,
}
