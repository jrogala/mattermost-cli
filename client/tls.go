package client

import "crypto/tls"

func tlsConfig(skipVerify bool) *tls.Config {
	return &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: skipVerify, //nolint:gosec // configurable per instance
	}
}
