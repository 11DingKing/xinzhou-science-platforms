package sequence

import (
	"errors"
	"sync"
)

type Generator struct {
	mu     sync.Mutex
	prefix string
	next   int64
}

func New(prefix string, start int64) *Generator {
	if start < 1 {
		start = 1
	}
	return &Generator{prefix: prefix, next: start}
}
func (g *Generator) Next() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	value := g.prefix + "-" + itoa(g.next)
	g.next++
	return value
}
func (g *Generator) Current() int64 { g.mu.Lock(); defer g.mu.Unlock(); return g.next }
func (g *Generator) Reserve(n int64) (string, string, error) {
	if n < 1 {
		return "", "", errors.New("reserve count must be positive")
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	first := g.prefix + "-" + itoa(g.next)
	g.next += n
	last := g.prefix + "-" + itoa(g.next-1)
	return first, last, nil
}
func itoa(value int64) string {
	if value == 0 {
		return "0"
	}
	negative := value < 0
	if negative {
		value = -value
	}
	digits := []byte{}
	for value > 0 {
		digits = append(digits, byte('0'+value%10))
		value /= 10
	}
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}
	if negative {
		return "-" + string(digits)
	}
	return string(digits)
}
