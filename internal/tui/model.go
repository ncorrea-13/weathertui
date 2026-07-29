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

type tickMsg time.Time

type dataMsg struct {
	current  owm.CurrentWeather
	forecast owm.ForecastData
}

type errMsg struct{ err error }

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

func fetchCmd(cfg config.Config) tea.Cmd {
	return func() tea.Msg {
		cw, err := owm.FetchCurrent(cfg)
		if err != nil {
			return errMsg{err}
		}
		fc, err := owm.FetchForecast(cfg)
		if err != nil {
			return errMsg{err}
		}
		return dataMsg{current: cw, forecast: fc}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "r":
			m.loading = true
			return m, fetchCmd(m.cfg)
		}
		return m, nil

	case tickMsg:
		m.nextRefresh = time.Now().Add(refreshInterval)
		return m, tea.Batch(fetchCmd(m.cfg), scheduleTick())

	case dataMsg:
		m.current = msg.current
		m.forecast = msg.forecast
		m.haveData = true
		m.loading = false
		m.err = nil
		m.lastUpdated = time.Now()
		if m.nextRefresh.IsZero() {
			m.nextRefresh = time.Now().Add(refreshInterval)
		}
		return m, nil

	case errMsg:
		m.err = msg.err
		m.loading = false
		if m.nextRefresh.IsZero() {
			m.nextRefresh = time.Now().Add(refreshInterval)
		}
		return m, nil
	}

	return m, nil
}
