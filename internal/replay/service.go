package replay

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/trustlot/trustlot/internal/db"
	"github.com/trustlot/trustlot/internal/model"
	"github.com/trustlot/trustlot/internal/recon"
)

type service struct {
	store *db.Store
}

// NewService creates a replay service backed by the given store.
func NewService(store *db.Store) Service {
	return &service{store: store}
}

func (s *service) CreateCaseFromException(ctx context.Context, exceptionID string) (*ReplayCase, error) {
	exc, err := s.store.GetException(ctx, exceptionID)
	if err != nil {
		return nil, fmt.Errorf("load exception: %w", err)
	}

	result, err := s.store.GetReconResultForException(ctx, exc.ReconRunID, exc.EntityType, exc.EntityID)
	if err != nil {
		return nil, fmt.Errorf("load recon result: %w", err)
	}

	run, err := s.store.GetReconRun(ctx, exc.ReconRunID)
	if err != nil {
		return nil, fmt.Errorf("load recon run: %w", err)
	}

	internal, custodian, err := s.loadRecordPair(ctx, exc, run)
	if err != nil {
		return nil, fmt.Errorf("load record pair: %w", err)
	}

	internalJSON, err := json.Marshal(internal)
	if err != nil {
		return nil, fmt.Errorf("marshal internal snapshot: %w", err)
	}
	custodianJSON, err := json.Marshal(custodian)
	if err != nil {
		return nil, fmt.Errorf("marshal custodian snapshot: %w", err)
	}

	cfg := recon.DefaultConfig()
	configJSON, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("marshal config snapshot: %w", err)
	}

	caseID := generateID()
	now := time.Now()
	excID := exc.ID

	row := db.ReplayCaseRow{
		ID:                  caseID,
		Description:         fmt.Sprintf("Captured from exception %s: %s on %s", exc.ID, exc.ReasonCode, exc.EntityType),
		Status:              "pending",
		SourceExceptionID:   &excID,
		EntityType:          exc.EntityType,
		ExpectedMatchStatus: string(result.Status),
		ExpectedReasonCode:  string(result.ReasonCode),
		InternalSnapshot:    internalJSON,
		CustodianSnapshot:   custodianJSON,
		ConfigSnapshot:      configJSON,
		CreatedAt:           now,
	}

	if err := s.store.InsertReplayCase(ctx, row); err != nil {
		return nil, fmt.Errorf("persist replay case: %w", err)
	}

	slog.Info("replay case created", "case_id", caseID, "exception_id", exceptionID, "entity_type", exc.EntityType)

	rc := rowToCase(row)
	return &rc, nil
}

func (s *service) RunCase(ctx context.Context, replayCaseID string) (*ReplayRunResult, error) {
	row, err := s.store.GetReplayCase(ctx, replayCaseID)
	if err != nil {
		return nil, fmt.Errorf("load replay case: %w", err)
	}

	actualStatus, actualRC, err := replayFromSnapshots(row)
	if err != nil {
		return nil, fmt.Errorf("replay execution: %w", err)
	}

	actualRCStr := ""
	if actualRC != nil {
		actualRCStr = string(*actualRC)
	}

	changed := string(actualStatus) != row.ExpectedMatchStatus || actualRCStr != row.ExpectedReasonCode

	// Determine improvement vs regression.
	// Improvement: was a break, now is a match or near-match.
	// Regression: was a match/near-match, now is a break.
	improvement := false
	regression := false
	if changed {
		wasBreak := row.ExpectedMatchStatus == string(model.StatusBreak)
		isBreak := actualStatus == model.StatusBreak
		if wasBreak && !isBreak {
			improvement = true
		} else if !wasBreak && isBreak {
			regression = true
		}
	}

	summary := buildSummary(changed, improvement, regression, row, actualStatus, actualRCStr)

	// Update case status based on outcome.
	newStatus := "passing"
	if regression {
		newStatus = "failing"
	}
	if err := s.store.UpdateReplayCaseStatus(ctx, replayCaseID, newStatus); err != nil {
		slog.Warn("failed to update replay case status", "case_id", replayCaseID, "error", err)
	}

	slog.Info("replay case executed", "case_id", replayCaseID, "changed", changed, "improvement", improvement, "regression", regression)

	return &ReplayRunResult{
		CaseID:              replayCaseID,
		ActualMatchStatus:   string(actualStatus),
		ActualReasonCode:    actualRCStr,
		ExpectedMatchStatus: row.ExpectedMatchStatus,
		ExpectedReasonCode:  row.ExpectedReasonCode,
		Changed:             changed,
		Regression:          regression,
		Improvement:         improvement,
		Summary:             summary,
	}, nil
}

