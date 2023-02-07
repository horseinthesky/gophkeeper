package client

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

var (
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle         = focusedStyle.Copy()
	noStyle             = lipgloss.NewStyle()
	helpStyle           = blurredStyle.Copy()
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	focusedButton = focusedStyle.Copy().Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

var entryMap = map[SecretKind]func() []textinput.Model{
	SecretCreds: newCreds,
	SecretText: newText,
	SecretBytes: newBytes,
	SecretCard: newCard,
}

func newCreds() []textinput.Model {
	inputs := make([]textinput.Model, 4)

	var t textinput.Model
	for i := range inputs {
		t = textinput.New()
		t.CursorStyle = cursorStyle
		t.CharLimit = 32

		switch i {
		case 0:
			t.Placeholder = "Secret Name"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 1:
			t.Placeholder = "Login"
		case 2:
			t.Placeholder = "Password"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		case 3:
			t.Placeholder = "Notes"
		}

		inputs[i] = t
	}

	return inputs
}

func newText() []textinput.Model {
	inputs := make([]textinput.Model, 3)

	var t textinput.Model
	for i := range inputs {
		t = textinput.New()
		t.CursorStyle = cursorStyle
		t.CharLimit = 100

		switch i {
		case 0:
			t.Placeholder = "Secret Name"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 1:
			t.Placeholder = "Text"
		case 2:
			t.Placeholder = "Notes"
		}

		inputs[i] = t
	}

	return inputs
}

func newBytes() []textinput.Model {
	inputs := make([]textinput.Model, 3)

	var t textinput.Model
	for i := range inputs {
		t = textinput.New()
		t.CursorStyle = cursorStyle
		t.CharLimit = 100

		switch i {
		case 0:
			t.Placeholder = "Secret Name"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 1:
			t.Placeholder = "Path to file"
		case 2:
			t.Placeholder = "Notes"
		}

		inputs[i] = t
	}

	return inputs
}

func newCard() []textinput.Model {
	inputs := make([]textinput.Model, 5)

	var t textinput.Model
	for i := range inputs {
		t = textinput.New()
		t.CursorStyle = cursorStyle
		t.CharLimit = 100

		switch i {
		case 0:
			t.Placeholder = "Secret Name"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 1:
			t.Placeholder = "Card Number"
		case 2:
			t.Placeholder = "Owner"
		case 3:
			t.Placeholder = "EXP"
		case 4:
			t.Placeholder = "CVV"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}

		inputs[i] = t
	}

	return inputs
}
