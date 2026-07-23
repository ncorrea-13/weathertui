# weathertui

Terminal UI for checking the current weather, written in Go. Uses OpenWeatherMap.

Inspired in [meteo-cli](https://codeberg.org/victorhck/meteo-cli) by Victorhck, adapted to use OpenWeatherMap instead of Meteoclimatic. `weathertui` is the Go port of that same idea: an interactive TUI with Bubble Tea instead of a bash loop with `curl`+`jq`. There is also a minimal bash version in the /scripts directory, which is kept as a reference and lightweight alternative. It works more as a lite version.

## Requirements

- Go 1.26+
- An OpenWeatherMap API key (free at <https://openweathermap.org/api>)

## Installation

### With `go install` (recommended)

```bash
go install github.com/ncorrea-13/weathertui/cmd/weathertui@latest
```

This builds and installs the `weathertui` binary into `$(go env GOPATH)/bin` (or `$GOBIN` if set). Make sure that folder is on your `$PATH`.

### From source

```bash
git clone git@github.com:ncorrea-13/weathertui.git
cd weathertui
make build      # produces ./weathertui at the project root
# or:
make install    # equivalent to `go install ./cmd/weathertui`
```

## Usage

```bash
./weathertui
```

On first run it asks for the API key, the city, and the country code (e.g. `AR`), and saves them to `~/.config/openweather.conf`:

```
OWM_API_KEY="your-api-key"
CITY="Mendoza"
COUNTRY="AR"
```

Subsequent runs read that file directly, without asking again.

## Development

```bash
make run      # go run ./cmd/weathertui
make test     # go test ./...
make install  # go install ./cmd/weathertui
make clean    # removes the compiled binary
```

## Structure

```
cmd/weathertui/     — entrypoint (wiring: config + tea.Program)
internal/config/    — reads/writes ~/.config/openweather.conf
internal/owm/        — OpenWeatherMap API client
internal/tui/        — Bubble Tea model (Model/Update/View), styles, icons
```

## `scripts/weather.sh` (the original bash version)

It's kept in the repo as a reference/lightweight alternative.

**What it does:**

1. Reads `~/.config/openweather.conf`. If the API key is missing, it exits with an error. If there's no city, it prompts for one via `stdin` and saves it with `600` permissions.
2. Builds the `/data/2.5/weather` API URL with the city, country, key and `units=metric`, and makes the request with `curl -fsSL`.
3. Parses the JSON response with `jq`.
4. Maps the OpenWeatherMap condition code to a Nerd Font icon + label.
5. Draws everything in a box with Unicode borders, with its own helpers to center text and compute the visible width of each cell while ignoring ANSI color codes.

**Usage:**

```bash
./scripts/weather.sh              # show the weather once and exit
./scripts/weather.sh -w           # watch mode, refresh every 60s (Ctrl+C to quit)
./scripts/weather.sh -w 30        # watch mode with a custom interval (30s)
./scripts/weather.sh -h           # help
```

**Dependencies:** `bash`, `curl`, `jq`, and a terminal with [Nerd Fonts](https://www.nerdfonts.com/) so the icons render correctly (otherwise they'll show up as blank spaces or boxes).

## License

[GPL-3.0](LICENSE), same as [meteo-cli](https://codeberg.org/victorhck/meteo-cli), the project this one is inspired by.
