package mailmerger

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

// CSV a csv parser
type CSV struct {
	mapHeaderIndex map[string]int
	rows           []Row
	sync           sync.Once
}

func (c *CSV) init() {
	c.sync.Do(func() {
		if c == nil {
			c = &CSV{}
		}
		c.mapHeaderIndex = make(map[string]int)
	})
}

// IsHeader check if the header is exists
func (c *CSV) IsHeader(header string) bool {
	_, ok := c.mapHeaderIndex[header]
	return ok
}

func (c *CSV) Parse(rd io.Reader) (err error) {
	c.init()

	cr := csv.NewReader(rd)
	headers, err := cr.Read()
	if err != nil {
		return
	}

	for i, header := range headers {
		if header == "" {
			continue
		}
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

func (c *CSV) Rows() []Row {
	return c.rows
}
