package ingest

import (
	"context"
	"testing"

	"github.com/trustlot/trustlot/internal/model"
)

func TestCSVParserValidInput(t *testing.T) {
	input := []byte("account_id,security,quantity,price\nNR-100201,AAPL,150,175.00\nNR-100201,MSFT,500,420.00\n")

	parser := &CSVParser{}
	custodian := model.Custodian{Name: "NorthRiver", Format: "csv"}

	records, err := parser.Parse(context.Background(), custodian, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	for i, rec := range records {
		if rec.SeqNum != i {
			t.Errorf("record %d: expected seq %d, got %d", i, i, rec.SeqNum)
		}
		if rec.Error != "" {
			t.Errorf("record %d: unexpected error: %s", i, rec.Error)
		}
		if rec.ParsedAt == nil {
			t.Errorf("record %d: ParsedAt should be set", i)
		}
		if len(rec.RawData) == 0 {
			t.Errorf("record %d: RawData should not be empty", i)
		}
	}
}

func TestCSVParserEmptyInput(t *testing.T) {
	parser := &CSVParser{}
	custodian := model.Custodian{Name: "NorthRiver", Format: "csv"}

	_, err := parser.Parse(context.Background(), custodian, []byte(""))
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestCSVParserHeaderOnly(t *testing.T) {
	input := []byte("account_id,security,quantity,price\n")

	parser := &CSVParser{}
	custodian := model.Custodian{Name: "NorthRiver", Format: "csv"}

	records, err := parser.Parse(context.Background(), custodian, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(records) != 0 {
		t.Errorf("expected 0 records for header-only input, got %d", len(records))
	}
}
