package url

import (
	"net/url"
	"strings"
)

func GetDomain(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	// Get the host from the parsed URL
	host := parsedURL.Host

	// If a port is included, strip it out
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}

	return host, nil
}
