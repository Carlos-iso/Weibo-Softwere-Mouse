package system

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// DefaultButtonMap representa o mapeamento padrão de 9 botões do mouse
var DefaultButtonMap = []int{1, 2, 3, 4, 5, 6, 7, 8, 9}

// SetButtonMap configura a ordem física dos botões no X11
func SetButtonMap(mapValues []int) error {
	id, err := FindMouseDeviceID()
	if err != nil {
		return err
	}

	var args []string
	args = append(args, "set-button-map", strconv.Itoa(id))
	for _, v := range mapValues {
		args = append(args, strconv.Itoa(v))
	}

	cmd := exec.Command("xinput", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("falha ao aplicar button map: %s (%w)", string(out), err)
	}

	return nil
}

// GetButtonMap retorna a lista atual de botões configurados
func GetButtonMap() ([]int, error) {
	id, err := FindMouseDeviceID()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("xinput", "get-button-map", strconv.Itoa(id))
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	fields := strings.Fields(string(out))
	var res []int
	for _, f := range fields {
		v, err := strconv.Atoi(f)
		if err == nil {
			res = append(res, v)
		}
	}

	return res, nil
}

// ButtonHandler define a assinatura de uma função executada quando um botão é clicado
type ButtonHandler func(button int, state string)

// ButtonListener monitora cliques em segundo plano via xinput test
type ButtonListener struct {
	deviceID int
	cmd      *exec.Cmd
	handlers map[int]ButtonHandler
	mu       sync.Mutex
	running  bool
}

// NewButtonListener cria um ouvinte para interceptar eventos de botões
func NewButtonListener() (*ButtonListener, error) {
	id, err := FindMouseDeviceID()
	if err != nil {
		return nil, err
	}

	return &ButtonListener{
		deviceID: id,
		handlers: make(map[int]ButtonHandler),
	}, nil
}

// RegisterHandler associa uma ação a um botão específico (ex: 8 para Back, 9 para Forward)
func (bl *ButtonListener) RegisterHandler(button int, handler ButtonHandler) {
	bl.mu.Lock()
	defer bl.mu.Unlock()
	bl.handlers[button] = handler
}

// Start inicia o processo em background
func (bl *ButtonListener) Start() error {
	bl.mu.Lock()
	if bl.running {
		bl.mu.Unlock()
		return nil
	}
	bl.running = true
	bl.mu.Unlock()

	bl.cmd = exec.Command("xinput", "test", strconv.Itoa(bl.deviceID))
	stdout, err := bl.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := bl.cmd.Start(); err != nil {
		return err
	}

	go func() {
		scanner := bufio.NewScanner(stdout)
		// Exemplo de saída: "button press   8" ou "button release 8"
		re := regexp.MustCompile(`button\s+(press|release)\s+(\d+)`)
		for scanner.Scan() {
			text := scanner.Text()
			matches := re.FindStringSubmatch(text)
			if len(matches) > 2 {
				state := matches[1]
				btn, _ := strconv.Atoi(matches[2])

				bl.mu.Lock()
				handler, exists := bl.handlers[btn]
				bl.mu.Unlock()

				if exists {
					handler(btn, state)
				}
			}
		}
	}()

	return nil
}

// Stop encerra o ouvinte
func (bl *ButtonListener) Stop() error {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	if !bl.running || bl.cmd == nil || bl.cmd.Process == nil {
		return nil
	}

	bl.running = false
	return bl.cmd.Process.Kill()
}
