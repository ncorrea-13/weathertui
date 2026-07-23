package tui

import "github.com/charmbracelet/lipgloss"

// Every role below gets its own color — nothing here is reused for two
// different meanings, except the box border deliberately matching the
// header (it's the same "frame" element).
var (
	colorCyan   = lipgloss.Color("14")  // header title/url + box border
	colorBlue   = lipgloss.Color("12")  // City/Condition labels
	colorGreen  = lipgloss.Color("10")  // City value
	colorYellow = lipgloss.Color("11")  // Temperature title
	colorMag    = lipgloss.Color("13")  // Next 24h / 5-day forecast titles
	colorWhite  = lipgloss.Color("15")  // Condition value
	colorGray   = lipgloss.Color("8")   // footer / hour labels / secondary text
	colorRed    = lipgloss.Color("9")   // errors
	colorTeal   = lipgloss.Color("43")  // Wind title
	colorPurple = lipgloss.Color("141") // Atmosphere title
	colorAmber  = lipgloss.Color("208") // Pressure label
	colorSky    = lipgloss.Color("39")  // Precipitation label
	colorAqua   = lipgloss.Color("87")  // Humidity label
	colorOrange = lipgloss.Color("215") // sparkline bars
	colorLilac  = lipgloss.Color("183") // forecast day names (Sat/Sun/...)

	styleTitle = lipgloss.NewStyle().Foreground(colorCyan).Bold(true)
	styleURL   = lipgloss.NewStyle().Foreground(colorCyan)
	styleLabel = lipgloss.NewStyle().Foreground(colorBlue).Bold(true)
	styleValue = lipgloss.NewStyle().Foreground(colorGreen)
	styleWhite = lipgloss.NewStyle().Foreground(colorWhite)
	styleGray  = lipgloss.NewStyle().Foreground(colorGray)
	styleError = lipgloss.NewStyle().Foreground(colorRed).Bold(true)

	styleSectionTitleTemp = lipgloss.NewStyle().Foreground(colorYellow).Bold(true)
	styleSectionTitleWind = lipgloss.NewStyle().Foreground(colorTeal).Bold(true)
	styleSectionTitleAtmo = lipgloss.NewStyle().Foreground(colorPurple).Bold(true)
	styleSectionTitleFc   = lipgloss.NewStyle().Foreground(colorMag).Bold(true)

	styleBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorCyan).
			Padding(0, 2)

	styleForecastCol = lipgloss.NewStyle().
				Align(lipgloss.Center).
				Width(11)

	styleForecastDay = lipgloss.NewStyle().Foreground(colorLilac).Bold(true)
	styleSpark       = lipgloss.NewStyle().Foreground(colorOrange)
)

func stylePressureLabel(s string) string { return lipgloss.NewStyle().Foreground(colorAmber).Render(s) }
func stylePrecipLabel(s string) string   { return lipgloss.NewStyle().Foreground(colorSky).Render(s) }
func styleHumidityLabel(s string) string { return lipgloss.NewStyle().Foreground(colorAqua).Render(s) }
