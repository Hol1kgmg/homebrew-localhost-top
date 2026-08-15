package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Hol1kgmg/homebrew-localhost-top/internal/i18n"
	"github.com/Hol1kgmg/homebrew-localhost-top/internal/ui"
)

var version = "dev"

func main() {
	i18n.SetLang(i18n.Detect())

	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Println("localhost-top " + version)
		return
	}

	p := tea.NewProgram(ui.New(version), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, i18n.T("startup_error")+"\n", err)
		os.Exit(1)
	}
}
