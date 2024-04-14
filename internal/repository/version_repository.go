package repository

import (
	"avito-test-assigment/internal/repository/shema"
	"avito-test-assigment/internal/repository/transaction_manager"
	"context"
	"errors"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	"time"
)

type VersionRepository struct {
	queryEngineProvider transaction_manager.QueryEngineProvider
}

func NewVersionRepository(tm transaction_manager.QueryEngineProvider) *TagRepository {
	return &TagRepository{queryEngineProvider: tm}
}

func (r *TagRepository) AddVersion(ctx context.Context, bannerID int64, banner string) error {
	querier := r.queryEngineProvider.GetQueryEngine(ctx)

	_, err := querier.Exec(ctx,
		"INSERT INTO banner_version(banner_id, banner, created_at) VALUES ($1, $2, $3)",
		bannerID, banner, time.Now())

	if err != nil {
		return err
	}

	return nil
}

func (r *TagRepository) RemoveOldVersions(ctx context.Context, bannerID int64, countOfLeftVersions int) error {
	querier := r.queryEngineProvider.GetQueryEngine(ctx)

	_, err := querier.Exec(ctx,
		`DELETE FROM banner_version WHERE id IN 
                                 (SELECT id FROM banner_version WHERE banner_id = $1 ORDER BY created_at DESC OFFSET $2)`,
		bannerID, countOfLeftVersions)

	if err != nil {
		return err
	}

	return nil
}

func (r *TagRepository) GetBannerVersions(ctx context.Context, bannerID int64) ([]*shema.BannerVersion, error) {
	querier := r.queryEngineProvider.GetQueryEngine(ctx)

	var versions []*shema.BannerVersion

	err := pgxscan.Select(ctx, querier, &versions, `SELECT * FROM banner_version WHERE banner_id = $1`, bannerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrObjectNotFound
		}
		return nil, err
	}

	return versions, nil
}
