package cmd

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/lukasz-lobocki/tabby"
)

/*
emitX509Table prints result in the form of a table.

	'thisX509CertsWithRevocations' Slice of structures describing the x509 certificates.
*/
func emitX509Table(thisX509CertsWithRevocations []tX509CertificateProvisionerRevocation) {

	table := new(tabby.Table)
	columns := getX509Columns()

	// Building slice of titles.
	var header []string
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

	if loggingLevel >= 1 { // Show info.
		logInfo.Println("header set.")
	}

	// Populate the table.
	for _, x509CertWithRevocation := range thisX509CertsWithRevocations {

		// Building slice of columns within a single row.
		var row []string
		for _, column := range columns {
			if column.isShown(config) {
				row = append(row,
					color.New(column.contentColor(x509CertWithRevocation)).SprintFunc()(
						column.contentSource(x509CertWithRevocation, config),
					),
				)
			}
		}

		if err := table.AppendRow(row); err != nil {
			logError.Panic(err)
		}
		if loggingLevel >= 3 { // Show info.
			logInfo.Printf("row [%s] appended.", x509CertWithRevocation.X509Certificate.SerialNumber.String())
		}

	}

	if loggingLevel >= 2 { // Show info.
		logInfo.Printf("%d rows appended.\n", len(thisX509CertsWithRevocations))
	}

	// Emit the table.
	if loggingLevel >= 3 { // Show spacing.
		table.Print(&tabby.Config{Spacing: "|", Padding: "."})
	} else {
		table.Print(nil)
	}
}

/*
emitX509CertsWithRevocationsJson prints result in the form of a json.

	'thisX509CertsWithRevocations' Slice of structures describing the x509 certificates.
*/
func emitX509CertsWithRevocationsJson(thisX509CertsWithRevocations []tX509CertificateProvisionerRevocation) {

	jsonInfo, err := json.MarshalIndent(thisX509CertsWithRevocations, "", "  ")
	if err != nil {
		logError.Panic(err)
	}

	fmt.Println(string(jsonInfo))

	if loggingLevel >= 2 { // Show info.
		logInfo.Printf("%d records marshalled.\n", len(thisX509CertsWithRevocations))
	}
}

/*
emitX509Markdown prints result in the form of markdown table.

	'thisX509CertsWithRevocations' Slice of certs.
*/
func emitX509Markdown(thisX509CertsWithRevocations []tX509CertificateProvisionerRevocation) {

	columns := getX509Columns()

	// Building slice of titles.
	var header []string
	for _, column := range columns {
		if column.isShown(config) {
			header = append(header, column.title())
		}
	}

	// Emitting titles.
	fmt.Println("| " + strings.Join(header, " | ") + " |")

	if loggingLevel >= 1 { // Show info.
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

	if loggingLevel >= 1 { // Show info.
		logInfo.Println("separator printed.")
	}

	// Iterating through certs.
	for _, x509CertWithRevocation := range thisX509CertsWithRevocations {

		// Building slice of columns within a single row.
		var row []string
		for _, column := range columns {
			if column.isShown(config) {
				if column.contentEscapeMD {
					row = append(row, escapeMarkdown(column.contentSource(x509CertWithRevocation, config)))
				} else {
					row = append(row, column.contentSource(x509CertWithRevocation, config))
				}
			}
		}

		// Emitting row.
		fmt.Println("| " + strings.Join(row, " | ") + " |")
	}

	if loggingLevel >= 2 { // Show info.
		logInfo.Printf("%d rows printed.\n", len(thisX509CertsWithRevocations))
	}
}

/*
emitOpenSsl prints result in the form of markdown table.

	'thisX509CertsWithRevocations' Slice of certs.
*/
func emitOpenSsl(thisX509CertsWithRevocations []tX509CertificateProvisionerRevocation) {
	for _, x509CertWithRevocation := range thisX509CertsWithRevocations {

		var revokedAt string

		// Construct RevokedAt string in compliance with specification.
		if len(x509CertWithRevocation.X509Revocation.ProvisionerID) > 0 {
			revokedAt = regexp.MustCompile(`[-T:]+`).
				ReplaceAllString(x509CertWithRevocation.X509Revocation.RevokedAt.UTC().
					Format(time.RFC3339), "")[2:]
		} else {
			revokedAt = ""
		}

		fmt.Printf("%s\t%s\t%s\t%039X\t%s\t%s\n",
			x509CertWithRevocation.Validity[0:1],
			regexp.MustCompile(`[-T:]+`).
				ReplaceAllString(x509CertWithRevocation.X509Certificate.NotAfter.UTC().
					Format(time.RFC3339), "")[2:], // Construct NotAfter string in compliance with specification.
			revokedAt,
			x509CertWithRevocation.X509Certificate.SerialNumber, // len(HEX)=39
			"unknown", // As per specification.
			x509CertWithRevocation.X509Certificate.Subject)
	}
}
