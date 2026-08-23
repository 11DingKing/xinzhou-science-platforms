package config

import (
	"testing"
)

func TestConfigValidation(t *testing.T) {
	if (Config{ListenAddr: ":8080", DatabaseURL: "file:test", SessionTTLSeconds: 3600}).Validate() != nil {
		t.Fatal("valid config")
	}
	if (Config{ListenAddr: "bad", DatabaseURL: "file:test", SessionTTLSeconds: 3600}).Validate() == nil {
		t.Fatal("bad address")
	}
	if DatabaseScheme("file:test") != "file" || !IsLocal("file:test") {
		t.Fatal("scheme")
	}
	if p, err := ParsePort(":9090"); err != nil || p != 9090 {
		t.Fatal("port")
	}
}
