package client

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"
)

var allEventTypes = []string{"session_start", "stop", "notification"}

func RunWizard(r io.Reader, w io.Writer) (*Client, error) {
	scanner := bufio.NewScanner(r)

	name, err := prompt(scanner, w, "Nombre del client")
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, fmt.Errorf("client name is required")
	}

	rawURL, err := prompt(scanner, w, "URL o IP")
	if err != nil {
		return nil, err
	}
	if rawURL == "" {
		return nil, fmt.Errorf("client URL is required")
	}

	portStr, err := prompt(scanner, w, "Puerto (default 80)")
	if err != nil {
		return nil, err
	}

	url := rawURL
	if !strings.Contains(url, "://") {
		url = "http://" + url
	}
	if portStr != "" && portStr != "80" {
		url = url + ":" + portStr
	}

	timeoutStr, err := prompt(scanner, w, "Timeout (default 2s)")
	if err != nil {
		return nil, err
	}
	timeout := 2 * time.Second
	if timeoutStr != "" {
		parsed, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid timeout %q: %w", timeoutStr, err)
		}
		timeout = parsed
	}

	eventChoice, err := prompt(scanner, w, "Eventos a recibir (todos / seleccionar)")
	if err != nil {
		return nil, err
	}

	var events []string
	if eventChoice == "seleccionar" || eventChoice == "select" {
		for _, et := range allEventTypes {
			include, err := prompt(scanner, w, fmt.Sprintf("  Incluir %s? (s/n)", et))
			if err != nil {
				return nil, err
			}
			if include == "s" || include == "si" || include == "y" || include == "yes" || include == "" {
				events = append(events, et)
			}
		}
	}

	c := &Client{
		Name:    name,
		URL:     url,
		Timeout: timeout,
		Events:  events,
	}

	if err := c.Validate(); err != nil {
		return nil, err
	}

	return c, nil
}

func prompt(scanner *bufio.Scanner, w io.Writer, label string) (string, error) {
	fmt.Fprintf(w, "? %s: ", label)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", err
		}
		return "", io.EOF
	}
	return strings.TrimSpace(scanner.Text()), nil
}
