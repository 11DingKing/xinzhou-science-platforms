package contract

import (
	"testing"
	"time"
)

func TestContractValidation(t *testing.T) {
	c := Contract{Name: "inspection", Version: "v1", Fields: []Field{{Name: "sku", Required: true, MaxLength: 5}, {Name: "note", MaxLength: 10}}}
	if c.Valid() != nil {
		t.Fatal("contract")
	}
	if len(c.Validate(map[string]string{"note": "12345678901"})) != 2 {
		t.Fatal("violations")
	}
	if len(RequiredNames(c)) != 1 || !Expired(time.Now().Add(-time.Hour), time.Now(), time.Minute) {
		t.Fatal("helpers")
	}
}
