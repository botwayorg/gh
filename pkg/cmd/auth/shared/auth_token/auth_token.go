package auth_token

import (
	"fmt"
	"os"
	"reflect"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var gh_token string
type tickMsg struct{}
type errMsg error

type model struct {
	textInput textinput.Model
	err       error
}

func initialModel() model {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 512
	ti.Width = 20
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '*'

	return model{
		textInput: ti,
		err:       nil,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Slice, reflect.Map:
		return v.Len() == 0
	}

	// compare the types directly with more general coverage
	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
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
					token := m.textInput.Value()
					value := reflect.ValueOf(token)

					// if the value passed in is the zero value of the appropriate type
					if isZero(value) && value.Kind() != reflect.Bool {
						m.textInput.Placeholder = "Value is required"
					} else {
						gh_token = token
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
	view := lipgloss.NewStyle().Bold(true).SetString("Paste your authentication token:").String()

	return fmt.Sprintf(
		"%s%s\n",
		view,
		m.textInput.View(),
	) + "\n"
}

func AuthToken() (string, error) {
	err := tea.NewProgram(initialModel()).Start()

	return gh_token, err
}
