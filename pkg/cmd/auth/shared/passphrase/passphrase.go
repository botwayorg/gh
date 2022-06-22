package passphrase

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var pass string
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
					pass = m.textInput.Value()

					return m, tea.Quit
			}

		case errMsg:
			m.err = msg

			return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)

	return m, cmd
}

func (m model) View() string {
	passname := lipgloss.NewStyle().Bold(true).SetString("Enter a passphrase for your new SSH key (Optional):").String()

	return fmt.Sprintf(
		"%s%s\n",
		passname,
		m.textInput.View(),
	) + "\n"
}

func Passphrase() (string, error) {
	err := tea.NewProgram(initialModel()).Start()

	return pass, err
}
