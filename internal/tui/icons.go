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

// Nerd Font glyphs, same set used by the original weather.sh so the look stays consistent.
const (
	iLocation  = "" // nf-fa-map_marker
	iGlobe     = "" // nf-fa-globe
	iThermo    = "" // nf-fa-thermometer
	iHumidity  = "" // nf-weather-humidity
	iArrowUp   = "" // nf-fa-arrow_up
	iArrowDown = "" // nf-fa-arrow_down
	iNow       = "" // nf-fa-dot_circle_o
	iFeels     = "" // nf-fa-hand_paper_o
	iBarometer = "" // nf-weather-barometer
	iWind      = "" // nf-weather-strong_wind
	iRaindrop  = "" // nf-weather-raindrop
	IconError  = "" // nf-fa-times_circle
	iClock     = "" // nf-fa-clock_o
	iRefresh   = "" // nf-fa-refresh
	iCalendar  = "" // nf-fa-calendar

	iSpeed = "" // nf-fa-tachometer
	iGust  = "" // nf-weather-windy
	iDir   = "" // nf-fa-compass

	iStorm   = "" // nf-weather-thunderstorm
	iDrizzle = "" // nf-weather-showers
	iRainy   = "" // nf-weather-rain
	iSnow    = "" // nf-weather-snow
	iFog     = "" // nf-weather-fog
	iClear   = "" // nf-weather-day_sunny
	iPartly  = "" // nf-weather-day_cloudy
	iCloudy  = "" // nf-weather-cloudy
)

func skyIcon(code int, fallback string) string {
	switch {
	case code >= 200 && code < 300:
		return iStorm + " Storm"
	case code >= 300 && code < 400:
		return iDrizzle + " Drizzle"
	case code >= 500 && code < 600:
		return iRainy + " Rain"
	case code >= 600 && code < 700:
		return iSnow + " Snow"
	case code >= 700 && code < 800:
		return iFog + " Fog"
	case code == 800:
		return iClear + " Clear"
	case code == 801:
		return iPartly + " Slightly cloudy"
	case code == 802:
		return iPartly + " Partly cloudy"
	case code == 803 || code == 804:
		return iCloudy + " Cloudy"
	default:
		return fallback
	}
}

// skyIconOnly is the same mapping but returns just the glyph, for compact forecast rows.
func skyIconOnly(code int) string {
	switch {
	case code >= 200 && code < 300:
		return iStorm
	case code >= 300 && code < 400:
		return iDrizzle
	case code >= 500 && code < 600:
		return iRainy
	case code >= 600 && code < 700:
		return iSnow
	case code >= 700 && code < 800:
		return iFog
	case code == 800:
		return iClear
	case code == 801, code == 802:
		return iPartly
	case code == 803, code == 804:
		return iCloudy
	default:
		return iCloudy
	}
}
