package owm

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"time"

	"github.com/ncorrea-13/weathertui/internal/config"
)

const httpTimeout = 10 * time.Second

var windDirs = [16]string{
	"N", "NNE", "NE", "ENE",
	"E", "ESE", "SE", "SSE",
	"S", "SSW", "SW", "WSW",
	"W", "WNW", "NW", "NNW",
}

// WindDirection converts a wind bearing in degrees to its 16-point compass label.
func WindDirection(deg int) string {
	idx := ((deg + 11) / 22) % 16
	if idx < 0 {
		idx += 16
	}
	return windDirs[idx]
}

// CurrentWeather is the subset of the /weather response we care about.
type CurrentWeather struct {
	Name        string
	CountryCode string
	TempNow     int
	TempMax     int
	TempMin     int
	FeelsLike   int
	Humidity    int
	Pressure    int
	WindSpeed   int // km/h
	WindGust    string
	WindDeg     int
	Rain        float64
	IconID      int
	Desc        string
}

type owmCurrentResp struct {
	Name string `json:"name"`
	Sys  struct {
		Country string `json:"country"`
	} `json:"sys"`
	Main struct {
		Temp      float64 `json:"temp"`
		TempMin   float64 `json:"temp_min"`
		TempMax   float64 `json:"temp_max"`
		FeelsLike float64 `json:"feels_like"`
		Humidity  int     `json:"humidity"`
		Pressure  int     `json:"pressure"`
	} `json:"main"`
	Wind struct {
		Speed float64 `json:"speed"`
		Deg   int     `json:"deg"`
		// Gust es *float64 (puntero) y no float64 a propósito: la API omite
		// por completo el campo "gust" cuando no hay ráfagas que reportar.
		// Con float64 no habría forma de distinguir "no vino el campo" de
		// "vino y es 0.0"; nil vs valor-seteado es la forma en Go de tener
		// lo que en Rust sería Option<f64> — sin envoltorio, es directamente
		// un puntero que puede ser nil. Se resuelve más abajo en
		// FetchCurrent con el `if r.Wind.Gust != nil`.
		Gust *float64 `json:"gust"`
	} `json:"wind"`
	Rain    map[string]float64 `json:"rain"`
	Weather []struct {
		ID          int    `json:"id"`
		Description string `json:"description"`
	} `json:"weather"`
}

func msToKmh(ms float64) int {
	return int(math.Round(ms * 3.6))
}

// fetchJSON no recibe context.Context — client.Get() usa el timeout fijo de
// httpTimeout (10s) como único mecanismo de corte, no hay forma de
// cancelarlo antes desde el caller (ver nota en tui/model.go:fetchCmd sobre
// qué pasa con esto si el usuario cierra el programa a mitad de un fetch).
func fetchJSON(reqURL string, out any) error {
	client := &http.Client{Timeout: httpTimeout}

	resp, err := client.Get(reqURL)
	if err != nil {
		return fmt.Errorf("error de red: %w", err)
	}
	// defer se evalúa (resp.Body se captura) en el momento de esta línea,
	// pero se ejecuta al retornar la función, sea por cualquiera de los
	// tres `return` de abajo o por un panic. Con múltiples puntos de salida
	// como acá, es lo que evita repetir el Close() en cada uno.
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error leyendo respuesta: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr struct {
			Message string `json:"message"`
		}
		// Error ignorado a propósito (`_ =`): esto es "best effort". Si el
		// body de un error HTTP no es JSON válido, apiErr se queda en su
		// zero value (Message == "") y el fmt.Errorf de abajo simplemente
		// imprime un mensaje vacío — no hay nada mejor que hacer con ese
		// error de parseo, ya estamos construyendo un error a partir de
		// otro error.
		_ = json.Unmarshal(body, &apiErr)
		return fmt.Errorf("API error (%d): %s", resp.StatusCode, apiErr.Message)
	}

	if err := json.Unmarshal(body, out); err != nil {
		// %w (no %v) envuelve el error original preservando su cadena, para
		// que quien llame más arriba pueda hacer errors.Is/errors.As sobre
		// él si necesita distinguir el tipo de fallo. Nadie en este
		// proyecto hace ese unwrap hoy, pero wrappear con %w al propagar es
		// el default idiomático en Go — %v perdería esa cadena para
		// siempre.
		return fmt.Errorf("error parseando JSON: %w", err)
	}
	return nil
}

