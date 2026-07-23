#!/usr/bin/env bash

# Copyright (C) 2026  ncorrea-13
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <https://www.gnu.org/licenses/>.

# Based on the original script by Victorhck (meteo-cli):
#   https://codeberg.org/victorhck/meteo-cli
# Adapted to use OpenWeatherMap instead of Meteoclimatic, so it also works
# for locations outside Spain/Portugal/Andorra (e.g. Argentina).
#
# On first run it asks for a city, saves it to a config file and reuses
# it on subsequent runs.
#
# Usage:
#   ./weather.sh              Show the weather once and exit.
#   ./weather.sh -w [SECONDS] Watch mode: refresh in a loop (default 60s).
#   ./weather.sh -h           Help.

CONFIG="$HOME/.config/openweather.conf"

WATCH=0
INTERVAL=60

#-------------------------------------------------------
# Argument parsing
#-------------------------------------------------------
while [[ $# -gt 0 ]]; do
  case "$1" in
  -w | --watch)
    WATCH=1
    if [[ -n "${2:-}" && "$2" =~ ^[0-9]+$ ]]; then
      INTERVAL=$2
      shift
    fi
    shift
    ;;
  -h | --help)
    echo "Usage: $0 [-w [SECONDS]]"
    echo "  -w, --watch [N]   Keep the script running, refreshing every N seconds (default 60)."
    echo "  -h, --help        Show this help."
    exit 0
    ;;
  *)
    echo "Unknown option: $1"
    exit 1
    ;;
  esac
done

#-------------------------------------------------------
# Icons (Nerd Fonts)
#-------------------------------------------------------
I_LOCATION=$''   # nf-fa-map_marker
I_GLOBE=$''      # nf-fa-globe
I_THERMO=$''     # nf-fa-thermometer
I_HUMIDITY=$''   # nf-weather-humidity
I_ARROW_UP=$''   # nf-fa-arrow_up
I_ARROW_DOWN=$'' # nf-fa-arrow_down
I_NOW=$''        # nf-fa-dot_circle_o
I_FEELS=$''      # nf-fa-hand_paper_o
I_BAROMETER=$''  # nf-weather-barometer
I_WIND=$''       # nf-weather-strong_wind
I_RAINDROP=$''   # nf-weather-raindrop
I_ERROR=$''      # nf-fa-times_circle
I_CLOCK=$''      # nf-fa-clock_o
I_REFRESH=$''    # nf-fa-refresh

I_STORM=$''   # nf-weather-thunderstorm
I_DRIZZLE=$'' # nf-weather-showers
I_RAINY=$''   # nf-weather-rain
I_SNOW=$''    # nf-weather-snow
I_FOG=$''     # nf-weather-fog
I_CLEAR=$''   # nf-weather-day_sunny
I_PARTLY=$''  # nf-weather-day_cloudy
I_CLOUDY=$''  # nf-weather-cloudy

