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

// Map transform row into a map
func (r Row) Map() map[string]interface{} {
	rmap := make(map[string]interface{}, len(r.records))
	for header := range r.mapHeaderIndex {
		rmap[header] = r.GetCell(header)
	}
	return rmap
}

// Csv a csv parser
type Csv struct {
	mapHeaderIndex map[string]int
	rows           []Row
	sync           sync.Once
}

func (c *Csv) init() {
	c.sync.Do(func() {
		if c == nil {
			c = &Csv{}
		}
		c.mapHeaderIndex = make(map[string]int)
	})
}

// IsHeader check if the header is exists
func (c *Csv) IsHeader(header string) bool {
	_, ok := c.mapHeaderIndex[header]
	return ok
}

func (c *Csv) Parse(rd io.Reader) (err error) {
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

func (c *Csv) Rows() []Row {
	return c.rows
}