func FetchCurrent(cfg config.Config) (CurrentWeather, error) {
	var cw CurrentWeather

	reqURL := "https://api.openweathermap.org/data/2.5/weather?" + url.Values{
		"q":     {queryLocation(cfg)},
		"appid": {cfg.APIKey},
		"units": {"metric"},
		"lang":  {"en"},
	}.Encode()

	var r owmCurrentResp
	if err := fetchJSON(reqURL, &r); err != nil {
		return cw, err
	}

	cw.Name = r.Name
	cw.CountryCode = r.Sys.Country
	cw.TempNow = int(math.Round(r.Main.Temp))
	cw.TempMax = int(math.Round(r.Main.TempMax))
	cw.TempMin = int(math.Round(r.Main.TempMin))
	cw.FeelsLike = int(math.Round(r.Main.FeelsLike))
	cw.Humidity = r.Main.Humidity
	cw.Pressure = r.Main.Pressure
	cw.WindSpeed = msToKmh(r.Wind.Speed)
	cw.WindDeg = r.Wind.Deg
	if r.Wind.Gust != nil {
		cw.WindGust = fmt.Sprintf("%d", msToKmh(*r.Wind.Gust))
	} else {
		cw.WindGust = "-"
	}
	if v, ok := r.Rain["1h"]; ok {
		cw.Rain = v
	} else if v, ok := r.Rain["3h"]; ok {
		cw.Rain = v
	}
	if len(r.Weather) > 0 {
		cw.IconID = r.Weather[0].ID
		cw.Desc = r.Weather[0].Description
	}

	return cw, nil
}

// ForecastPoint is a single 3-hour step from the /forecast endpoint.
type ForecastPoint struct {
	Time   time.Time
	Temp   float64
	IconID int
	Pop    float64 // probability of precipitation, 0..1
}

// ForecastDay groups the 3-hour steps of a calendar day into a min/max summary.
type ForecastDay struct {
	Date   time.Time
	Min    int
	Max    int
	IconID int
}

type ForecastData struct {
	Points []ForecastPoint
	Days   []ForecastDay
}

type owmForecastResp struct {
	List []struct {
		Dt   int64 `json:"dt"`
		Main struct {
			Temp float64 `json:"temp"`
		} `json:"main"`
		Weather []struct {
			ID int `json:"id"`
		} `json:"weather"`
		Pop float64 `json:"pop"`
	} `json:"list"`
}

func FetchForecast(cfg config.Config) (ForecastData, error) {
	var fd ForecastData

	reqURL := "https://api.openweathermap.org/data/2.5/forecast?" + url.Values{
		"q":     {queryLocation(cfg)},
		"appid": {cfg.APIKey},
		"units": {"metric"},
		"lang":  {"en"},
	}.Encode()

	var r owmForecastResp
	if err := fetchJSON(reqURL, &r); err != nil {
		return fd, err
	}

	type dayAgg struct {
		date        time.Time
		min, max    float64
		iconAtNoon  int
		bestNoonGap time.Duration
		set         bool
	}
	// dayOrder existe solo por esto: iterar un map en Go no tiene orden
	// garantizado (el runtime lo randomiza a propósito, justamente para que
	// nadie dependa de un orden implícito). days sirve para agregar por
	// clave "YYYY-MM-DD" en O(1); dayOrder es la lista aparte que recuerda
	// en qué orden cronológico apareció cada clave por primera vez, para
	// poder reconstruir fd.Days abajo en orden.
	dayOrder := []string{}
	days := map[string]*dayAgg{}

	for _, e := range r.List {
		t := time.Unix(e.Dt, 0).Local()
		iconID := 0
		if len(e.Weather) > 0 {
			iconID = e.Weather[0].ID
		}

		fd.Points = append(fd.Points, ForecastPoint{
			Time:   t,
			Temp:   e.Main.Temp,
			IconID: iconID,
			Pop:    e.Pop,
		})

		key := t.Format("2006-01-02")
		d, ok := days[key]
		if !ok {
			d = &dayAgg{date: t, min: e.Main.Temp, max: e.Main.Temp}
			days[key] = d
			dayOrder = append(dayOrder, key)
		}
		if e.Main.Temp < d.min {
			d.min = e.Main.Temp
		}
		if e.Main.Temp > d.max {
			d.max = e.Main.Temp
		}
		noon := time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, t.Location())
		gap := t.Sub(noon)
		if gap < 0 {
			gap = -gap
		}
		if !d.set || gap < d.bestNoonGap {
			d.bestNoonGap = gap
			d.iconAtNoon = iconID
			d.set = true
		}
	}

	for _, key := range dayOrder {
		d := days[key]
		fd.Days = append(fd.Days, ForecastDay{
			Date:   d.date,
			Min:    int(math.Round(d.min)),
			Max:    int(math.Round(d.max)),
			IconID: d.iconAtNoon,
		})
	}

	return fd, nil
}

func queryLocation(cfg config.Config) string {
	if cfg.Country != "" {
		return cfg.City + "," + cfg.Country
	}
	return cfg.City
}
