package repository

import (
	"avito-test-assigment/banner_app/internal/repository/shema"
	"avito-test-assigment/banner_app/internal/repository/transaction_manager"
	"context"
	"errors"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	"time"
)

type BannerRepository struct {
	queryEngineProvider transaction_manager.QueryEngineProvider
}

func NewBannerRepository(tm transaction_manager.QueryEngineProvider) *BannerRepository {
	return &BannerRepository{queryEngineProvider: tm}
}

func (r *BannerRepository) GetByTagAndFeature(ctx context.Context, tagID int64, featureID int64) (*shema.Banner, error) {
	querier := r.queryEngineProvider.GetQueryEngine(ctx)

	var banner shema.Banner
	err := pgxscan.Get(ctx, querier, &banner, `SELECT id, content, is_active, created_at FROM banner WHERE id = 
		            (SELECT banner_id FROM banner_feature_tag WHERE tag_id = $1 and feature_id = $2)`, tagID, featureID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrObjectNotFound
		}
		return nil, err
	}

	return &banner, nil
}

func (r *BannerRepository) GetById(ctx context.Context, bannerID int64) (*shema.Banner, error) {
	querier := r.queryEngineProvider.GetQueryEngine(ctx)

	var banner shema.Banner
	err := pgxscan.Get(ctx, querier, &banner,
		`SELECT id, content, is_active, created_at FROM banner WHERE id = $1`, bannerID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrObjectNotFound
		}
		return nil, err
	}

	return &banner, nil
}

func (r *BannerRepository) DeleteByID(ctx context.Context, bannerID int64) error {
	querier := r.queryEngineProvider.GetQueryEngine(ctx)
	comTag, err := querier.Exec(ctx, "DELETE FROM banner WHERE id = $1", bannerID)
	if err != nil {
		return err
	}

	if comTag.RowsAffected() == 0 {
		return ErrObjectNotFound
	}

	return nil
}

func (r *BannerRepository) Insert(ctx context.Context, banner shema.Banner) (int64, error) {
	querier := r.queryEngineProvider.GetQueryEngine(ctx)

	var id int64

	err := querier.QueryRow(ctx, "INSERT INTO banner (content, is_active) VALUES ($1, $2) RETURNING id",
		banner.Content, banner.IsActive).Scan(&id)
	if err != nil {
		return id, err
	}

	return id, nil
}

func (r *BannerRepository) Update(ctx context.Context, banner shema.Banner) error {
	querier := r.queryEngineProvider.GetQueryEngine(ctx)

	comTag, err := querier.Exec(ctx,
		"UPDATE banner SET content = $1, is_active = $2, updated_at = $3 WHERE id = $4",
		banner.Content, banner.IsActive, time.Now(), banner.ID)

	if comTag.RowsAffected() == 0 {
		return ErrObjectNotFound
	}
	if err != nil {
		return err
	}

	return nil
}

func (r *BannerRepository) AddTag(ctx context.Context, bannerID int64, tagID int64, featureID int64) error {
	querier := r.queryEngineProvider.GetQueryEngine(ctx)

	_, err := querier.Exec(ctx,
		"INSERT INTO banner_feature_tag(banner_id, tag_id, feature_id) VALUES ($1, $2, $3)",
		bannerID, tagID, featureID)

	if err != nil {
		return err
	}

	return nil
}

func (r *BannerRepository) ListTags(ctx context.Context, bannerID int64) ([]*int64, error) {
	querier := r.queryEngineProvider.GetQueryEngine(ctx)

	var tags []*int64
	err := pgxscan.Select(ctx,
		querier, &tags, "SELECT tag_id FROM banner_feature_tag WHERE banner_id = $1", bannerID)
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func (r *BannerRepository) DeleteTags(ctx context.Context, bannerID int64) error {
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

func (r *BannerRepository) GetFeature(ctx context.Context, bannerID int64) (int64, error) {
	querier := r.queryEngineProvider.GetQueryEngine(ctx)

	var feature int64
	err := pgxscan.Get(ctx,
		querier, &feature, "SELECT feature_id FROM banner_feature_tag WHERE banner_id = $1 LIMIT 1", bannerID)
	if err != nil {
		return feature, err
	}
	return feature, nil
}

func (r *BannerRepository) ListByTag(ctx context.Context, tagID int64, offset int, limit int) ([]*shema.Banner, error) {
	querier := r.queryEngineProvider.GetQueryEngine(ctx)

	var banners []*shema.Banner
	err := pgxscan.Select(ctx,
		querier, &banners, `SELECT id, is_active, content, created_at, updated_at FROM banner WHERE id IN 
                           (SELECT banner_id FROM banner_feature_tag WHERE tag_id = $1 LIMIT $2 OFFSET $3)`,
		tagID, limit, offset)
	if err != nil {
		return nil, err
	}
	return banners, nil
}

func (r *BannerRepository) ListByFeature(ctx context.Context, featureID int64, offset int, limit int) ([]*shema.Banner, error) {
	querier := r.queryEngineProvider.GetQueryEngine(ctx)

	var banners []*shema.Banner
	err := pgxscan.Select(ctx, querier, &banners,
		`SELECT id, is_active, content, created_at, updated_at FROM banner WHERE id IN 
                           (SELECT DISTINCT banner_id FROM banner_feature_tag WHERE feature_id = $1 LIMIT $2 OFFSET $3)`,
		featureID, limit, offset)
	if err != nil {
		return nil, err
	}
	return banners, nil
}
