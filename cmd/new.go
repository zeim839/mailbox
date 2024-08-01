package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"os"
	"strings"
)

var (
	src          string = ""
	sub          string = ""
	bod          string = ""
	enterPressed        = false

	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle  = focusedStyle
	noStyle      = lipgloss.NewStyle()
	helpStyle    = blurredStyle

	focusedSubmitButton = focusedStyle.Render("[ Submit ]")
	blurredSubmitButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
	focusedCancelButton = focusedStyle.Render("[ Cancel ]")
	blurredCancelButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Cancel"))
)

type newCmdModel struct {
	focusIndex int
	inputs     []textinput.Model
}

func initialNewModel() newCmdModel {
	m := newCmdModel{
		inputs: make([]textinput.Model, 3),
	}
	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 32

		switch i {
		case 0:
			t.Placeholder = "user@example.com"
			t.Prompt = "From (email): "
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
			t.SetValue(src)
		case 1:
			t.Placeholder = "My subject"
			t.Prompt = "Subject:      "
			t.CharLimit = 64
			t.SetValue(sub)
		case 2:
			t.Placeholder = "Message"
			t.Prompt = "Message:      "
			t.CharLimit = 10000
			t.SetValue(bod)
		}

		m.inputs[i] = t
	}

	return m
}

func (m newCmdModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m newCmdModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "ctrl+p":
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs) + 1
			}
			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					// Set focused state
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = focusedStyle
					m.inputs[i].TextStyle = focusedStyle
					continue
				}
				// Remove focused state
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].TextStyle = noStyle
			}
			return m, tea.Batch(cmds...)

		// Set focus to next input
		case "tab", "shift+tab", "enter", "up", "down", "ctrl+n":
			s := msg.String()

			// Did the user press enter while the cancel button was focused?
			// If so, exit.
			if s == "enter" && m.focusIndex == len(m.inputs) {
				return m, tea.Quit
			}

			// Did the user press enter while the submit button was focused?
			// If so, attempt to submit message.
			if s == "enter" && m.focusIndex == len(m.inputs)+1 {
				enterPressed = true
				return m, tea.Quit
			}

			// Cycle indexes
			if s == "up" || s == "ctrl+p" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs)+1 {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					// Set focused state
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = focusedStyle
					m.inputs[i].TextStyle = focusedStyle
					continue
				}
				// Remove focused state
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)
		}
	}

	// Handle character input and blinking
	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m *newCmdModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
		switch i {
		case 0:
			src = m.inputs[i].Value()
		case 1:
			sub = m.inputs[i].Value()
		case 2:
			bod = m.inputs[i].Value()
		}
	}

	return tea.Batch(cmds...)
}

func (m newCmdModel) View() string {
	var b strings.Builder

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	submitButton := &blurredSubmitButton
	cancelButton := &blurredCancelButton
	if m.focusIndex == len(m.inputs) {
		cancelButton = &focusedCancelButton
	}
	if m.focusIndex == len(m.inputs)+1 {
		submitButton = &focusedSubmitButton
	}
	fmt.Fprintf(&b, "\n\n%s %s\n\n", *cancelButton, *submitButton)
	b.WriteString(helpStyle.Render("(enter & arrow keys, ctrl+c to exit)"))

	return b.String()
}

func init() {
	newCmd.Flags().StringVarP(&src, "from", "f", "", "The source email address")
	newCmd.Flags().StringVarP(&sub, "sub", "s", "", "The message subject")
	newCmd.Flags().StringVarP(&bod, "msg", "m", "", "The message body")
	rootCmd.AddCommand(newCmd)
}

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Submit a new message to the contact form",
	Run: func(cmd *cobra.Command, args []string) {
		validateAPI()

		// Submit without launching TUI.
		if src != "" && sub != "" && bod != "" {
			submit()
			return
		}

		submitTUI()
	},
}

func submit() {

	// JSON payload
	payload := map[string]string{
		"from":    src,
		"subject": sub,
		"message": bod,
	}

	// Convert payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		os.Exit(1)
	}

	// Create a new request
	req, err := http.NewRequest("POST", api+"/submit", bytes.NewBuffer(jsonPayload))
	if err != nil {
		fmt.Println("Error creating request:", err)
		os.Exit(1)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if usr != "" && pwd != "" {
		basicAuth := base64.StdEncoding.EncodeToString(
			[]byte(usr + ":" + pwd))

		req.Header.Add("Authorization", "Basic "+basicAuth)
	}

	// Create HTTP client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		os.Exit(1)
	}

	// Success.
	if resp.StatusCode == 200 {
		fmt.Println("Response Status:", resp.Status)
		fmt.Println("(success)")
		os.Exit(0)
	}

	// Error message.
	var responseData commonResponse
	json.Unmarshal(body, &responseData)
	fmt.Println("Server error:", resp.Status)
	if responseData.Error != "" {
		fmt.Println(responseData.Error)
		os.Exit(1)
	}
}

func submitTUI() {
	if _, err := tea.NewProgram(initialNewModel()).Run(); err != nil {
		fmt.Printf("could not start program: %s\n", err)
		os.Exit(1)
	}
	if enterPressed {
		submit()
	}
	os.Exit(0)
}
