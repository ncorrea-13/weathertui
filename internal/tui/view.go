package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/ncorrea-13/weathertui/internal/owm"
)

func (m Model) View() string {
	var body string

	switch {
	case !m.haveData && m.loading:
		body = styleGray.Render("Cargando " + m.cfg.City + "...")
	case !m.haveData && m.err != nil:
		body = styleError.Render(IconError+" "+m.err.Error()) + "\n\n" +
			styleGray.Render("r = reintentar · q = salir")
	default:
		body = m.renderDashboard()
	}

	box := styleBox.Render(body)

	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
	}
	return box
}

func (m Model) renderDashboard() string {
	sections := []string{
		m.renderHeader(),
		"",
		m.renderLocation(),
		"",
		m.renderTemperature(),
		"",
		m.renderWindAtmosphere(),
	}

	if len(m.forecast.Points) > 0 {
		sections = append(sections, "", m.renderSparkline())
	}
	if len(m.forecast.Days) > 0 {
		sections = append(sections, "", m.renderForecast())
	}

	sections = append(sections, "", m.renderFooter())

	// Center every section as a block against the widest one (usually the
	// 5-day forecast row), instead of leaving narrower sections flush left.
	return lipgloss.JoinVertical(lipgloss.Center, sections...)
}

func (m Model) renderHeader() string {
	title := styleTitle.Render(iClear + "  OPENWEATHERMAP")
	url := styleURL.Render("https://openweathermap.org/")
	return lipgloss.JoinVertical(lipgloss.Center, title, url)
}

func (m Model) renderLocation() string {
	loc := fmt.Sprintf("%s %s%s, %s", iLocation, styleLabel.Render(padRight("City", 11)), styleValue.Render(m.current.Name), m.current.CountryCode)
	cond := fmt.Sprintf("%s %s%s", iGlobe, styleWhite.Render(padRight("Condition", 11)), skyIcon(m.current.IconID, m.current.Desc))
	return lipgloss.JoinVertical(lipgloss.Left, loc, cond)
}

// padRight pads plain ASCII labels (no icons/wide runes) to width. Only use
// it on strings you know are single-byte-per-cell; for anything with icons
// or accented/wide runes use padVisual, which measures actual cell width.
func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// padVisual pads s to width terminal cells, measuring width the same way
// lipgloss does. Needed for anything that mixes icons/degree signs with
// plain text, where len(s) (byte count) would under-pad and ragged the
// box's right border.
func padVisual(s string, width int) string {
	if w := lipgloss.Width(s); w < width {
		s += strings.Repeat(" ", width-w)
	}
	return s
}

// centerTitle centers a rendered title within width, widening to fit the
// title itself if it's already longer than the column it sits over.
func centerTitle(title string, width int) string {
	if w := lipgloss.Width(title); w > width {
		width = w
	}
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, title)
}

// iconPrefixWidth is the terminal-cell width of a leading "icon + space"
// prefix (every Nerd Font glyph used here renders as 1 cell).
const iconPrefixWidth = 2

// centerTitleAfterIcon centers title over textWidth (the row text with its
// icon prefix stripped out), then shifts it right by iconPrefixWidth so it
// still lines up with icon-prefixed rows below it. Without this, the icon
// pulls the visible text rightward but the title stays centered over the
// icon+text combo, so the title looks off-center relative to what you
// actually read.
func centerTitleAfterIcon(title string, textWidth int) string {
	return strings.Repeat(" ", iconPrefixWidth) + centerTitle(title, textWidth)
}

