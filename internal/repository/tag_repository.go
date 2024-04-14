package repository

import (
	"avito-test-assigment/internal/repository/transaction_manager"
	"context"
	"github.com/georgysavva/scany/pgxscan"
)

type TagRepository struct {
	queryEngineProvider transaction_manager.QueryEngineProvider
}

func NewTagRepository(tm transaction_manager.QueryEngineProvider) *TagRepository {
	return &TagRepository{queryEngineProvider: tm}
}

func (r *TagRepository) AddTag(ctx context.Context, bannerID int64, tagID int64, featureID int64) error {
	querier := r.queryEngineProvider.GetQueryEngine(ctx)

	_, err := querier.Exec(ctx,
		"INSERT INTO banner_feature_tag(banner_id, tag_id, feature_id) VALUES ($1, $2, $3)",
		bannerID, tagID, featureID)

	if err != nil {
		return err
	}

	return nil
}

func (r *TagRepository) ListTags(ctx context.Context, bannerID int64) ([]*int64, error) {
	querier := r.queryEngineProvider.GetQueryEngine(ctx)

	var tags []*int64
	err := pgxscan.Select(ctx,
		querier, &tags, "SELECT tag_id FROM banner_feature_tag WHERE banner_id = $1", bannerID)
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func (r *TagRepository) DeleteTags(ctx context.Context, bannerID int64) error {
	querier := r.queryEngineProvider.GetQueryEngine(ctx)

	_, err := querier.Exec(ctx, "DELETE FROM banner_feature_tag WHERE banner_id = $1", bannerID)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	return nil
}
