package ghe

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var ghe_host string
type tickMsg struct{}
type errMsg error

type model struct {
	textInput textinput.Model
	err       error
}

func initialModel() model {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return model{
		textInput: ti,
		err:       nil,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
				case tea.KeyCtrlC, tea.KeyEsc:
					os.Exit(0)

					return m, tea.Quit

				case tea.KeyEnter:
					hostname := m.textInput.Value()

					if len(strings.TrimSpace(hostname)) < 1 {
						m.textInput.Placeholder = "a value is required"
						m.textInput.SetValue("")
					} else if strings.ContainsRune(hostname, '/') || strings.ContainsRune(hostname, ':') {
						m.textInput.Placeholder = "invalid hostname"
						m.textInput.SetValue("")
					} else {
						ghe_host = hostname
						return m, tea.Quit
					}
			}

		case errMsg:
			m.err = msg

			return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)

	return m, cmd
}

func (m model) View() string {
	ghe_hostname := lipgloss.NewStyle().Bold(true).SetString("GHE hostname:").String()

	return fmt.Sprintf(
		"%s%s\n",
		ghe_hostname,
		m.textInput.View(),
	) + "\n"
}

func GHE() (string, error) {
	err := tea.NewProgram(initialModel()).Start()

	return ghe_host, err
}