// twoColRows lays out label/value pairs as two aligned columns, sizing the
// left column to whatever its longest row actually needs (measured in
// terminal cells, not bytes) instead of a hardcoded width.
func twoColRows(pairs [][2]string) string {
	leftWidth := 0
	for _, p := range pairs {
		if w := lipgloss.Width(p[0]); w > leftWidth {
			leftWidth = w
		}
	}
	leftWidth += 2

	rows := make([]string, len(pairs))
	for i, p := range pairs {
		rows[i] = padVisual(p[0], leftWidth) + p[1]
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m Model) renderTemperature() string {
	c := m.current

	body := twoColRows([][2]string{
		{fmt.Sprintf("%s Now   : %d °C", iNow, c.TempNow), fmt.Sprintf("%s Max : %d °C", iArrowUp, c.TempMax)},
		{fmt.Sprintf("%s Feels : %d °C", iFeels, c.FeelsLike), fmt.Sprintf("%s Min : %d °C", iArrowDown, c.TempMin)},
	})
	title := centerTitle(styleSectionTitleTemp.Render(iThermo+" Temperature"), lipgloss.Width(body))

	return lipgloss.JoinVertical(lipgloss.Left, title, body)
}

// renderWindAtmosphere places Wind (speed/gust/direction) and Atmosphere
// (pressure/precipitation/humidity) side by side, each under its own title
// — grouping pressure and humidity under "Wind" was misleading, they're not
// wind readings.
func (m Model) renderWindAtmosphere() string {
	return lipgloss.JoinHorizontal(lipgloss.Top, m.renderWind(), "   ", m.renderAtmosphere())
}

// iconRows renders icon+text pairs as a left-aligned block and returns it
// alongside the max text width (icon excluded), so a title above it can be
// centered over the text via centerTitleAfterIcon.
func iconRows(pairs [][2]string) (body string, textWidth int) {
	rows := make([]string, len(pairs))
	for i, p := range pairs {
		if w := lipgloss.Width(p[1]); w > textWidth {
			textWidth = w
		}
		rows[i] = p[0] + " " + p[1]
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...), textWidth
}

func (m Model) renderWind() string {
	c := m.current

	body, textWidth := iconRows([][2]string{
		{iSpeed, fmt.Sprintf("Speed : %d km/h", c.WindSpeed)},
		{iGust, fmt.Sprintf("Gust  : %s km/h", c.WindGust)},
		{iDir, fmt.Sprintf("Dir   : %d° (%s)", c.WindDeg, owm.WindDirection(c.WindDeg))},
	})
	title := centerTitleAfterIcon(styleSectionTitleWind.Render(iWind+" Wind"), textWidth)

	return lipgloss.JoinVertical(lipgloss.Left, title, "", body)
}

func (m Model) renderAtmosphere() string {
	c := m.current

	body, textWidth := iconRows([][2]string{
		{iBarometer, fmt.Sprintf("%s %d hPa", stylePressureLabel("Pressure:"), c.Pressure)},
		{iRaindrop, fmt.Sprintf("%s %.1f mm", stylePrecipLabel("Precipitation:"), c.Rain)},
		{iHumidity, fmt.Sprintf("%s %d %%", styleHumidityLabel("Humidity:"), c.Humidity)},
	})
	title := centerTitleAfterIcon(styleSectionTitleAtmo.Render(iBarometer+" Atmosphere"), textWidth)

	return lipgloss.JoinVertical(lipgloss.Left, title, "", body)
}

// sparkBarWidth is how many terminal cells each forecast point's bar draws
// as, and sparkColWidth how much horizontal room its hour label gets —
// widening the chart so each time step actually reads at a glance instead
// of packing into a dense one-line sparkline.
const (
	sparkBarWidth = 4
	sparkColWidth = sparkBarWidth + 2
)

func (m Model) renderSparkline() string {
	pts := m.forecast.Points
	n := 8
	if len(pts) < n {
		n = len(pts)
	}
	slice := pts[:n]

	temps := make([]float64, len(slice))
	min, max := slice[0].Temp, slice[0].Temp
	for i, p := range slice {
		temps[i] = p.Temp
		if p.Temp < min {
			min = p.Temp
		}
		if p.Temp > max {
			max = p.Temp
		}
	}
	// []rune(string) es obligatorio acá, no cosmético: sparkline devuelve
	// bloques Unicode (▁▂▃...) que ocupan 3 bytes cada uno en UTF-8. Indexar
	// el string directo (sparkline(temps)[i]) daría bytes sueltos de un
	// caracter multi-byte, no el caracter completo. Convertir a []rune
	// primero paga el costo de decodificar todo el string una vez, y a
	// cambio barRunes[i] indexa por caracter real.
	barRunes := []rune(sparkline(temps))

	cols := make([]string, len(slice))
	for i, p := range slice {
		bar := styleSpark.Render(strings.Repeat(string(barRunes[i]), sparkBarWidth))
		hour := styleGray.Render(p.Time.Format("15h"))
		cols[i] = lipgloss.NewStyle().Width(sparkColWidth).Align(lipgloss.Center).
			Render(lipgloss.JoinVertical(lipgloss.Center, bar, hour))
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, cols...)

	rangeText := styleGray.Render(fmt.Sprintf("%.0f°C ↔ %.0f°C", min, max))

	// Range on the same row as the title, pushed to the right edge of the
	// chart instead of trailing it on its own line below.
	title := styleSectionTitleFc.Render(iClock + " Next 24h")
	header := padVisual(title, lipgloss.Width(row)-lipgloss.Width(rangeText)) + rangeText

	return lipgloss.JoinVertical(lipgloss.Left, header, "", row)
}

func (m Model) renderForecast() string {
	days := m.forecast.Days
	today := time.Now().Format("2006-01-02")
	if len(days) > 0 && days[0].Date.Format("2006-01-02") == today && len(days) > 5 {
		days = days[1:] // today is already covered by the current-weather section above
	}
	if len(days) > 5 {
		days = days[:5]
	}

	cols := make([]string, len(days))
	for i, d := range days {
		day := styleForecastDay.Render(d.Date.Format("Mon"))
		icon := skyIconOnly(d.IconID)
		temps := fmt.Sprintf("%d°/%d°", d.Min, d.Max)
		cols[i] = styleForecastCol.Render(lipgloss.JoinVertical(lipgloss.Center, day, icon, temps))
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, cols...)

	title := centerTitle(styleSectionTitleFc.Render(iCalendar+" 5-day forecast"), lipgloss.Width(row))
	return lipgloss.JoinVertical(lipgloss.Left, title, "", row)
}

func (m Model) renderFooter() string {
	updated := fmt.Sprintf("%s %s", iClock, m.lastUpdated.Format("15:04:05"))

	var refresh string
	if remaining := time.Until(m.nextRefresh); remaining > 0 {
		refresh = fmt.Sprintf("%s next in %ds", iRefresh, int(remaining.Seconds())+1)
	} else {
		refresh = iRefresh + " refreshing..."
	}

	line := styleGray.Render(fmt.Sprintf("%s · %s · r refresh · q quit", updated, refresh))

	if m.err != nil {
		errLine := styleError.Render(IconError + " " + m.err.Error())
		return lipgloss.JoinVertical(lipgloss.Left, errLine, line)
	}
	return line
}
