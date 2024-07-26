package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/lukasz-lobocki/tabby"
)

/*
emitSshCertsTable prints result in the form of a table.

	'thisSshCerts' slice of structures describing the ssh certificates
*/
func emitSshCertsTable(thisSshCerts []tSshCertificateWithRevocation) {
	table := new(tabby.Table)

	thisColumns := getSshColumns()

	var thisHeader []string
	/* Building slice of titles */
	for _, thisColumn := range thisColumns {
		if thisColumn.isShown(config) {
			thisHeader = append(thisHeader,
				color.New(thisColumn.titleColor).SprintFunc()(
					thisColumn.title(),
				),
			)
		}
	}

	/* Set the header */

	if err := table.SetHeader(thisHeader); err != nil {
		logError.Panic("Setting header failed. %w", err)
	}

	if loggingLevel >= 1 {
		logInfo.Println("header set.")
	}

	/* Populate the table */

	for _, sshCert := range thisSshCerts {

		var thisRow []string
		/* Building slice of columns within a single row*/

		for _, thisColumn := range thisColumns {

			if thisColumn.isShown(config) {
				thisRow = append(thisRow,
					color.New(thisColumn.contentColor(sshCert)).SprintFunc()(
						thisColumn.contentSource(sshCert, config),
					),
				)
			}
		}

		if err := table.AppendRow(thisRow); err != nil {
			logError.Panic(err)
		}
		if loggingLevel >= 3 {
			logInfo.Printf("row [%s] appended.", strconv.FormatUint(sshCert.SshCertificate.Serial, 10))
		}

	}

	if loggingLevel >= 2 {
		logInfo.Printf("%d rows appended.\n", len(thisSshCerts))
	}

	/* Emit the table */

	if loggingLevel >= 3 {
		table.Print(&tabby.Config{Spacing: "|", Padding: "."})
	} else {
		table.Print(nil)
	}
}

/*
emitSshCertsJson prints result in the form of a json

	'thisSshCerts' slice of structures describing the ssh certificates
*/
func emitSshCertsJson(thisSshCerts []tSshCertificateWithRevocation) {
	jsonInfo, err := json.MarshalIndent(thisSshCerts, "", "  ")
	if err != nil {
		logError.Panic(err)
	}
	fmt.Println(string(jsonInfo))
	if loggingLevel >= 2 {
		logInfo.Printf("%d records marshalled.\n", len(thisSshCerts))
	}
}

/*
emitX509Markdown prints result in the form of markdown table

	'thisSshCerts' slice of structures describing the certs.
*/
func emitSshCertsMarkdown(thisSshCertificatesWithRevocations []tSshCertificateWithRevocation) {
	thisColumns := getSshColumns()

	var thisHeader []string

	/* Building slice of titles */

	for _, thisColumn := range thisColumns {
		if thisColumn.isShown(config) {
			thisHeader = append(thisHeader, thisColumn.title())
		}
	}

	/* Emitting titles */

	fmt.Println("| " + strings.Join(thisHeader, " | ") + " |")

	if loggingLevel >= 1 {
		logInfo.Println("header printed.")
	}

	/* Emit markdown line that separates header from body table */

	var thisSeparator []string

	for _, thisColumn := range thisColumns {
		if thisColumn.isShown(config) {
			thisSeparator = append(thisSeparator, getThisAlignChar()[thisColumn.contentAlignMD])
		}
	}
	fmt.Println("| " + strings.Join(thisSeparator, " | ") + " |")

	if loggingLevel >= 1 {
		logInfo.Println("separator printed.")
	}

	/* Iterating through repos */

	for _, thisSshCertificateWithRevocation := range thisSshCertificatesWithRevocations {

		var thisRow []string

		/* Building slice of columns within a single row*/

		for _, thisColumn := range thisColumns {
			if thisColumn.isShown(config) {
				if thisColumn.contentEscapeMD {
					thisRow = append(thisRow, escapeMarkdown(thisColumn.contentSource(thisSshCertificateWithRevocation, config)))
				} else {
					thisRow = append(thisRow, thisColumn.contentSource(thisSshCertificateWithRevocation, config))
				}
			}
		}

		/* Emitting row */

		fmt.Println("| " + strings.Join(thisRow, " | ") + " |")
	}

	if loggingLevel >= 2 {
		logInfo.Printf("%d rows printed.\n", len(thisSshCertificatesWithRevocations))
	}
}
