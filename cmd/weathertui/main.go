// Copyright (C) 2026  ncorrea-13
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ncorrea-13/weathertui/internal/config"
	"github.com/ncorrea-13/weathertui/internal/tui"
)

func main() {
	cfg, err := config.Ensure()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", tui.IconError, err)
		os.Exit(1)
	}

	// tui.New(cfg) devuelve un tui.Model, y tea.NewProgram pide algo que
	// cumpla la interfaz tea.Model (Init/Update/View). En ningún lado del
	// código de tui dice "tui.Model implements tea.Model" — Go chequea la
	// interfaz de forma implícita/estructural: si el struct tiene esos tres
	// métodos con esa firma, ya la satisface. Es lo más parecido en Go a
	// los traits de Rust, pero sin un `impl Trait for Type` explícito: la
	// única forma de confirmar que tui.Model realmente la cumple es que
	// esta línea compile.
	p := tea.NewProgram(tui.New(cfg), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", tui.IconError, err)
		os.Exit(1)
	}
}
