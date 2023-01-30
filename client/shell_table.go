package client

import (
	"context"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	shellStyle = lipgloss.NewStyle().Padding(1, 2)
)

const divisor = 4

type item struct {
	name, kind string
}

func (s item) Title() string       { return s.name }
func (s item) Description() string { return s.kind }
func (s item) FilterValue() string { return s.name }

type model struct {
	list   list.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) View() string {
	return shellStyle.Render(m.list.View())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := shellStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	return m, cmd
}

func (c *Client) runShell(ctx context.Context) {
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

	m := model{list: list.New(items, list.NewDefaultDelegate(), 0, 0)}
	m.list.Title = "My Secrets"

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		c.log.Error().Err(err).Msg("Alas, there's been a shell error")
	}
	c.log.Info().Msg("shell shut down")
}
