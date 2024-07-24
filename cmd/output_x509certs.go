package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/lukasz-lobocki/tabby"
)

/*
emitX509Table prints result in the form of a table

	'thisX509CertsWithRevocations' slice of structures describing the x509 certificates
*/
func emitX509Table(thisX509CertsWithRevocations []tX509CertificateAndRevocation) {
	table := new(tabby.Table)

	thisColumns := getX509Columns()

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

	for _, x509CertAndRevocation := range thisX509CertsWithRevocations {

		var thisRow []string
		/* Building slice of columns within a single row*/
		for _, thisColumn := range thisColumns {

			if thisColumn.isShown(config) {
				thisRow = append(thisRow,
					color.New(thisColumn.contentColor(x509CertAndRevocation)).SprintFunc()(
						thisColumn.contentSource(x509CertAndRevocation, config),
					),
				)
			}
		}

		if err := table.AppendRow(thisRow); err != nil {
			logError.Panic(err)
		}
		if loggingLevel >= 3 {
			logInfo.Printf("row [%s] appended.", x509CertAndRevocation.X509Certificate.SerialNumber.String())
		}

	}

	if loggingLevel >= 2 {
		logInfo.Printf("%d rows appended.\n", len(thisX509CertsWithRevocations))
	}

	/* Emit the table */

	if loggingLevel >= 3 {
		table.Print(&tabby.Config{Spacing: "|", Padding: "."})
	} else {
		table.Print(nil)
	}
}

/*
emitX509CertsWithRevocationsJson prints result in the form of a json

	'thisX509CertsWithRevocations' slice of structures describing the x509 certificates
*/
func emitX509CertsWithRevocationsJson(thisX509CertsWithRevocations []tX509CertificateAndRevocation) {
	jsonInfo, err := json.MarshalIndent(thisX509CertsWithRevocations, "", "  ")
	if err != nil {
		logError.Panic(err)
	}
	fmt.Println(string(jsonInfo))
	if loggingLevel >= 2 {
		logInfo.Printf("%d records marshalled.\n", len(thisX509CertsWithRevocations))
	}
}

/*
emitX509Markdown prints result in the form of markdown table

	'thisX509CertsWithRevocations' slice of structures certs
*/
func emitX509Markdown(thisX509CertsAndRevocations []tX509CertificateAndRevocation) {
	thisColumns := getX509Columns()

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

	for _, thisX509CertsAndRevocation := range thisX509CertsAndRevocations {

		var thisRow []string

		/* Building slice of columns within a single row*/

		for _, thisColumn := range thisColumns {
			if thisColumn.isShown(config) {
				if thisColumn.contentEscapeMD {
					thisRow = append(thisRow, escapeMarkdown(thisColumn.contentSource(thisX509CertsAndRevocation, config)))
				} else {
					thisRow = append(thisRow, thisColumn.contentSource(thisX509CertsAndRevocation, config))
				}
			}
		}

		/* Emitting row */

		fmt.Println("| " + strings.Join(thisRow, " | ") + " |")
	}

	if loggingLevel >= 2 {
		logInfo.Printf("%d rows printed.\n", len(thisX509CertsAndRevocations))
	}
}
