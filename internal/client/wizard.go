package client

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

var allEventTypes = []string{"session_start", "stop", "notification"}
var allProviders = []string{"claude", "gemini"}

func RunWizard(r io.Reader, w io.Writer) (*Client, error) {
	scanner := bufio.NewScanner(r)

	name, err := prompt(scanner, w, "Client name")
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, fmt.Errorf("client name is required")
	}

	rawURL, err := prompt(scanner, w, "URL or IP")
	if err != nil {
		return nil, err
	}
	if rawURL == "" {
		return nil, fmt.Errorf("client URL is required")
	}

	portStr, err := prompt(scanner, w, "Port (default 80)")
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

	fmt.Fprintln(w, "  Comma-separated values accepted for providers and events.")

	providerChoice, err := prompt(scanner, w, fmt.Sprintf("Providers (%s or all)", strings.Join(allProviders, ",")))
	if err != nil {
		return nil, err
	}

	providers := parseChoice(providerChoice)

	eventChoice, err := prompt(scanner, w, fmt.Sprintf("Events (%s or all)", strings.Join(allEventTypes, ",")))
	if err != nil {
		return nil, err
	}

	events := parseChoice(eventChoice)

	c := &Client{
		Name:      name,
		URL:       url,
		Events:    events,
		Providers: providers,
	}

	if err := c.Validate(); err != nil {
		return nil, err
	}

	return c, nil
}

func parseChoice(input string) []string {
	if input == "" || input == "all" {
		return nil
	}
	parts := strings.Split(input, ",")
	var result []string
	for _, p := range parts {
		result = append(result, strings.TrimSpace(p))
	}
	return result
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
