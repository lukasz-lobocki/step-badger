package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/lukasz-lobocki/tabby"
)

/*
emitTable prints rows in the form of a table.

	'rows' slice of structures describing the records.
	'columns' the column definitions to render.
	'rowLabel' returns a human-readable label for a row, used in logging.
*/
func emitTable[T any](rows []T, columns []tColumn[T], rowLabel func(T) string) {

	table := new(tabby.Table)

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
		logError.Panicf("setting header failed: %v", err)
	}

	if loggingLevel >= 1 { // Show info.
		logInfo.Println("header set.")
	}

	// Populate the table.
	for _, row := range rows {

		// Building slice of columns within a single row.
		var cells []string
		for _, column := range columns {
			if column.isShown(config) {
				cells = append(cells,
					color.New(column.contentColor(row)).SprintFunc()(
						column.contentSource(row, config),
					),
				)
			}
		}

		if err := table.AppendRow(cells); err != nil {
			logError.Panic(err)
		}
		if loggingLevel >= 2 { // Show info.
			logInfo.Printf("row [%s] appended.", rowLabel(row))
		}

	}

	if loggingLevel >= 2 { // Show info.
		logInfo.Printf("%d rows appended.\n", len(rows))
	}

	// Emit the table.
	if loggingLevel >= 3 { // Show spacing.
		table.Print(&tabby.Config{Spacing: "|", Padding: "."})
	} else {
		table.Print(nil)
	}
}

/*
emitJson prints rows in the form of a json.

	'rows' slice of structures describing the records.
*/
func emitJson[T any](rows []T) {

	jsonInfo, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		logError.Panic(err)
	}

	fmt.Println(string(jsonInfo))

	if loggingLevel >= 2 { // Show info.
		logInfo.Printf("%d records marshalled.\n", len(rows))
	}
}

/*
emitPlain prints rows in the plain form.

	'rows' slice of structures describing the records.
	'columns' the column definitions to render.
*/
func emitPlain[T any](rows []T, columns []tColumn[T]) {

	// Building slice of titles.
	var header []string
	for _, column := range columns {
		if column.isShown(config) {
			header = append(header, column.title())
		}
	}

	// Emitting titles.
	fmt.Println(strings.Join(header, "\t"))

	if loggingLevel >= 1 { // Show info.
		logInfo.Println("header printed.")
	}

	// Iterating through rows.
	for _, row := range rows {

		// Building slice of columns within a single row.
		var cells []string
		for _, column := range columns {
			if column.isShown(config) {
				cells = append(cells, column.contentSource(row, config))
			}
		}

		// Emitting row.
		fmt.Println(strings.Join(cells, "\t"))
	}

	if loggingLevel >= 2 { // Show info.
		logInfo.Printf("%d rows printed.\n", len(rows))
	}
}

/*
emitMarkdown prints rows in the form of a markdown table.

	'rows' slice of structures describing the records.
	'columns' the column definitions to render.
*/
func emitMarkdown[T any](rows []T, columns []tColumn[T]) {

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
			separator = append(separator, alignChars[column.contentAlignMD])
		}
	}
	fmt.Println("| " + strings.Join(separator, " | ") + " |")

	if loggingLevel >= 1 { // Show info.
		logInfo.Println("separator printed.")
	}

	// Iterating through rows.
	for _, row := range rows {

		// Building slice of columns within a single row.
		var cells []string
		for _, column := range columns {
			if column.isShown(config) {
				if column.contentEscapeMD {
					cells = append(cells, escapeMarkdown(column.contentSource(row, config)))
				} else {
					cells = append(cells, column.contentSource(row, config))
				}
			}
		}

		// Emitting row.
		fmt.Println("| " + strings.Join(cells, " | ") + " |")
	}

	if loggingLevel >= 2 { // Show info.
		logInfo.Printf("%d rows printed.\n", len(rows))
	}
}
