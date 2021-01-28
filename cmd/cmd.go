package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var processAttachments, useGitDates, verbose bool
var toAppend string

func init() {
	RootCmd.PersistentFlags().BoolVarP(
		&processAttachments,
		"process-attachments",
		"p",
		false,
		"Replace links to local files with Bear-compatible tags to ease processing",
	)
	RootCmd.PersistentFlags().BoolVarP(
		&useGitDates,
		"git-dates",
		"g",
		false,
		"Instead of using OS creation / modification dates of Markdown file, use the dates from git commit history (must be in a git repo & have git CLI)",
	)
	RootCmd.PersistentFlags().StringVarP(
		&toAppend,
		"append",
		"a",
		"",
		"Text to append to end of Markdown file. Use %f to template the original filename.",
	)
	RootCmd.PersistentFlags().BoolVarP(
		&verbose,
		"verbose",
		"v",
		false,
		"Verbose",
	)
}

// RootCmd handles the base case for textbundler: processing Markdown files.
var RootCmd = &cobra.Command{
	Use:   "textbundler [file] [file2] [file3]...",
	Short: "Convert markdown files into textbundles",
	Run: func(md *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Fprintln(os.Stderr, "Please pass at least one argument.")
			os.Exit(1)
		}

		for _, file := range args {
			if err := process(file); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
	},
}

func process(file string) error {

	return nil
}

// Execute begins the CLI processing flow
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
