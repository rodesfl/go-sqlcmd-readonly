package sqlcmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// ndjsonFormatterType implements Formatter and writes results as NDJSON.
// It emits:
//   - one schema object per result set
//   - one row object per row
//   - a single summary object at the end of the batch
type ndjsonFormatterType struct {
	out   io.Writer
	vars  *Variables
	cols  []string

	// summary information, populated via SetSummary
	rowCount int64
	limited  bool
	limit    int64
}

type ndjsonSchema struct {
	Type    string   `json:"type"`
	Columns []string `json:"columns"`
}

type ndjsonRow struct {
	Type   string   `json:"type"`
	Values []string `json:"values"`
}

type ndjsonSummary struct {
	Type     string `json:"type"`
	RowCount int64  `json:"row_count"`
	Limited  bool   `json:"limited"`
	Limit    int64  `json:"limit"`
}

// NewNDJSONFormatter returns a Formatter that writes NDJSON to the provided writer.
func NewNDJSONFormatter() Formatter {
	return &ndjsonFormatterType{}
}

func (f *ndjsonFormatterType) BeginBatch(_ string, vars *Variables, out io.Writer, _ io.Writer) {
	f.out = out
	f.vars = vars
}

func (f *ndjsonFormatterType) EndBatch() {
	if f.out == nil {
		return
	}
	enc := json.NewEncoder(f.out)
	_ = enc.Encode(ndjsonSummary{
		Type:     "summary",
		RowCount: f.rowCount,
		Limited:  f.limited,
		Limit:    f.limit,
	})
}

func (f *ndjsonFormatterType) BeginResultSet(cols []*sql.ColumnType) {
	f.cols = make([]string, len(cols))
	for i, c := range cols {
		f.cols[i] = c.Name()
	}
	if f.out == nil {
		return
	}
	enc := json.NewEncoder(f.out)
	_ = enc.Encode(ndjsonSchema{
		Type:    "schema",
		Columns: f.cols,
	})
}

func (f *ndjsonFormatterType) EndResultSet() {
	// no-op for NDJSON; separation between result sets is implicit
}

func (f *ndjsonFormatterType) AddRow(row *sql.Rows) string {
	if f.out == nil {
		return ""
	}
	cols, err := row.Columns()
	if err != nil {
		return ""
	}

	r := make([]interface{}, len(cols))
	for i := range r {
		r[i] = new(interface{})
	}
	if err := row.Scan(r...); err != nil {
		return ""
	}

	values := make([]string, len(cols))
	for i, v := range r {
		j := v.(*interface{})
		if *j == nil {
			values[i] = "NULL"
			continue
		}
		switch x := (*j).(type) {
		case []byte:
			values[i] = string(x)
		case string:
			values[i] = x
		case time.Time:
			values[i] = x.Format(time.RFC3339Nano)
		case fmt.Stringer:
			values[i] = x.String()
		case bool:
			if x {
				values[i] = "1"
			} else {
				values[i] = "0"
			}
		default:
			values[i] = fmt.Sprintf("%v", x)
		}
	}

	enc := json.NewEncoder(f.out)
	_ = enc.Encode(ndjsonRow{
		Type:   "row",
		Values: values,
	})

	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (f *ndjsonFormatterType) AddMessage(string) {
	// Messages are not emitted in NDJSON mode; they will still be handled by existing paths.
}

func (f *ndjsonFormatterType) AddError(error) {
	// Errors are not emitted in NDJSON mode; they will still be handled by existing paths.
}

func (f *ndjsonFormatterType) XmlMode(bool) {
	// XML mode is not applicable for NDJSON; ignore.
}

func (f *ndjsonFormatterType) IsXmlMode() bool {
	return false
}

// SetSummary allows runQuery to provide row count and limit details for the summary.
func (f *ndjsonFormatterType) SetSummary(rowCount int64, limited bool, limit int64) {
	f.rowCount = rowCount
	f.limited = limited
	f.limit = limit
}