func (s *service) ListCases(ctx context.Context) ([]ReplayCase, error) {
	rows, err := s.store.ListReplayCases(ctx)
	if err != nil {
		return nil, fmt.Errorf("list replay cases: %w", err)
	}
	out := make([]ReplayCase, len(rows))
	for i, row := range rows {
		out[i] = rowToCase(row)
	}
	return out, nil
}

func (s *service) GetCase(ctx context.Context, replayCaseID string) (*ReplayCaseDetail, error) {
	row, err := s.store.GetReplayCase(ctx, replayCaseID)
	if err != nil {
		return nil, fmt.Errorf("load replay case: %w", err)
	}

	detail := ReplayCaseDetail{
		ReplayCase: rowToCase(row),
	}

	// Unmarshal snapshots into generic maps for JSON display.
	if row.InternalSnapshot != nil {
		var v any
		if json.Unmarshal(row.InternalSnapshot, &v) == nil {
			detail.InternalSnapshot = v
		}
	}
	if row.CustodianSnapshot != nil {
		var v any
		if json.Unmarshal(row.CustodianSnapshot, &v) == nil {
			detail.CustodianSnapshot = v
		}
	}
	if row.ConfigSnapshot != nil {
		var v any
		if json.Unmarshal(row.ConfigSnapshot, &v) == nil {
			detail.ConfigSnapshot = v
		}
	}

	return &detail, nil
}

// replayFromSnapshots deserializes stored snapshots and runs current rules
// against them. This is the core deterministic replay path — it never touches
// the live database.
func replayFromSnapshots(row db.ReplayCaseRow) (model.MatchStatus, *model.ReasonCode, error) {
	if row.InternalSnapshot == nil || row.CustodianSnapshot == nil {
		return model.StatusBreak, nil, fmt.Errorf("incomplete snapshots for case %s", row.ID)
	}

	var cfg recon.ReconConfig
	if row.ConfigSnapshot != nil {
		if err := json.Unmarshal(row.ConfigSnapshot, &cfg); err != nil {
			return "", nil, fmt.Errorf("unmarshal config: %w", err)
		}
	} else {
		cfg = recon.DefaultConfig()
	}

	switch row.EntityType {
	case "position":
		return replayPosition(row.InternalSnapshot, row.CustodianSnapshot, cfg)
	case "transaction":
		return replayTransaction(row.InternalSnapshot, row.CustodianSnapshot, cfg)
	default:
		return "", nil, fmt.Errorf("unsupported entity type: %s", row.EntityType)
	}
}

func replayPosition(internalJSON, custodianJSON []byte, cfg recon.ReconConfig) (model.MatchStatus, *model.ReasonCode, error) {
	var internal, custodian model.Position
	if err := json.Unmarshal(internalJSON, &internal); err != nil {
		return "", nil, fmt.Errorf("unmarshal internal position: %w", err)
	}
	if err := json.Unmarshal(custodianJSON, &custodian); err != nil {
		return "", nil, fmt.Errorf("unmarshal custodian position: %w", err)
	}

	rules := recon.PositionRules()
	status, rc, _ := recon.ApplyRules(rules, recon.RuleContext{
		Internal:  internal,
		Custodian: custodian,
		Config:    cfg,
	})
	return status, rc, nil
}

func replayTransaction(internalJSON, custodianJSON []byte, cfg recon.ReconConfig) (model.MatchStatus, *model.ReasonCode, error) {
	var internal, custodian model.Transaction
	if err := json.Unmarshal(internalJSON, &internal); err != nil {
		return "", nil, fmt.Errorf("unmarshal internal transaction: %w", err)
	}
	if err := json.Unmarshal(custodianJSON, &custodian); err != nil {
		return "", nil, fmt.Errorf("unmarshal custodian transaction: %w", err)
	}

	rules := recon.TransactionRules()
	status, rc, _ := recon.ApplyRules(rules, recon.RuleContext{
		Internal:  internal,
		Custodian: custodian,
		Config:    cfg,
	})
	return status, rc, nil
}

