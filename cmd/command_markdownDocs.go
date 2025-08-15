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
	Short:   "Generate markdown docs.",
	Run:     func(cmd *cobra.Command, args []string) { exportMarkdownMain(args) },
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"markdowndocs"},
	Hidden:  true, // Hide command from the tree.

	DisableFlagsInUseLine: true,

	Example: `  step-badger markdownDocs ~/tmp`,
	Long: `
Generate markdown docs for the entire command tree.`,
	Use: `markdownDocs <PATH> [flags]

Arguments:
  PATH   location for the result`,
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
