package cmd

import (
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"golang.org/x/crypto/ssh"
)

type tSshColumn struct {
	isShown         func(tConfig) bool
	title           func() string
	titleColor      color.Attribute
	contentSource   func(ssh.Certificate) string
	contentColor    func(ssh.Certificate) color.Attribute
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

			contentSource:   func(x ssh.Certificate) string { return strconv.FormatUint(x.Serial, 10) },
			contentColor:    func(_ ssh.Certificate) color.Attribute { return color.FgWhite }, // Static color
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: false,
		},

		tSshColumn{
			isShown:    func(_ tConfig) bool { return true },        // Always shown
			title:      func() string { return "Valid principals" }, // Static title
			titleColor: color.Bold,

			contentSource:   func(x ssh.Certificate) string { return strings.Join(x.ValidPrincipals, ",") },
			contentColor:    func(_ ssh.Certificate) color.Attribute { return color.FgHiWhite }, // Static color
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tSshColumn{
			isShown:    func(tc tConfig) bool { return tc.showKeyId }, // Always shown
			title:      func() string { return "Key ID" },             // Static title
			titleColor: color.Bold,

			contentSource:   func(x ssh.Certificate) string { return x.KeyId },
			contentColor:    func(_ ssh.Certificate) color.Attribute { return color.FgHiWhite }, // Static color
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tSshColumn{
			isShown:    func(_ tConfig) bool { return true },  // Always shown
			title:      func() string { return "Not before" }, // Static title
			titleColor: color.Bold,

			contentSource: func(x ssh.Certificate) string {
				return time.Unix(int64(x.ValidAfter), 0).UTC().Format(time.DateOnly)
			},
			contentColor:    func(_ ssh.Certificate) color.Attribute { return color.FgHiBlack }, // Static color
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tSshColumn{
			isShown:    func(_ tConfig) bool { return true }, // Always shown
			title:      func() string { return "Not after" }, // Static title
			titleColor: color.Bold,

			contentSource: func(x ssh.Certificate) string {
				return time.Unix(int64(x.ValidBefore), 0).UTC().Format(time.DateOnly)
			},
			contentColor:    func(_ ssh.Certificate) color.Attribute { return color.FgHiBlack }, // Static color
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},

		tSshColumn{
			isShown:    func(_ tConfig) bool { return true }, // Always shown
			title:      func() string { return "Validity" },  // Static title
			titleColor: color.Bold,

			contentSource: func(x ssh.Certificate) string {

				if time.Now().After(time.Unix(int64(x.ValidBefore), 0)) {
					return "Expired"
				} else {
					return "Valid"
				}

			},
			contentColor: func(x ssh.Certificate) color.Attribute {

				if time.Now().After(time.Unix(int64(x.ValidBefore), 0)) {
					return color.FgHiBlack // Expired
				} else {
					return color.FgGreen // Valid
				}

			}, // Static color
			contentAlignMD:  ALIGN_LEFT,
			contentEscapeMD: true,
		},
	)

	return thisColumns
}
