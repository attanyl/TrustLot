package ingest

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"time"

	"github.com/trustlot/trustlot/internal/model"
)

// CSVParser handles NorthRiver-style CSV feeds.
type CSVParser struct{}

func (p *CSVParser) Parse(_ context.Context, custodian model.Custodian, data []byte) ([]model.RawRecord, error) {
	r := csv.NewReader(bytes.NewReader(data))

	// Read and validate header
	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("csv: read header: %w", err)
	}
	if len(header) == 0 {
		return nil, fmt.Errorf("csv: empty header")
	}

	var records []model.RawRecord
	seq := 0
	now := time.Now()

	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			records = append(records, model.RawRecord{
				SeqNum:   seq,
				RawData:  nil,
				ParsedAt: &now,
				Error:    fmt.Sprintf("csv: parse row %d: %v", seq, err),
			})
			seq++
			continue
		}

		// Reconstruct the raw line for storage
		var buf bytes.Buffer
		w := csv.NewWriter(&buf)
		w.Write(row)
		w.Flush()

		records = append(records, model.RawRecord{
			SeqNum:   seq,
			RawData:  buf.Bytes(),
			ParsedAt: &now,
		})
		seq++
	}

	return records, nil
}
