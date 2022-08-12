package mailmerge

import (
	"encoding/csv"
	"io"
	"sync"
)

// Row represent a Row in csv
type Row struct {
	mapHeaderIndex map[string]int
	records        []string
}

// GetCell get a cell value by a header key
func (r Row) GetCell(key string) (record string) {
	idx, ok := r.mapHeaderIndex[key]
	if !ok {
		return
	}
	return r.records[idx]
}

// CsvParser a csv parser
type CsvParser struct {
	mapHeaderIndex map[string]int
	rows           []Row
	sync           sync.Once
}

func (c *CsvParser) init() {
	c.sync.Do(func() {
		if c == nil {
			c = &CsvParser{}
		}
		c.mapHeaderIndex = make(map[string]int)
	})
}

func (c *CsvParser) Parse(rd io.Reader) (err error) {
	c.init()

	cr := csv.NewReader(rd)
	headers, err := cr.Read()
	if err != nil {
		return
	}

	for i, header := range headers {
		c.mapHeaderIndex[header] = i
	}

	for {
		records, err := cr.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		c.rows = append(c.rows, Row{
			mapHeaderIndex: c.mapHeaderIndex,
			records:        records,
		})
	}

	return nil
}

func (c *CsvParser) Rows() []Row {
	return c.rows
}
