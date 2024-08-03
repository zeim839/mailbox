package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/zeim839/mailbox/cmd/table"
	"io"
	"net/http"
	"os"
)

type browseCmdModel struct {
	table table.Model
}

func (m browseCmdModel) Init() tea.Cmd { return nil }

func (m browseCmdModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m browseCmdModel) View() string {
	return m.table.View() + "\n"
}

func init() {
	rootCmd.AddCommand(browseCmd)
}

func fetchTablePage(page int64) ([]table.Row, int64) {

	// Create new request.
	url := fmt.Sprintf("%s/entries/?page=%d", api, page)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("server failed to respond", url)
		return []table.Row{}, -1
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
		fmt.Println("server failed to respond:", url)
		return []table.Row{}, -1
	}
	defer resp.Body.Close()

	// Read the response body.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []table.Row{}, -1
	}

	if resp.StatusCode == 200 {
		var responseData readAllResponse
		if err := json.Unmarshal(body, &responseData); err != nil {
			return []table.Row{}, -1
		}
		rows := []table.Row{}
		for _, val := range responseData.Entries {
			rows = append(rows, table.Row{
				val.ID,
				val.From,
				val.Subject,
				val.Message,
			})
		}
		next := int64(-1)
		if responseData.Page < responseData.PageCount-1 {
			next = responseData.Page + 1
		}
		return rows, next
	}

	// Error message.
	var responseData commonResponse
	json.Unmarshal(body, &responseData)
	fmt.Println("Server error:", resp.Status)
	if responseData.Error != "" {
		fmt.Println(responseData.Error)
	}
	os.Exit(1)
	return []table.Row{}, -1
}

func fetchTableData() []table.Row {
	next := int64(0)
	rows := []table.Row{}
	for next > -1 {
		r, page := fetchTablePage(next)
		rows = append(rows, r...)
		next = page
	}
	return rows
}

func deleteTableRows(rows []table.Row) error {
	for _, row := range rows {

		// Create a new request.
		url := fmt.Sprintf("%s/entry/%s", api, row[0])
		req, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			return fmt.Errorf("error creating request: %s", err.Error())
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
			return fmt.Errorf("error making request: %s", err.Error())
		}
		defer resp.Body.Close()

		// Read the response body.
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error reading response: %s", err.Error())
		}

		if resp.StatusCode == 200 {
			continue
		}

		// Error message.
		var responseData commonResponse
		json.Unmarshal(body, &responseData)
		fmt.Println("Server error:", resp.Status)
		if responseData.Error != "" {
			return fmt.Errorf("error reading response: %s", err.Error())
		}
	}
	return nil
}

var browseCmd = &cobra.Command{
	Use:   "browse",
	Short: "Browse contact form submissions",
	Run: func(cmd *cobra.Command, args []string) {
		validateAPI()
		columns := []table.Column{
			{Title: "ID", Width: 5},
			{Title: "From", Width: 15},
			{Title: "Subject", Width: 15},
			{Title: "Message", Width: 30},
		}

		rows := fetchTableData()
		if len(rows) == 0 {
			fmt.Println("No submissions found")
			os.Exit(1)
		}

		t := table.New(
			table.WithColumns(columns),
			table.WithRows(rows),
			table.WithFocused(true),
			table.WithHeight(10),
			table.WithRefreshFn(fetchTableData),
			table.WithDeleteFn(deleteTableRows),
		)

		s := table.DefaultStyles()
		s.Header = s.Header.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			BorderBottom(true).
			Bold(false)
		s.Selected = s.Selected.
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Bold(false)
		t.SetStyles(s)

		m := browseCmdModel{t}
		if _, err := tea.NewProgram(m).Run(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	},
}
