package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ButtonAction define a ação configurada para um botão do mouse
type ButtonAction struct {
	Type  string `json:"type"`  // "default", "key", "macro", "dpi_cycle", "media"
	Value string `json:"value"` // Tecla, atalho ou identificador de macro
}

// Profile representa um perfil completo de configuração do mouse
type Profile struct {
	ID             string                  `json:"id"`
	Name           string                  `json:"name"`
	ActiveDPI      int                     `json:"active_dpi"`
	DPIStages      []int                   `json:"dpi_stages"`
	PollingRate    int                     `json:"polling_rate"` // 125, 250, 500, 1000 Hz
	RGBMode        string                  `json:"rgb_mode"`
	RGBColor       string                  `json:"rgb_color"` // Hex "#RRGGBB"
	RGBBrightness  int                     `json:"rgb_brightness"`
	RGBSpeed       int                     `json:"rgb_speed"`
	Buttons        map[string]ButtonAction `json:"buttons"`
}

// NextDPI avança ciclicamente para o próximo estágio de DPI e retorna o novo valor
func (p *Profile) NextDPI() int {
	if len(p.DPIStages) == 0 {
		p.DPIStages = []int{800, 1200, 1600, 2400, 3200}
	}
	currentIdx := -1
	for i, stage := range p.DPIStages {
		if stage == p.ActiveDPI {
			currentIdx = i
			break
		}
	}
	nextIdx := (currentIdx + 1) % len(p.DPIStages)
	p.ActiveDPI = p.DPIStages[nextIdx]
	return p.ActiveDPI
}

// Config gerencia as preferências persistentes do aplicativo
type Config struct {
	ActiveProfileID string             `json:"active_profile_id"`
	Profiles        map[string]Profile `json:"profiles"`
}

// DefaultConfig retorna as configurações padrão de fábrica
func DefaultConfig() *Config {
	defaultProfile := Profile{
		ID:            "default",
		Name:          "Padrão",
		ActiveDPI:     1600,
		DPIStages:     []int{800, 1600, 3200, 4800, 6000},
		PollingRate:   500,
		RGBMode:       "Rainbow",
		RGBColor:      "#FF0000",
		RGBBrightness: 100,
		RGBSpeed:      50,
		Buttons: map[string]ButtonAction{
			"button_1": {Type: "default", Value: "left_click"},
			"button_2": {Type: "default", Value: "right_click"},
			"button_3": {Type: "default", Value: "middle_click"},
			"button_4": {Type: "default", Value: "forward"},
			"button_5": {Type: "default", Value: "backward"},
			"button_6": {Type: "dpi_cycle", Value: "cycle"},
		},
	}

	return &Config{
		ActiveProfileID: "default",
		Profiles: map[string]Profile{
			"default": defaultProfile,
		},
	}
}

// GetConfigPath retorna o caminho do arquivo de configuração em ~/.config/weibo-control/config.json
func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "weibo-control")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// Load lê o arquivo de configuração ou cria o padrão caso não exista
func Load() (*Config, error) {
	path, err := GetConfigPath()
	if err != nil {
		return DefaultConfig(), err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := DefaultConfig()
			_ = Save(cfg)
			return cfg, nil
		}
		return nil, fmt.Errorf("erro ao ler arquivo de configuração: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), fmt.Errorf("erro ao decodificar JSON de configuração: %w", err)
	}

	return &cfg, nil
}

// Save persiste a configuração em disco em formato JSON identado
func Save(cfg *Config) error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao serializar configuração: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}
