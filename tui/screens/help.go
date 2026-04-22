package screens

import (
	"strings"
)

type KeyBinding struct {
	Keys string
	Desc string
}

type KeyGroup struct {
	Title    string
	Bindings []KeyBinding
}

type Helpable interface {
	HelpTitle() string
	HelpGroups() []KeyGroup
}

var GlobalHelpGroup = KeyGroup{
	Title: "Global",
	Bindings: []KeyBinding{
		{Keys: "?", Desc: "Toggle this help"},
		{Keys: "esc / q", Desc: "Back"},
		{Keys: "ctrl+c", Desc: "Quit"},
	},
}

func RenderHelpOverlay(title string, groups []KeyGroup) string {
	var b strings.Builder
	header := "Help -- " + title
	b.WriteString("\n")
	b.WriteString(HeaderStyle.Render(header))
	b.WriteString("\n\n")

	keyWidth := 0
	for _, g := range groups {
		for _, kb := range g.Bindings {
			if n := len(kb.Keys); n > keyWidth {
				keyWidth = n
			}
		}
	}
	for _, kb := range GlobalHelpGroup.Bindings {
		if n := len(kb.Keys); n > keyWidth {
			keyWidth = n
		}
	}

	writeGroup := func(g KeyGroup) {
		b.WriteString("  " + SelectedStyle.Render(g.Title) + "\n")
		for _, kb := range g.Bindings {
			pad := strings.Repeat(" ", keyWidth-len(kb.Keys))
			b.WriteString("    " + NormalStyle.Render(kb.Keys) + pad + "  " + DimStyle.Render(kb.Desc) + "\n")
		}
		b.WriteString("\n")
	}

	for _, g := range groups {
		writeGroup(g)
	}
	writeGroup(GlobalHelpGroup)

	b.WriteString(DimStyle.Render("  Press ? or esc to return."))
	return b.String()
}
