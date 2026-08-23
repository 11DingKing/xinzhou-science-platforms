package pagination

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type Request struct {
	Page, PageSize                  int
	Sort, Direction, Search, Region string
}
type Result[T any] struct {
	Items                 []T
	Page, PageSize, Total int
	HasNext               bool
}

func Parse(values url.Values) Request {
	page, _ := strconv.Atoi(values.Get("page"))
	size, _ := strconv.Atoi(values.Get("page_size"))
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	direction := strings.ToLower(values.Get("direction"))
	if direction != "asc" && direction != "desc" {
		direction = "asc"
	}
	return Request{Page: page, PageSize: size, Sort: values.Get("sort"), Direction: direction, Search: strings.TrimSpace(values.Get("search")), Region: strings.TrimSpace(values.Get("region"))}
}
func (r Request) Offset() int { return (r.Page - 1) * r.PageSize }
func (r Request) Validate(allowed map[string]bool) error {
	if r.Sort != "" && !allowed[r.Sort] {
		return fmt.Errorf("sort field %q is not allowed", r.Sort)
	}
	return nil
}
func BuildWhere(r Request) (string, []any) {
	clauses := []string{"1=1"}
	args := []any{}
	if r.Search != "" {
		clauses = append(clauses, "(sku LIKE ? OR display_name LIKE ?)")
		q := "%" + r.Search + "%"
		args = append(args, q, q)
	}
	if r.Region != "" {
		clauses = append(clauses, "region=?")
		args = append(args, r.Region)
	}
	return strings.Join(clauses, " AND "), args
}
func ClampTotal(total, pageSize int) int {
	if total < 0 {
		return 0
	}
	if pageSize < 1 {
		return total
	}
	return (total + pageSize - 1) / pageSize
}
func Map[T any, U any](in []T, fn func(T) U) []U {
	out := make([]U, 0, len(in))
	for _, v := range in {
		out = append(out, fn(v))
	}
	return out
}
func Empty[T any]() Result[T] { return Result[T]{Items: []T{}, Page: 1, PageSize: 20} }
