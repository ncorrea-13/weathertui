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

package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ncorrea-13/weathertui/internal/config"
	"github.com/ncorrea-13/weathertui/internal/owm"
)

const refreshInterval = 5 * time.Second

// tickMsg, dataMsg y errMsg son los tea.Msg que este modelo sabe manejar.
// tea.Msg es "any" (interface{}) a nivel de firma —Bubbletea no tiene forma
// de restringirlo a un enum cerrado como haría Rust con un match exhaustivo—
// así que la exhaustividad del switch en Update() la garantiza el
// programador, no el compilador. Cada tipo concreto que Init/Update
// devuelven en un tea.Cmd termina, tarde o temprano, en uno de estos casos.
type tickMsg time.Time

type dataMsg struct {
	current  owm.CurrentWeather
	forecast owm.ForecastData
}

type errMsg struct{ err error }

// Model is the bubbletea model driving the weather dashboard: it owns the
// fetched data and refresh state. Rendering lives in view.go.
type Model struct {
	cfg config.Config

	current     owm.CurrentWeather
	forecast    owm.ForecastData
	haveData    bool
	err         error
	loading     bool
	lastUpdated time.Time
	nextRefresh time.Time

	width, height int
}

// New builds the initial model for the given config, ready to hand to
// tea.NewProgram.
//
// Model se pasa y devuelve por valor en todo este archivo (nunca *Model).
// Es la contraparte, en Go, del modelo de ownership de Rust: acá no hay
// "mover" el struct ni pedirlo prestado, cada vez que Update recibe un
// Model se lleva su propia copia (shallow copy de sus campos), la muta
// libremente y la devuelve. El struct original en manos de Bubbletea queda
// intacto hasta que esa copia vuelve. Es el patrón "immutable update" típico
// de arquitecturas Elm, y es lo que permite no necesitar mutex en ningún
// lado: el único lugar que puede tocar el estado es Update, y corre en un
// único goroutine (el loop de eventos de Bubbletea).
func New(cfg config.Config) Model {
	return Model{cfg: cfg, loading: true}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(fetchCmd(m.cfg), scheduleTick())
}

func scheduleTick() tea.Cmd {
	return tea.Tick(refreshInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// fetchCmd envuelve la llamada HTTP en un tea.Cmd, que es literalmente
// `func() tea.Msg`. Bubbletea ejecuta ese closure en su propio goroutine y
// mete el tea.Msg resultante en un channel interno que el loop principal
// lee y despacha a Update. Por eso fetchCmd puede hacer I/O bloqueante
// (net/http es sync) sin congelar la UI: el bloqueo pasa en un goroutine
// aparte, nunca en el que dibuja pantalla.
//
// Nota de concurrencia que falta: ninguna de las dos llamadas a owm.Fetch*
// recibe un context.Context, así que si el usuario aprieta "q" mientras hay
// un fetch en vuelo, ese request HTTP sigue viajando igual — no hay forma
// de cancelarlo desde este código. Bubbletea corta el programa, pero el
// http.Client sigue esperando su respuesta en background hasta el timeout
// de 10s de owm.httpTimeout. Con context.Context propagado desde acá se
// podría cancelar ese Get() al salir.
func fetchCmd(cfg config.Config) tea.Cmd {
	return func() tea.Msg {
		cw, err := owm.FetchCurrent(cfg)
		if err != nil {
			return errMsg{err}
		}
		// Si el clima actual falló ya devolvimos arriba: no tiene sentido
		// pedir el forecast si ni siquiera pudimos autenticar/conectar.
		// Es "fail fast" secuencial, no hay goroutines separadas para cada
		// endpoint porque total tenemos que esperar a los dos antes de
		// poder dibujar el dashboard.
		fc, err := owm.FetchForecast(cfg)
		if err != nil {
			return errMsg{err}
		}
		return dataMsg{current: cw, forecast: fc}
	}
}

// Update es el único punto de mutación del estado en todo el programa. El
// type switch sobre msg.(type) es la forma idiomática de Go de simular
// pattern matching sobre una interfaz: en Rust esto sería un `match` sobre
// un enum y el compilador te avisaría si te falta un variant; en Go, si
// aparece un tea.Msg nuevo que no se cubre, simplemente cae al `return m,
// nil` final sin aviso del compilador. La responsabilidad de que el switch
// esté completo es del programador.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			// tea.Quit es en sí mismo un tea.Cmd (una func() tea.Msg que
			// devuelve tea.QuitMsg{}); no aborta el programa desde acá,
			// solo le indica al runtime que termine su loop prolijamente
			// en la próxima vuelta.
			return m, tea.Quit
		case "r":
			m.loading = true
			return m, fetchCmd(m.cfg)
		}
		return m, nil

	case tickMsg:
		// Bubbletea no tiene setInterval: tea.Tick dispara una sola vez.
		// Para tener un refresh periódico hay que volver a agendar el
		// próximo tick cada vez que el anterior llega — de ahí que
		// scheduleTick() se vuelva a meter en el tea.Batch de esta rama.
		m.nextRefresh = time.Now().Add(refreshInterval)
		return m, tea.Batch(fetchCmd(m.cfg), scheduleTick())

	case dataMsg:
		m.current = msg.current
		m.forecast = msg.forecast
		m.haveData = true
		m.loading = false
		m.err = nil
		m.lastUpdated = time.Now()
		// time.Time{} (zero value) es un valor válido y usable sin construir
		// nada: IsZero() es la forma idiomática de preguntar "¿esto nunca se
		// seteó?" en vez de necesitar un Option<time.Time>/nullable como en
		// Rust o PL/SQL. Este chequeo cubre el primer fetch exitoso, que
		// llega antes que el primer tickMsg y por lo tanto antes de que
		// nextRefresh se haya seteado por esa vía.
		if m.nextRefresh.IsZero() {
			m.nextRefresh = time.Now().Add(refreshInterval)
		}
		return m, nil

	case errMsg:
		// A propósito NO se pisa m.current/m.forecast: si ya había datos
		// de un fetch anterior, se siguen mostrando (dashboard "stale pero
		// visible") y solo se agrega el error en el footer. Ver
		// view.go:renderFooter.
		m.err = msg.err
		m.loading = false
		if m.nextRefresh.IsZero() {
			m.nextRefresh = time.Now().Add(refreshInterval)
		}
		return m, nil
	}

	return m, nil
}
