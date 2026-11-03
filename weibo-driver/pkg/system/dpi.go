package system

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// BaseDPI é o valor nominal de referência do sensor (1200 DPI = escala 1.0)
const BaseDPI = 1200

// FindMouseDeviceID busca o ID do dispositivo XInput do mouse Weibo
func FindMouseDeviceID() (int, error) {
	// Tenta busca direta por nome específico de ponteiro
	cmd := exec.Command("xinput", "list", "--id-only", "pointer:YICHIP 2.4G Receiver Mouse")
	out, err := cmd.Output()
	if err == nil {
		idStr := strings.TrimSpace(string(out))
		if id, err := strconv.Atoi(idStr); err == nil {
			return id, nil
		}
	}

	// Fallback: listar e buscar por regex
	cmd = exec.Command("xinput", "list")
	out, err = cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("xinput list falhou: %w", err)
	}

	lines := strings.Split(string(out), "\n")
	re := regexp.MustCompile(`YICHIP.*Receiver.*Mouse.*id=(\d+).*pointer`)
	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if len(matches) > 1 {
			id, err := strconv.Atoi(matches[1])
			if err == nil {
				return id, nil
			}
		}
	}

	return 0, fmt.Errorf("mouse Weibo não encontrado no xinput")
}

// DPIToScale converte um valor de DPI nominal em multiplicador de matriz
func DPIToScale(dpi int) float64 {
	if dpi <= 0 {
		return 1.0
	}
	// 800 DPI -> ~0.67
	// 1200 DPI -> 1.0
	// 1600 DPI -> 1.33
	// 2400 DPI -> 2.0
	// 3200 DPI -> 2.67
	return float64(dpi) / float64(BaseDPI)
}

// ScaleToDPI converte uma escala em DPI nominal mais próximo
func ScaleToDPI(scale float64) int {
	return int(scale*float64(BaseDPI) + 0.5)
}

// ApplySensitivity aplica o multiplicador de DPI real na matriz de transformação de coordenadas
func ApplySensitivity(dpi int) error {
	id, err := FindMouseDeviceID()
	if err != nil {
		return err
	}

	scale := DPIToScale(dpi)
	return SetCoordinateScale(id, scale)
}

// SetCoordinateScale aplica a escala diretamente no XInput
func SetCoordinateScale(deviceID int, scale float64) error {
	s := fmt.Sprintf("%.4f", scale)
	// Matriz de Transformação:
	// [ s  0  0 ]
	// [ 0  s  0 ]
	// [ 0  0  1 ]
	cmd := exec.Command("xinput", "set-prop", strconv.Itoa(deviceID),
		"Coordinate Transformation Matrix",
		s, "0", "0",
		"0", s, "0",
		"0", "0", "1",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("falha ao aplicar matriz xinput: %s (%w)", string(output), err)
	}

	return nil
}

// GetCurrentSensitivity lê a escala ativa no XInput
func GetCurrentSensitivity() (float64, int, error) {
	id, err := FindMouseDeviceID()
	if err != nil {
		return 1.0, BaseDPI, err
	}

	cmd := exec.Command("xinput", "list-props", strconv.Itoa(deviceID(id)))
	out, err := cmd.Output()
	if err != nil {
		return 1.0, BaseDPI, err
	}

	re := regexp.MustCompile(`Coordinate Transformation Matrix.*:\s+([0-9\.]+),`)
	matches := re.FindStringSubmatch(string(out))
	if len(matches) > 1 {
		scale, err := strconv.ParseFloat(matches[1], 64)
		if err == nil {
			return scale, ScaleToDPI(scale), nil
		}
	}

	return 1.0, BaseDPI, nil
}

func deviceID(id int) int {
	return id
}
