package client

import (
	"context"
	"encoding/json"
	"fmt"
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

type choiceItem string

func (c choiceItem) Title() string       { return string(c) }
func (c choiceItem) Description() string { return "" }
func (c choiceItem) FilterValue() string { return "" }

type mode int

const (
	main mode = iota
	choice
	entry
	show
)

type model struct {
	mode mode
	goph *Client

	list    list.Model // Main menu
	choices list.Model // New secret kinds menu

	inputs     []textinput.Model // New secret params input
	focusIndex int               // Index for new secret param

	selectedSecretKind SecretKind      // Selected secret kind for new secret
	viewport           viewport.Model  // Display secret info
	secretBytesContent []byte          // Content of bytes secret - file content
	input              textinput.Model // File path to save bytes secret content on disk
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

	if m.mode == entry {
		var b strings.Builder

		for i := range m.inputs {
			b.WriteString(m.inputs[i].View())
			if i < len(m.inputs)-1 {
				b.WriteRune('\n')
			}
		}

		button := &blurredButton
		if m.focusIndex == len(m.inputs) {
			button = &focusedButton
		}
		fmt.Fprintf(&b, "\n\n%s\n\n", *button)

		return b.String()
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
		m.choices.SetSize(msg.Width-h, msg.Height-v)
	case tea.KeyMsg:
		if m.input.Focused() {
			switch {
			case key.Matches(msg, keyMap.Enter):
				err := saveOnDisk(m.input.Value(), m.secretBytesContent)
				if err != nil {
					m.input.SetValue("")
					m.input.Placeholder = fmt.Sprintf("invalid file path: %s", err.Error())
					return m, nil
				}

				m.input.SetValue("")
				m.input.Blur()
			case key.Matches(msg, keyMap.Back):
				m.input.SetValue("")
				m.input.Blur()
			}

			// Only log keypresses for the input field when it's focused
			m.input, cmd = m.input.Update(msg)
			cmds = append(cmds, cmd)

			return m, tea.Batch(cmds...)
		}

		switch m.mode {
		case choice:
			switch {
			case key.Matches(msg, keyMap.Back):
				m.mode = main
				return m, nil
			case key.Matches(msg, keyMap.Enter):
				i, _ := m.choices.SelectedItem().(choiceItem)
				kind := stringToSecretKind[string(i)]
				m.inputs = entryMap[kind]()
				m.selectedSecretKind = kind
				m.mode = entry
				return m, nil
			default:
				m.choices, cmd = m.choices.Update(msg)
				cmds = append(cmds, cmd)
			}
		case entry:
			switch msg.String() {
			// Go back to secrets list
			case "esc":
				m.mode = main
				return m, nil
			// Set focus to next input
			case "tab", "shift+tab", "enter", "up", "down":
				s := msg.String()

				// Did the user press enter while the submit button was focused?
				// If so, exit.
				if s == "enter" && m.focusIndex == len(m.inputs) {
					dbSecret, err := m.goph.storeSecretFromEntry(m.selectedSecretKind, m.inputs)
					if err != nil {
						m.goph.log.Error().Err(err).Msgf("failed to save secret %s", m.inputs[0].Value())
						return m, nil
					}

					m.mode = main
					insCmd := m.list.InsertItem(0, item{name: dbSecret.Name, kind: secretKindToString[m.selectedSecretKind]})
					statusCmd := m.list.NewStatusMessage(statusMessageStyle("Added " + m.inputs[0].Value()))
					return m, tea.Batch(insCmd, statusCmd)
				}

				// Cycle indexes
				if s == "up" || s == "shift+tab" {
					m.focusIndex--
				} else {
					m.focusIndex++
				}

				if m.focusIndex > len(m.inputs) {
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
			cmd := m.updateInputs(msg)
			return m, cmd
		case show:
			switch {
			case key.Matches(msg, keyMap.Back):
				m.mode = main
				return m, nil
			case key.Matches(msg, keyMap.Save):
				m.goph.log.Info().Msg(secretKindToString[m.selectedSecretKind])
				if m.selectedSecretKind == SecretBytes {
					m.input.Focus()
					return m, nil
				}
			}
		default:
			// Don't match any of the keys below if we're actively filtering.
			if m.list.FilterState() == list.Filtering {
				break
			}

			switch {
			case key.Matches(msg, keyMap.Enter):
				i, _ := m.list.SelectedItem().(item)

				m.viewport = viewport.New(200, 10)

				// Load secret from DB. Decrypt if needed
				dbSecret, err := m.goph.GetSecret(stringToSecretKind[i.kind], i.name)
				if err != nil {
					m.viewport.SetContent(err.Error())
					m.mode = show
					return m, nil
				}

				// Save content of bytes secret for the user decides to save to disk
				if dbSecret.Kind == int32(SecretBytes) {
					var payload BytesPayload
					json.Unmarshal(dbSecret.Value, &payload)
					m.secretBytesContent = payload.Bytes
				}

				// Load secret display data
				secretContent, err := m.goph.loadSecretContentFromEntry(dbSecret)
				if err != nil {
					m.viewport.SetContent(err.Error())
				} else {
					m.viewport.SetContent(secretContent)
				}
				m.selectedSecretKind = SecretKind(dbSecret.Kind)
				m.mode = show
				return m, nil
			case key.Matches(msg, keyMap.Create):
				m.mode = choice
				return m, nil
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

				m.goph.DeleteSecret(stringToSecretKind[i.kind], i.name)
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

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (c *Client) runShell(ctx context.Context) {
	// Init input model
	input := textinput.New()
	input.Prompt = "$ "
	input.Placeholder = "filepath save to"
	input.CharLimit = 50

	// Init list model
	items := []list.Item{}

	secrets, err := c.storage.GetSecretsByUser(ctx, c.config.User)
	if err != nil {
		c.log.Error().Err(err).Msg("failed to list user '%s' secrets")
		return
	}

	for _, secret := range secrets {
		if secret.Deleted {
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
		choices = append(choices, choiceItem(secretKindString))
	}

	// Setup TUI
	m := model{
		goph:    c,
		list:    list.New(items, list.NewDefaultDelegate(), 0, 0),
		choices: list.New(choices, list.NewDefaultDelegate(), 0, 0),
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