func buildSummary(changed, improvement, regression bool, row db.ReplayCaseRow, actualStatus model.MatchStatus, actualRC string) string {
	if !changed {
		return "Replay matched original outcome."
	}
	if improvement {
		return fmt.Sprintf("Replay improved: %s/%s became %s.",
			row.ExpectedMatchStatus, row.ExpectedReasonCode, actualStatus)
	}
	if regression {
		detail := string(actualStatus)
		if actualRC != "" {
			detail += "/" + actualRC
		}
		return fmt.Sprintf("Replay regressed: %s became %s.",
			row.ExpectedMatchStatus, detail)
	}
	// Changed but same severity (e.g., different reason code but still BREAK).
	detail := string(actualStatus)
	if actualRC != "" {
		detail += "/" + actualRC
	}
	return fmt.Sprintf("Replay changed: %s/%s became %s.",
		row.ExpectedMatchStatus, row.ExpectedReasonCode, detail)
}

// loadRecordPair mirrors the explain service pattern: load entity by ID, find pair.
func (s *service) loadRecordPair(ctx context.Context, exc model.Exception, run model.ReconRun) (any, any, error) {
	switch exc.EntityType {
	case "position":
		return s.loadPositionPair(ctx, exc.EntityID, run)
	case "transaction":
		return s.loadTransactionPair(ctx, exc.EntityID, run)
	default:
		return nil, nil, fmt.Errorf("unsupported entity type: %s", exc.EntityType)
	}
}

func (s *service) loadPositionPair(ctx context.Context, entityID string, run model.ReconRun) (any, any, error) {
	pos, err := s.store.GetPosition(ctx, entityID)
	if err != nil {
		return nil, nil, fmt.Errorf("load position %s: %w", entityID, err)
	}

	oppositeSource := "custodian"
	if pos.Source == "custodian" {
		oppositeSource = "internal"
	}

	paired, err := s.store.FindPairedPosition(ctx, pos.AccountID, pos.SecurityID, run.AsOfDate, oppositeSource)
	if err != nil {
		// Unmatched record — store what we have.
		if pos.Source == "internal" {
			return pos, nil, nil
		}
		return nil, pos, nil
	}

	if pos.Source == "internal" {
		return pos, paired, nil
	}
	return paired, pos, nil
}

func (s *service) loadTransactionPair(ctx context.Context, entityID string, run model.ReconRun) (any, any, error) {
	txn, err := s.store.GetTransaction(ctx, entityID)
	if err != nil {
		return nil, nil, fmt.Errorf("load transaction %s: %w", entityID, err)
	}

	oppositeSource := "custodian"
	if txn.Source == "custodian" {
		oppositeSource = "internal"
	}

	paired, err := s.store.FindPairedTransaction(ctx, txn.AccountID, txn.SecurityID, txn.TxnType, txn.TradeDate, oppositeSource)
	if err != nil {
		if txn.Source == "internal" {
			return txn, nil, nil
		}
		return nil, txn, nil
	}

	if txn.Source == "internal" {
		return txn, paired, nil
	}
	return paired, txn, nil
}

func rowToCase(row db.ReplayCaseRow) ReplayCase {
	return ReplayCase{
		ID:                  row.ID,
		Description:         row.Description,
		Status:              row.Status,
		SourceExceptionID:   row.SourceExceptionID,
		EntityType:          row.EntityType,
		ExpectedMatchStatus: row.ExpectedMatchStatus,
		ExpectedReasonCode:  row.ExpectedReasonCode,
		CreatedAt:           row.CreatedAt.Format(time.RFC3339),
	}
}

func generateID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hex.EncodeToString(b[0:4]),
		hex.EncodeToString(b[4:6]),
		hex.EncodeToString(b[6:8]),
		hex.EncodeToString(b[8:10]),
		hex.EncodeToString(b[10:16]),
	)
}
