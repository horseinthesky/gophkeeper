package client

import (
	"context"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	shellStyle         = lipgloss.NewStyle().Padding(1, 2)
	docStyle           = lipgloss.NewStyle().Margin(1, 2)
	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
				Render
)

type item struct {
	name, kind string
}

func (s item) Title() string       { return s.name }
func (s item) Description() string { return s.kind }
func (s item) FilterValue() string { return s.name }

type model struct {
	goph  *Client
	list  list.Model
	input textinput.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) View() string {
	if m.input.Focused() {
		return shellStyle.Render(m.input.View())
	}

	return shellStyle.Render(m.list.View())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := shellStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	case tea.KeyMsg:
		if m.input.Focused() {
			switch {
			case key.Matches(msg, keyMap.Quit):
				return m, tea.Quit
			case key.Matches(msg, keyMap.Enter):
				data := strings.Fields(m.input.Value())
				kind, _ := strconv.ParseInt(data[0], 10, 64)
				m.goph.SetSecret(context.Background(), SecretKind(kind), data[1], []byte(data[2]))
				insCmd := m.list.InsertItem(0, item{name: data[1], kind: data[1]})
				statusCmd := m.list.NewStatusMessage(statusMessageStyle("Added " + data[1]))
				m.input.Blur()
				m.input.SetValue("")
				return m, tea.Batch(insCmd, statusCmd)
			case key.Matches(msg, keyMap.Back):
				m.input.SetValue("")
				m.input.Blur()
			}
			m.input, cmd = m.input.Update(msg)
		} else {
			// Don't match any of the keys below if we're actively filtering.
			if m.list.FilterState() == list.Filtering {
				break
			}

			switch {
			case key.Matches(msg, keyMap.Quit):
				return m, tea.Quit
			case key.Matches(msg, keyMap.Create):
				m.input.Focus()
				cmd = textinput.Blink
				// case key.Matches(msg, keyMap.Enter):
				// 	activeProject := m.list.SelectedItem().(project.Project)
				// 	entry := InitEntry(constants.Er, activeProject.ID, constants.P)
				// 	return entry.Update(constants.WindowSize)
			case key.Matches(msg, keyMap.Delete):
				i, ok := m.list.SelectedItem().(item)
				if !ok {
					return m, nil
				}

				index := m.list.Index()
				m.list.RemoveItem(index)
				if len(m.list.Items()) == 0 {
					keyMap.Delete.SetEnabled(false)
				}

				m.goph.DeleteSecret(context.Background(), toID[i.kind], i.name)
				statusCmd := m.list.NewStatusMessage(statusMessageStyle("Deleted " + i.Title()))
				return m, tea.Batch(statusCmd)
			}
		}
	}
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (c *Client) runShell(ctx context.Context) {
	// Init input model
	input := textinput.New()
	input.Prompt = "enter new secret> "
	input.Placeholder = "my supersecret secret"
	input.CharLimit = 250
	input.Width = 50

	// Init list model
	items := []list.Item{}

	secrets, err := c.ListSecrets(ctx)
	if err != nil {
		c.log.Error().Err(err).Msg("failed to list user '%s' secrets")
		return
	}

	for _, secret := range secrets {
		if secret.Deleted.Bool {
			continue
		}
		items = append(items, item{name: secret.Name.String, kind: toString[SecretKind(secret.Kind.Int32)]})
	}

	m := model{goph: c, list: list.New(items, list.NewDefaultDelegate(), 0, 0), input: input}
	m.list.Title = "My Secrets"
	m.list.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			keyMap.Create,
			keyMap.Delete,
		}
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		c.log.Error().Err(err).Msg("Alas, there's been a shell error")
	}
	c.log.Info().Msg("shell shut down")
}
