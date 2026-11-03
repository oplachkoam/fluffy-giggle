package run

import (
	"encoding/json"
	"fmt"
	"time"
	"uuid"

	"fluffy-giggle/core/internal/domain/entity"
	"fluffy-giggle/core/internal/domain/vo"
	db "fluffy-giggle/core/internal/storage/sqlite/sqlc"
)

// row is the common shape of every runs query; listings carry the image size
// instead of the bytes.
type row struct {
	ID, CreatedAt, UpdatedAt, Status, Phase, AdvisorMode, AdvisorOutcome, ReviewDecision string
	Source, CaseID, BenchmarkID, InputText                                               string
	Image                                                                                []byte
	ImageSize                                                                            int
	ImageMime, CanariesJson, AnonymizedText, VaultJson, FindingsJson, CardJson, Draft    string
	AdviceJson, FinalAnonymized, FinalText                                               string
	JudgeScore                                                                           *int64
	JudgeReason                                                                          string
	Gate2Passes, Escalated, CanariesLeaked                                               int64
	Error                                                                                string
}

func fromRun(r db.Run) row {
	return row{
		ID: r.ID, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt, Status: r.Status, Phase: r.Phase,
		AdvisorMode: r.AdvisorMode, AdvisorOutcome: r.AdvisorOutcome, ReviewDecision: r.ReviewDecision,
		Source: r.Source, CaseID: r.CaseID, BenchmarkID: r.BenchmarkID, InputText: r.InputText,
		Image: r.Image, ImageSize: len(r.Image), ImageMime: r.ImageMime, CanariesJson: r.CanariesJson,
		AnonymizedText: r.AnonymizedText, VaultJson: r.VaultJson, FindingsJson: r.FindingsJson,
		CardJson: r.CardJson, Draft: r.Draft, AdviceJson: r.AdviceJson,
		FinalAnonymized: r.FinalAnonymized, FinalText: r.FinalText, JudgeScore: r.JudgeScore,
		JudgeReason: r.JudgeReason, Gate2Passes: r.Gate2Passes, Escalated: r.Escalated,
		CanariesLeaked: r.CanariesLeaked, Error: r.Error,
	}
}

func fromListRow(r db.ListRunsRow) row {
	return fromListRunsByBenchmarkRow(db.ListRunsByBenchmarkRow(r))
}

func fromListRunsByBenchmarkRow(r db.ListRunsByBenchmarkRow) row {
	var size int
	if r.ImageSize != nil {
		size = int(*r.ImageSize)
	}
	return row{
		ID: r.ID, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt, Status: r.Status, Phase: r.Phase,
		AdvisorMode: r.AdvisorMode, AdvisorOutcome: r.AdvisorOutcome, ReviewDecision: r.ReviewDecision,
		Source: r.Source, CaseID: r.CaseID, BenchmarkID: r.BenchmarkID, InputText: r.InputText,
		ImageSize: size, ImageMime: r.ImageMime, CanariesJson: r.CanariesJson,
		AnonymizedText: r.AnonymizedText, VaultJson: r.VaultJson, FindingsJson: r.FindingsJson,
		CardJson: r.CardJson, Draft: r.Draft, AdviceJson: r.AdviceJson,
		FinalAnonymized: r.FinalAnonymized, FinalText: r.FinalText, JudgeScore: r.JudgeScore,
		JudgeReason: r.JudgeReason, Gate2Passes: r.Gate2Passes, Escalated: r.Escalated,
		CanariesLeaked: r.CanariesLeaked, Error: r.Error,
	}
}

func toEntity(r row) (*entity.Run, error) {
	id, err := uuid.Parse(r.ID)
	if err != nil {
		return nil, fmt.Errorf("parse run id: %w", err)
	}
	benchmarkID := uuid.Nil()
	if r.BenchmarkID != "" {
		if benchmarkID, err = uuid.Parse(r.BenchmarkID); err != nil {
			return nil, fmt.Errorf("parse benchmark id: %w", err)
		}
	}
	createdAt, err := parseTime(r.CreatedAt)
	if err != nil {
		return nil, err
	}
	updatedAt, err := parseTime(r.UpdatedAt)
	if err != nil {
		return nil, err
	}
	var canaries, advice []string
	var vault []vo.VaultEntry
	var findings []vo.Span
	var card *vo.Card
	for name, pair := range map[string]struct {
		raw string
		dst any
	}{
		"canaries": {r.CanariesJson, &canaries},
		"advice":   {r.AdviceJson, &advice},
		"vault":    {r.VaultJson, &vault},
		"findings": {r.FindingsJson, &findings},
		"card":     {r.CardJson, &card},
	} {
		if pair.raw == "" {
			continue
		}
		if err := json.Unmarshal([]byte(pair.raw), pair.dst); err != nil {
			return nil, fmt.Errorf("decode run %s: %w", name, err)
		}
	}
	var score int
	if r.JudgeScore != nil {
		score = int(*r.JudgeScore)
	}

	return entity.RestoreRun(&entity.RunState{
		ID:             id,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
		Status:         vo.RunStatus(r.Status),
		Phase:          vo.RunPhase(r.Phase),
		AdvisorMode:    vo.AdvisorMode(r.AdvisorMode),
		AdvisorOutcome: vo.AdvisorOutcome(r.AdvisorOutcome),
		ReviewDecision: vo.ReviewDecision(r.ReviewDecision),
		Source:         vo.Source(r.Source),
		CaseID:         r.CaseID,
		BenchmarkID:    benchmarkID,
		InputText:      r.InputText,
		Image:          r.Image,
		ImageSize:      r.ImageSize,
		ImageMIME:      r.ImageMime,
		Canaries:       canaries,
		Anonymized:     r.AnonymizedText,
		Vault:          vault,
		Findings:       findings,
		Card:           card,
		Draft:          r.Draft,
		Advice:         advice,
		FinalAnonymous: r.FinalAnonymized,
		FinalText:      r.FinalText,
		JudgeScore:     score,
		JudgeReason:    r.JudgeReason,
		Gate2Passes:    int(r.Gate2Passes),
		Escalated:      r.Escalated != 0,
		CanariesLeaked: int(r.CanariesLeaked),
		Error:          r.Error,
	})
}

func parseTime(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse time %q: %w", s, err)
	}
	return t, nil
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func encode(v any) string {
	raw, _ := json.Marshal(v)
	return string(raw)
}

func benchmarkIDString(id uuid.UUID) string {
	if id == uuid.Nil() {
		return ""
	}
	return id.String()
}

func boolInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

func scorePointer(score int) *int64 {
	if score == 0 {
		return nil
	}
	v := int64(score)
	return &v
}
