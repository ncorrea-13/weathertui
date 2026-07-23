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

package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ~/.config/openweather.conf
type Config struct {
	APIKey  string
	City    string
	Country string
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "openweather.conf"), nil
}

func load(path string) (Config, error) {
	var cfg Config

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		value = strings.Trim(strings.TrimSpace(value), `"`)
		switch strings.TrimSpace(key) {
		case "OWM_API_KEY":
			cfg.APIKey = value
		case "CITY":
			cfg.City = value
		case "COUNTRY":
			cfg.Country = value
		}
	}
	return cfg, scanner.Err()
}

func save(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	body := fmt.Sprintf("OWM_API_KEY=%q\nCITY=%q\nCOUNTRY=%q\n", cfg.APIKey, cfg.City, cfg.Country)
	return os.WriteFile(path, []byte(body), 0o600)
}

// API key must already be there, city/country get asked).
func Ensure() (Config, error) {
	p, err := configPath()
	if err != nil {
		return Config{}, err
	}

	cfg, err := load(p)
	if err != nil {
		return Config{}, err
	}

	reader := bufio.NewReader(os.Stdin)

	if cfg.APIKey == "" {
		fmt.Print("OpenWeatherMap API key (https://openweathermap.org/api): ")
		key, _ := reader.ReadString('\n')
		cfg.APIKey = strings.TrimSpace(key)
		if cfg.APIKey == "" {
			return cfg, errors.New("OWM_API_KEY was not found in ~/.config/openweather.conf")
		}
	}

	if cfg.City == "" {
		fmt.Println()
		fmt.Println("City to check the weather for")
		fmt.Println("Example: Mendoza")
		fmt.Println()
		fmt.Print("> City: ")
		city, _ := reader.ReadString('\n')
		cfg.City = strings.TrimSpace(city)

		fmt.Print("> Country code (e.g. AR): ")
		country, _ := reader.ReadString('\n')
		cfg.Country = strings.TrimSpace(country)
	}

	if err := save(p, cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
