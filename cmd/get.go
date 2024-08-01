package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"os"
)

func init() {
	rootCmd.AddCommand(getCmd)
}

var getCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Fetch a contact form submission",
	Long: `Fetch a contact form submission by its ID. If the document
is found, it is returned to stdout as JSON. Otherwise, a
JSON error is returned.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		validateAPI()

		// Create a new request.
		url := fmt.Sprintf("%s/entry/%s", api, args[0])
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Println("Error creating request:", err)
			os.Exit(1)
		}

		// Add basic authentication (if applicable).
		if usr != "" && pwd != "" {
			basicAuth := base64.StdEncoding.EncodeToString(
				[]byte(usr + ":" + pwd))

			req.Header.Add("Authorization", "Basic "+basicAuth)
		}

		// Create a client and send the request.
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error making the request:", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		// Read the response body.
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading the response:", err)
			os.Exit(1)
		}

		if resp.StatusCode == 200 {
			fmt.Println(string(body))
			os.Exit(0)
		}

		// Error message.
		var responseData commonResponse
		json.Unmarshal(body, &responseData)
		fmt.Println("Server error:", resp.Status)
		if responseData.Error != "" {
			fmt.Println(responseData.Error)
		}
		os.Exit(1)
	},
}
