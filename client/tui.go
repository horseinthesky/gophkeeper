package client

import (
	"context"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
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

func (i item) Title() string       { return i.name }
func (i item) Description() string { return i.kind }
func (i item) FilterValue() string { return i.name }

type choiceItem struct {
	name string
}

func (c choiceItem) Title() string       { return c.name }
func (c choiceItem) Description() string { return c.name }
func (c choiceItem) FilterValue() string { return c.name }

type mode int

const (
	main mode = iota
	choice
	show
)

type model struct {
	mode     mode
	goph     *Client
	list     list.Model
	input    textinput.Model
	choices  list.Model
	viewport viewport.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) View() string {
	if m.input.Focused() {
		return shellStyle.Render(m.input.View())
	}

	if m.mode == choice {
		return shellStyle.Render(m.choices.View())
	}

	if m.mode == show {
		return shellStyle.Render(m.viewport.View())
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
			case key.Matches(msg, keyMap.Enter):
				data := strings.Fields(m.input.Value())
				kind, _ := strconv.ParseInt(data[0], 10, 64)
				m.goph.SetSecret(context.Background(), SecretKind(kind), data[1], []byte(data[2]))
				insCmd := m.list.InsertItem(0, item{name: data[1], kind: secretKindToString[SecretKind(kind)]})
				statusCmd := m.list.NewStatusMessage(statusMessageStyle("Added " + data[1]))
				m.input.Blur()
				m.input.SetValue("")
				return m, tea.Batch(insCmd, statusCmd)
			case key.Matches(msg, keyMap.Back):
				m.input.SetValue("")
				m.input.Blur()
			}
			m.input, cmd = m.input.Update(msg)
		} else if m.mode == choice {
			switch {
			case key.Matches(msg, keyMap.Back):
				m.mode = main
				return m, nil
			default:
				m.choices, cmd = m.choices.Update(msg)
				cmds = append(cmds, cmd)
			}
		} else if m.mode == show {
			switch {
			case key.Matches(msg, keyMap.Back):
				m.mode = main
				return m, nil
			}
		} else {
			// Don't match any of the keys below if we're actively filtering.
			if m.list.FilterState() == list.Filtering {
				break
			}

			switch {
			case key.Matches(msg, keyMap.Enter):
				i, ok := m.list.SelectedItem().(item)
				if !ok {
					return m, nil
				}

				secret, _ := m.goph.GetSecret(context.Background(), stringToSecretKind[i.kind], i.name)
				m.viewport = viewport.New(30, 5)
				// m.viewport = viewport.New(shellStyle.GetFrameSize())
				m.viewport.SetContent(string(secret.Value))
				m.mode = show
				return m, nil
			case key.Matches(msg, keyMap.Create):
				m.mode = choice
				return m, nil
				// m.input.Focus()
				// cmd = textinput.Blink
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

				m.goph.DeleteSecret(context.Background(), stringToSecretKind[i.kind], i.name)
				statusCmd := m.list.NewStatusMessage(statusMessageStyle("Deleted " + i.Title()))
				return m, tea.Batch(statusCmd)
			}
		}
	}

	// List update must be here for filtering to work
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
		items = append(
			items,
			item{
				name: secret.Name,
				kind: secretKindToString[SecretKind(secret.Kind)],
			},
		)
	}

	// Init choice model
	choices := []list.Item{}
	for _, secretKindString := range secretKindToString {
		choices = append(choices, choiceItem{name: secretKindString})
	}

	// Setup TUI
	m := model{
		goph:    c,
		list:    list.New(items, list.NewDefaultDelegate(), 0, 0),
		choices: list.New(choices, list.NewDefaultDelegate(), 30, 18),
		input:   input,
	}
	m.list.Title = "My Secrets"
	m.list.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			keyMap.Create,
			keyMap.Delete,
		}
	}
	m.choices.Title = "Choose new secret type"
	m.choices.SetFilteringEnabled(false)
	m.choices.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			keyMap.Back,
		}
	}

	// Run TUI
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		c.log.Error().Err(err).Msg("Alas, there's been a shell error")
	}
	c.log.Info().Msg("shell shut down")
}
