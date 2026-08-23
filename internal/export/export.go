package export

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"
)

type Record struct {
	ID        int64     `json:"id"`
	SKU       string    `json:"sku"`
	BatchCode string    `json:"batch_code"`
	Region    string    `json:"region"`
	Status    string    `json:"status"`
	Score     int       `json:"score"`
	At        time.Time `json:"at"`
}

func JSON(w io.Writer, records []Record) error { return json.NewEncoder(w).Encode(records) }
func CSV(w io.Writer, records []Record) error {
	writer := csv.NewWriter(w)
	if err := writer.Write([]string{"id", "sku", "batch_code", "region", "status", "score", "at"}); err != nil {
		return err
	}
	for _, r := range records {
		if err := writer.Write([]string{strconv.FormatInt(r.ID, 10), r.SKU, r.BatchCode, r.Region, r.Status, strconv.Itoa(r.Score), r.At.UTC().Format(time.RFC3339)}); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}
func Lines(w io.Writer, records []Record) error {
	bw := bufio.NewWriter(w)
	for _, r := range records {
		if _, err := fmt.Fprintf(bw, "%d %s %s %s %s %d\n", r.ID, r.SKU, r.BatchCode, r.Region, r.Status, r.Score); err != nil {
			return err
		}
	}
	return bw.Flush()
}
func Filter(records []Record, status string, minScore int) []Record {
	out := []Record{}
	for _, r := range records {
		if status != "" && r.Status != status {
			continue
		}
		if r.Score < minScore {
			continue
		}
		out = append(out, r)
	}
	return out
}
