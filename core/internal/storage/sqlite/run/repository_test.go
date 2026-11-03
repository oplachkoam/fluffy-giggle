package run

import (
	"context"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"fluffy-giggle/core/internal/domain/entity"
	"fluffy-giggle/core/internal/domain/vo"
	sqlitedb "fluffy-giggle/core/internal/platform/database/sqlite"
	"fluffy-giggle/core/migrations"
)

func TestRepository(t *testing.T) {
	t.Parallel()

	conn, err := sqlitedb.Open(sqlitedb.Config{Path: t.TempDir() + "/test.db"})
	require.NoError(t, err)
	require.NoError(t, sqlitedb.Migrate(context.Background(), conn, migrations.Files))
	repo := New(conn)
	now := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)

	var ids []uuid.UUID
	for i := range 3 {
		input := entity.NewRunInput{
			Source:      vo.SourceText,
			AdvisorMode: vo.AdvisorOn,
			Text:        "письмо",
			CaseID:      "case-01",
		}
		if i == 1 {
			input = entity.NewRunInput{
				Source:      vo.SourceImage,
				AdvisorMode: vo.AdvisorOff,
				Image:       []byte{1, 2, 3},
				ImageMIME:   "image/png",
			}
		}
		run, err := entity.NewRun(input, now.Add(time.Duration(i)*time.Second))
		require.NoError(t, err)
		require.NoError(t, repo.Create(t.Context(), run))
		ids = append(ids, run.ID())
	}

	t.Run("get loads the image and round-trips state", func(t *testing.T) {
		run, err := repo.Get(t.Context(), ids[1])
		require.NoError(t, err)
		assert.Equal(t, []byte{1, 2, 3}, run.Image())
		assert.True(t, run.HasImage())

		vault := vo.NewVault()
		vault.Assign(vo.PIIPerson, "Иванов", "Иванова")
		require.NoError(t, run.Start(now))
		run.SetAnonymized("[PERSON_1]", vault, 2, now)
		run.SetDraft(vo.Card{Category: "выплата"}, "черновик", now)
		require.NoError(t, run.SetJudgement(vo.Judgement{Score: 4, Reason: "ок"}, now))
		require.NoError(t, repo.Update(t.Context(), run))

		again, err := repo.Get(t.Context(), ids[1])
		require.NoError(t, err)
		assert.Equal(t, vo.PhaseCompose, again.Phase())
		assert.Equal(t, []string{"Иванова", "Иванов"}, again.Vault().Entries()[0].Forms)
		assert.Equal(t, "выплата", again.Card().Category)
		assert.Equal(t, 4, again.JudgeScore())
	})

	t.Run("list pages newest first without image bytes", func(t *testing.T) {
		page, err := repo.List(t.Context(), 2, uuid.Nil())
		require.NoError(t, err)
		require.Len(t, page, 2)
		assert.Equal(t, ids[2], page[0].ID())
		assert.Equal(t, ids[1], page[1].ID())
		assert.True(t, page[1].HasImage())
		assert.Empty(t, page[1].Image())

		rest, err := repo.List(t.Context(), 2, page[1].ID())
		require.NoError(t, err)
		require.Len(t, rest, 1)
		assert.Equal(t, ids[0], rest[0].ID())
	})

	t.Run("missing run is nil", func(t *testing.T) {
		run, err := repo.Get(t.Context(), uuid.NewV7())
		require.NoError(t, err)
		assert.Nil(t, run)
	})
}
