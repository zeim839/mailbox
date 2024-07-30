package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/zeim839/mailbox/data"
	"net/url"
	"os"
	"regexp"
)

var (
	rootCmd = &cobra.Command{
		Use:   "mbx",
		Short: "Mailbox is a website contact form manager",
		Long: `Mailbox is a CLI tool for managing website contact forms.
It interacts with a remote API server to fetch and modify
submissions. Full documentation is available at:
<https://github.com/zeim839/mailbox>`,
	}

	api string
	usr string
	pwd string
)

type commonResponse struct {
	Error string `json:"error"`
}

type readAllResponse struct {
	Page       int64       `json:"page" binding:"required"`
	PageCount  int64       `json:"page_count" binding:"required"`
	EntryCount int64       `json:"entry_count" binding:"required"`
	Entries    []data.Form `json:"entries" binding:"required"`
}

func init() {
	rootCmd.PersistentFlags().StringVar(&api, "api", "", "(Required) HTTP API endpoint")
	rootCmd.PersistentFlags().StringVar(&usr, "usr", "", "(Optional) Your basic auth username")
	rootCmd.PersistentFlags().StringVar(&pwd, "pwd", "", "(Optional) Your basic auth password")
}

func validateAPI() {
	parsedURL, err := url.Parse(api)
	if err != nil {
		fmt.Println("Error: invalid argument for \"--api\" flag (required)")
		fmt.Println(fmt.Errorf("invalid API endpoint: %v", err))
		os.Exit(1)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		fmt.Println("Error: invalid argument for \"--api\" flag (required)")
		fmt.Println(fmt.Errorf("invalid scheme: must be http or https"))
		os.Exit(1)
	}

	re := regexp.MustCompile(`([^:]/)/+`)

	// Replace repeated slashes with a single slash
	correctedPath := re.ReplaceAllString(parsedURL.Path, "$1")

	// Remove trailing slash if present
	if len(correctedPath) > 0 && correctedPath[len(correctedPath)-1] == '/' {
		correctedPath = correctedPath[:len(correctedPath)-1]
	}

	parsedURL.Path = correctedPath
	api = parsedURL.String()
}

// Execute the cobra command line interface.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}
