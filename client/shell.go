package client

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type newSecretAddition struct {
	choices   []string // items on the to-do list
	cursor    int      // which to-do list item our cursor is pointing at
	cancelled bool
}

func newSecret() newSecretAddition {
	return newSecretAddition{
		choices: []string{"creds", "card", "text"},
	}
}

func (m newSecretAddition) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m newSecretAddition) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			m.cancelled = true
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			// _, ok := m.selected[m.cursor]
			// if ok {
			// 	delete(m.selected, m.cursor)
			// } else {
			// 	m.selected[m.cursor] = struct{}{}
			// }
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m newSecretAddition) View() string {
	if m.cancelled {
		return ""
	}

	// The header
	s := "What kind of secret you wanna add?\n\n"

	// Iterate over our choices
	for i, choice := range m.choices {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Is this choice selected?
		// checked := " " // not selected
		// if _, ok := m.selected[i]; ok {
		// 	checked = "x" // selected!
		// }

		// Render the row
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	// The footer
	s += "\nPress q to quit.\n"

	// Send the UI for rendering
	return s
}

// func (c *Client) Shell() {
// 	p := tea.NewProgram(newSecret())
// 	if _, err := p.Run(); err != nil {
// 		c.log.Error().Err(err).Msg("Alas, there's been a shell error")
// 	}
// 	c.log.Info().Msg("shell shut down")
// }
