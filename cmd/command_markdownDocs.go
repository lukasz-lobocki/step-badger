/*
Copyright © 2024 Lukasz Lobocki
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// markdownDocsCmd represents the markdownDocs command
var markdownDocsCmd = &cobra.Command{
	Long: `
Generate markdown docs for the entire command tree.`,
	Short:                 "Generate markdown docs.",
	DisableFlagsInUseLine: true,
	Use: `markdownDocs <PATH> [flags]

Arguments:
  PATH   location for the result`,

	Aliases: []string{"markdowndocs"},

	Args: cobra.ExactArgs(1),

	Example: `  step-badger markdownDocs ~/tmp`,

	Hidden: true, // Hide command from the tree.

	Run: func(cmd *cobra.Command, args []string) {
		exportMarkdownMain(args)
	},
}

func init() {
	rootCmd.AddCommand(markdownDocsCmd)

	// Hide help command.
	markdownDocsCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	//Do not sort flags.
	markdownDocsCmd.Flags().SortFlags = false
}

func exportMarkdownMain(args []string) {

	checkLogginglevel(args)

	err := doc.GenMarkdownTree(rootCmd, args[0])
	if err != nil {
		logError.Fatalln(err)
	}
}
