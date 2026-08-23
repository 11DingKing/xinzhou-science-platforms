package config

import (
	"errors"
	"net"
	"net/url"
	"strconv"
	"strings"
)

var ErrInvalidConfig = errors.New("invalid configuration")

func (c Config) Validate() error {
	if strings.TrimSpace(c.ListenAddr) == "" || strings.TrimSpace(c.DatabaseURL) == "" || c.SessionTTLSeconds < 60 {
		return ErrInvalidConfig
	}
	if _, _, err := net.SplitHostPort(c.ListenAddr); err != nil && !strings.HasPrefix(c.ListenAddr, ":") {
		return ErrInvalidConfig
	}
	return nil
}
func ParsePort(addr string) (int, error) {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(port)
}
func DatabaseScheme(dsn string) string {
	if u, err := url.Parse(dsn); err == nil && u.Scheme != "" {
		return u.Scheme
	}
	if strings.HasPrefix(dsn, "file:") {
		return "file"
	}
	return ""
}
func IsLocal(dsn string) bool { return DatabaseScheme(dsn) == "file" }
