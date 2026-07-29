# weathertui

Terminal UI for checking the current weather, written in Go.

![[screenshots/example.png]](https://github.com/ncorrea-13/weathertui/blob/main/screenshots/example.png)

Inspired in [meteo-cli](https://codeberg.org/victorhck/meteo-cli) by Victorhck, adapted to use OpenWeatherMap instead of Meteoclimatic. There is also a minimal bash version in the /scripts directory, which is kept as a reference and lightweight alternative. It works more as a lite version.

## Requirements

- Go 1.26+
- An OpenWeatherMap API key (free at <https://openweathermap.org/api>)

## Installation

### With `go install` (recommended)

```bash
go install github.com/ncorrea-13/weathertui/cmd/weathertui@latest
```

### Prebuilt binaries

You can download them from [Github Releases](https://github.com/ncorrea-13/weathertui/releases).

### From source

```bash
git clone https://github.com/ncorrea-13/weathertui
cd weathertui
make build      # produces ./weathertui at the project root
# or:
make install    # equivalent to `go install ./cmd/weathertui`
```

## Usage

```bash
./weathertui
```

On first run it asks for the API key, the city, and the country code, and saves them to `~/.config/openweather.conf`:

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

## `scripts/weathertui.sh`

It's kept in the repo as a reference/lightweight alternative.

**What it does:**

1. Reads `~/.config/openweather.conf`. If the API key is missing, it exits with an error.
2. Builds the query and makes the request with `curl -fsSL`.
3. Parses the JSON response with `jq`.
4. Maps the OpenWeatherMap condition code to a Nerd Font icon + label.
5. Draws everything in the box.

**Usage:**

```bash
./scripts/weathertui.sh              # show the weather once and exit
./scripts/weathertui.sh -w           # watch mode, refresh every 60s (Ctrl+C to quit)
./scripts/weathertui.sh -w 30        # watch mode with a custom interval (30s)
./scripts/weathertui.sh -h           # help
```

**Dependencies:** `bash`, `curl`, `jq`, and a Nerd Font installed so the icons render correctly.

## License

[GPL-3.0](LICENSE), same as [meteo-cli](https://codeberg.org/victorhck/meteo-cli), the project this one is inspired by.
