package client

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
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

	timeoutStr, err := prompt(scanner, w, "Timeout in ms (default 2000)")
	if err != nil {
		return nil, err
	}
	timeout := 2000
	if timeoutStr != "" {
		parsed, err := strconv.Atoi(timeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid timeout %q: %w", timeoutStr, err)
		}
		timeout = parsed
	}

	providerChoice, err := prompt(scanner, w, "Providers to receive (all / select)")
	if err != nil {
		return nil, err
	}

	var providers []string
	if providerChoice == "select" {
		for _, p := range allProviders {
			include, err := prompt(scanner, w, fmt.Sprintf("  Include %s? (y/n)", p))
			if err != nil {
				return nil, err
			}
			if include == "y" || include == "yes" || include == "" {
				providers = append(providers, p)
			}
		}
	}

	eventChoice, err := prompt(scanner, w, "Events to receive (all / select)")
	if err != nil {
		return nil, err
	}

	var events []string
	if eventChoice == "select" {
		for _, et := range allEventTypes {
			include, err := prompt(scanner, w, fmt.Sprintf("  Include %s? (y/n)", et))
			if err != nil {
				return nil, err
			}
			if include == "y" || include == "yes" || include == "" {
				events = append(events, et)
			}
		}
	}

	c := &Client{
		Name:      name,
		URL:       url,
		Timeout:   timeout,
		Events:    events,
		Providers: providers,
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