#-------------------------------------------------------
# Read config
#-------------------------------------------------------
if [[ -f "$CONFIG" ]]; then
  while IFS='=' read -r key value; do
    case "$key" in
    OWM_API_KEY)
      OWM_API_KEY=${value//\"/}
      ;;
    CITY)
      CITY=${value//\"/}
      ;;
    COUNTRY)
      COUNTRY=${value//\"/}
      ;;
    esac
  done <"$CONFIG"
fi

if [[ -z "${OWM_API_KEY:-}" ]]; then
  echo "$I_ERROR Could not find OWM_API_KEY in $CONFIG"
  echo "   Add the line: OWM_API_KEY=\"your_key_here\""
  exit 1
fi

if [[ -z "${CITY:-}" ]]; then
  clear
  echo ""
  echo "City to check the weather for"
  echo "Example: Mendoza"
  echo ""
  read -rp "> City: " CITY
  read -rp "> Country code (e.g. AR): " COUNTRY

  {
    echo "OWM_API_KEY=\"$OWM_API_KEY\""
    echo "CITY=\"$CITY\""
    echo "COUNTRY=\"$COUNTRY\""
  } >"$CONFIG"
  chmod 600 "$CONFIG"
  echo
fi

BLUE=$'\e[94m'
GREEN=$'\e[92m'
CYAN=$'\e[96m'
YELLOW=$'\e[93m'
MAG=$'\e[95m'
WHITE=$'\e[97m'
GRAY=$'\e[90m'
RESET=$'\e[0m'

print_color() {
  printf "%b%-*s%b" "$1" "$2" "$3" "$RESET"
}

#-------------------------------------------------------
# Box drawing helpers
#-------------------------------------------------------
BOX_WIDTH=47

TOPLEFT="╔"
TOPRIGHT="╗"
BOTLEFT="╚"
BOTRIGHT="╝"
VBAR="║"
HBAR=$(printf '%*s' "$((BOX_WIDTH + 2))" '')
HBAR=${HBAR// /═}

visible_len() {
  local s=$1 esc=$'\e' plain="" prefix rest
  while [[ "$s" == *"${esc}["* ]]; do
    prefix="${s%%"${esc}["*}"
    rest="${s#*"${esc}["}"
    rest="${rest#*m}"
    plain+="$prefix"
    s="$rest"
  done
  plain+="$s"
  printf '%s' "${#plain}"
}

pad_field() {
  local text="$1" width="$2" len pad
  len=$(visible_len "$text")
  pad=$((width - len))
  ((pad < 0)) && pad=0
  printf '%s%*s' "$text" "$pad" ""
}

box_row() {
  local content="$1" len pad
  len=$(visible_len "$content")
  pad=$((BOX_WIDTH - len))
  ((pad < 0)) && pad=0
  printf "%s %s%*s %s\n" "$VBAR" "$content" "$pad" "" "$VBAR"
}

box_row_centered() {
  local content="$1" bias="${2:-0}" len total left right
  len=$(visible_len "$content")
  total=$((BOX_WIDTH - len))
  ((total < 0)) && total=0
  left=$((total / 2 + bias))
  ((left < 0)) && left=0
  right=$((total - left))
  ((right < 0)) && right=0
  printf "%s %*s%s%*s %s\n" "$VBAR" "$left" "" "$content" "$right" "" "$VBAR"
}

box_top() { printf "%b%s%s%s%b\n" "$CYAN" "$TOPLEFT" "$HBAR" "$TOPRIGHT" "$RESET"; }
box_bottom() { printf "%b%s%s%s%b\n" "$CYAN" "$BOTLEFT" "$HBAR" "$BOTRIGHT" "$RESET"; }
box_blank() { box_row ""; }

box_section() {
  local maxlen=0 len l padding
  for l in "$@"; do
    len=$(visible_len "$l")
    ((len > maxlen)) && maxlen=$len
  done
  local pad=$(((BOX_WIDTH - maxlen) / 2))
  ((pad < 0)) && pad=0
  padding=$(printf '%*s' "$pad" '')
  for l in "$@"; do
    box_row "${padding}${l}"
  done
}

COL1_WIDTH=22

two_col_block() {
  local header="$1"
  shift
  local rows=("$@")
  local i c1 c2 combined maxlen=0 len
  local built=()
  for ((i = 0; i < ${#rows[@]}; i += 2)); do
    c1="${rows[i]}"
    c2="${rows[i + 1]}"
    combined="$(pad_field "$c1" "$COL1_WIDTH")${c2}"
    built+=("$combined")
    len=$(visible_len "$combined")
    ((len > maxlen)) && maxlen=$len
  done

  local hlen
  hlen=$(visible_len "$header")
  ((hlen > maxlen)) && maxlen=$hlen

  local pad=$(((BOX_WIDTH - maxlen) / 2))
  ((pad < 0)) && pad=0
  local padding
  padding=$(printf '%*s' "$pad" '')

  local hpad=$(((maxlen - hlen) / 2 - 1))
  ((hpad < 0)) && hpad=0
  local hpadding
  hpadding=$(printf '%*s' "$hpad" '')

  box_row "${padding}${hpadding}${header}"
  for b in "${built[@]}"; do
    box_row "${padding}${b}"
  done
}

two_col_block_wind_left() {
  local header="$1"
  shift
  local rows=("$@")
  local i c1 c2 combined maxlen=0 len
  local built=()
  for ((i = 0; i < ${#rows[@]}; i += 2)); do
    c1="${rows[i]}"
    c2="${rows[i + 1]}"
    combined="$(pad_field "$c1" "$COL1_WIDTH")${c2}"
    built+=("$combined")
    len=$(visible_len "$combined")
    ((len > maxlen)) && maxlen=$len
  done

  local hlen
  hlen=$(visible_len "$header")
  ((hlen > maxlen)) && maxlen=$hlen

  local pad=$(((BOX_WIDTH - maxlen) / 2))
  ((pad < 0)) && pad=0
  local padding
  padding=$(printf '%*s' "$pad" '')

  box_row "${padding}${header}"
  for b in "${built[@]}"; do
    box_row "${padding}${b}"
  done
}

#-------------------------------------------------------
# Fetch, parse and print the current weather
#-------------------------------------------------------
show_weather() {
  QUERY="${CITY}${COUNTRY:+,$COUNTRY}"
  URL="https://api.openweathermap.org/data/2.5/weather?q=$(printf '%s' "$QUERY" | sed 's/ /%20/g')&appid=${OWM_API_KEY}&units=metric&lang=en"

  JSON=$(curl -fsSL "$URL")

  if [[ $? -ne 0 || -z "$JSON" ]]; then
    echo "$I_ERROR Error downloading data from OpenWeatherMap."
    return 1
  fi

  COD=$(echo "$JSON" | jq -r '.cod')
  if [[ "$COD" != "200" ]]; then
    MSG=$(echo "$JSON" | jq -r '.message')
    echo "$I_ERROR API error ($COD): $MSG"
    return 1
  fi

  NAME=$(echo "$JSON" | jq -r '.name')
  COUNTRY_CODE=$(echo "$JSON" | jq -r '.sys.country')

  TNOW=$(echo "$JSON" | jq -r '.main.temp | round')
  TMAX=$(echo "$JSON" | jq -r '.main.temp_max | round')
  TMIN=$(echo "$JSON" | jq -r '.main.temp_min | round')
  FEELS=$(echo "$JSON" | jq -r '.main.feels_like | round')

  HNOW=$(echo "$JSON" | jq -r '.main.humidity')
  BNOW=$(echo "$JSON" | jq -r '.main.pressure')

  WNOW=$(echo "$JSON" | jq -r '(.wind.speed * 3.6) | round') # m/s -> km/h
  WGUST=$(echo "$JSON" | jq -r 'if .wind.gust then (.wind.gust * 3.6 | round) else "-" end')
  WDIR=$(echo "$JSON" | jq -r '.wind.deg // 0')

  RAIN=$(echo "$JSON" | jq -r 'if .rain."1h" then .rain."1h" elif .rain."3h" then .rain."3h" else 0 end')

  ICON_ID=$(echo "$JSON" | jq -r '.weather[0].id')
  DESC=$(echo "$JSON" | jq -r '.weather[0].description')

  DIRS=(
    "N" "NNE" "NE" "ENE"
    "E" "ESE" "SE" "SSE"
    "S" "SSW" "SW" "WSW"
    "W" "WNW" "NW" "NNW"
  )

  INDEX=$(((WDIR + 11) / 22))
  INDEX=$((INDEX % 16))
  DIR=${DIRS[$INDEX]}

  # Map the OpenWeatherMap condition code to an icon/label
  case "$ICON_ID" in
  2*) SKY="$I_STORM Storm" ;;
  3*) SKY="$I_DRIZZLE Drizzle" ;;
  5*) SKY="$I_RAINY Rain" ;;
  6*) SKY="$I_SNOW Snow" ;;
  7*) SKY="$I_FOG Fog" ;;
  800) SKY="$I_CLEAR Clear" ;;
  801) SKY="$I_PARTLY Slightly cloudy" ;;
  802) SKY="$I_PARTLY Partly cloudy" ;;
  803 | 804) SKY="$I_CLOUDY Cloudy" ;;
  *) SKY="$DESC" ;;
  esac

  clear

  box_top
  box_row_centered "$(printf "%b%s  OPENWEATHERMAP%b" "$CYAN" "$I_CLEAR" "$RESET")" -1
  box_row_centered "$(printf "%bhttps://openweathermap.org/%b" "$CYAN" "$RESET")"
  box_blank

  box_section \
    "$(
      printf "%s " "$I_LOCATION"
      print_color "$BLUE" 11 "City"
      printf "%b%s, %s%b" "$GREEN" "$NAME" "$COUNTRY_CODE" "$RESET"
    )" \
    "$(
      printf "%s " "$I_GLOBE"
      print_color "$WHITE" 11 "Condition"
      printf "%s" "$SKY"
    )"
  box_blank

  two_col_block \
    "$(printf "%b%s Temperature%b" "$YELLOW" "$I_THERMO" "$RESET")" \
    "$(printf "%s Now : %s °C" "$I_NOW" "$TNOW")" "$(printf "%s Max : %s °C" "$I_ARROW_UP" "$TMAX")" \
    "$(printf "%s Feels : %s °C" "$I_FEELS" "$FEELS")" "$(printf "%s Min : %s °C" "$I_ARROW_DOWN" "$TMIN")"
  box_blank

  two_col_block_wind_left \
    "$(printf "%b%s Wind%b" "$GREEN" "$I_WIND" "$RESET")" \
    "$(printf "Speed : %s km/h" "$WNOW")" "$(
      printf "%s " "$I_BAROMETER"
      print_color "$MAG" 0 "Pressure:"
    ) $BNOW hPa" \
    "$(printf "Gust  : %s km/h" "$WGUST")" "$(
      printf "%s " "$I_RAINDROP"
      print_color "$CYAN" 0 "Precipitation:"
    ) $RAIN mm" \
    "$(printf "Dir   : %s° (%s)" "$WDIR" "$DIR")" "$(
      printf "%s " "$I_HUMIDITY"
      print_color "$BLUE" 0 "Humidity:"
    ) $HNOW %"

  box_bottom

  if [[ $WATCH -eq 1 ]]; then
    printf "\n%s%s %s %s every %ss · Ctrl+C to quit%s\n" \
      "$GRAY" "$I_CLOCK" "$(date '+%H:%M:%S')" "$I_REFRESH" "$INTERVAL" "$RESET"
  fi
}

if [[ $WATCH -eq 1 ]]; then
  trap 'printf "\n"; exit 0' INT TERM
  while true; do
    show_weather
    sleep "$INTERVAL"
  done
else
  show_weather
fi
