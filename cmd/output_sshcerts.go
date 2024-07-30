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

	'thisSshCerts' Slice of structures describing the ssh certificates.
*/
func emitSshCertsTable(thisSshCerts []tSshCertificateWithRevocation) {
	table := new(tabby.Table)

	columns := getSshColumns()

	var header []string

	// Building slice of titles.
	for _, column := range columns {
		if column.isShown(config) {
			header = append(header,
				color.New(column.titleColor).SprintFunc()(
					column.title(),
				),
			)
		}
	}

	// Set the header.
	if err := table.SetHeader(header); err != nil {
		logError.Panic("Setting header failed. %w", err)
	}

	if loggingLevel >= 1 {
		logInfo.Println("header set.")
	}

	// Populate the table.
	for _, sshCert := range thisSshCerts {

		var row []string

		// Building slice of columns within a single row.
		for _, column := range columns {

			if column.isShown(config) {
				row = append(row,
					color.New(column.contentColor(sshCert)).SprintFunc()(
						column.contentSource(sshCert, config),
					),
				)
			}
		}

		if err := table.AppendRow(row); err != nil {
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
emitSshCertsJson prints result in the form of a json.

	'thisSshCerts' Slice of structures describing the ssh certificates.
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
emitX509Markdown prints result in the form of markdown table.

	'thisSshCerts' Slice of structures describing the certs.
*/
func emitSshCertsMarkdown(thisSshCertificatesWithRevocations []tSshCertificateWithRevocation) {
	columns := getSshColumns()

	var header []string

	// Building slice of titles.
	for _, column := range columns {
		if column.isShown(config) {
			header = append(header, column.title())
		}
	}

	// Emitting titles.
	fmt.Println("| " + strings.Join(header, " | ") + " |")

	if loggingLevel >= 1 {
		logInfo.Println("header printed.")
	}

	// Emit markdown line that separates header from body table.
	var separator []string

	for _, column := range columns {
		if column.isShown(config) {
			separator = append(separator, getAlignChar()[column.contentAlignMD])
		}
	}
	fmt.Println("| " + strings.Join(separator, " | ") + " |")

	if loggingLevel >= 1 {
		logInfo.Println("separator printed.")
	}

	// Iterating through certs.
	for _, sshCertificateWithRevocation := range thisSshCertificatesWithRevocations {

		var row []string

		// Building slice of columns within a single row.
		for _, column := range columns {
			if column.isShown(config) {
				if column.contentEscapeMD {
					row = append(row, escapeMarkdown(column.contentSource(sshCertificateWithRevocation, config)))
				} else {
					row = append(row, column.contentSource(sshCertificateWithRevocation, config))
				}
			}
		}

		// Emitting row.
		fmt.Println("| " + strings.Join(row, " | ") + " |")
	}

	if loggingLevel >= 2 {
		logInfo.Printf("%d rows printed.\n", len(thisSshCertificatesWithRevocations))
	}
}
