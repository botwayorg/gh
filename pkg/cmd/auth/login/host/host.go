package host

import (
	"fmt"
	"io"
	"os"

	"github.com/botwayorg/gh/core/ghinstance"
	"github.com/botwayorg/gh/pkg/cmd/auth/login/host/ghe"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	listHeight = 4
	titleStyle        = lipgloss.NewStyle().Bold(true)
	questionStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#04B575"))
	itemStyle         = lipgloss.NewStyle().PaddingLeft(2)
	selectedItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#1E90FF"))
	paginationStyle   = list.DefaultStyles().PaginationStyle
)

type item string
var host string
type itemDelegate struct{}

func (i item) FilterValue() string { return "" }
func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)

	if !ok {
		return
	}

	str := fmt.Sprintf("%s", i)

	fn := itemStyle.Render

	if index == m.Index() {
		fn = func(s string) string {
			return selectedItemStyle.Render("> " + s)
		}
	}

	fmt.Fprintf(w, fn(str))
}

type model struct {
	list     list.Model
	items    []item
	choice   string
	quitting bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.list.SetWidth(msg.Width)

			return m, nil

		case tea.KeyMsg:
			switch keypress := msg.String(); keypress {
				case "ctrl+c":
					m.quitting = true

					os.Exit(0)

					return m, tea.Quit

				case "enter":
					i, ok := m.list.SelectedItem().(item)

					if ok {
						m.choice = string(i)
					}

					return m, tea.Quit
			}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	return m, cmd
}

func (m model) View() string {
	if m.choice != "" {
		host = m.choice
		return ""
	}

	return questionStyle.Render("?") + titleStyle.Render(" What account do you want to log into?") + "\n" + m.list.View()
}

func Host() (string, error) {
	items := []list.Item{
		item("GitHub.com"),
		item("GitHub Enterprise Server"),
	}

	l := list.New(items, itemDelegate{}, 20, listHeight)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.SetShowHelp(false)

	m := model{list: l}

	err := tea.NewProgram(m).Start()

	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	isEnterprise := host == "GitHub Enterprise Server"

	hostname := ghinstance.Default()

	if isEnterprise {
		var err error

		hostname, err = ghe.GHE()

		if err != nil {
			return "", fmt.Errorf("could not prompt: %w", err)
		}
	}

	fmt.Println(lipgloss.NewStyle().Bold(true).SetString("Host: ").String() + hostname)

	return hostname, err
}
