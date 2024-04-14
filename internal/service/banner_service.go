package service

import (
	"avito-test-assigment/internal/model"
	"avito-test-assigment/internal/repository"
	"avito-test-assigment/internal/repository/shema"
	"context"
	"encoding/json"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"strconv"
	"time"
)

type VersionRepository interface {
	AddVersion(ctx context.Context, bannerID int64, banner string) error
	RemoveOldVersions(ctx context.Context, bannerID int64, countOfLeftVersions int) error
	GetBannerVersions(ctx context.Context, bannerID int64) ([]*shema.BannerVersion, error)
}

type TagRepository interface {
	AddTag(ctx context.Context, bannerID int64, tagID int64, featureID int64) error
	ListTags(ctx context.Context, bannerID int64) ([]*int64, error)
	DeleteTags(ctx context.Context, bannerID int64) error
}

type BannerRepository interface {
	GetByTagAndFeature(ctx context.Context, tagID int64, featureID int64) (*shema.Banner, error)
	GetById(ctx context.Context, bannerID int64) (*shema.Banner, error)
	DeleteByID(ctx context.Context, bannerID int64) error
	Insert(ctx context.Context, banner shema.Banner) (int64, error)
	Update(ctx context.Context, banner shema.Banner) error
	GetFeature(ctx context.Context, bannerID int64) (int64, error)
	ListByTag(ctx context.Context, tagID int64, offset int, limit int) ([]*shema.Banner, error)
	ListByFeature(ctx context.Context, featureID int64, offset int, limit int) ([]*shema.Banner, error)
	List(ctx context.Context, offset int, limit int) ([]*shema.Banner, error)
	DeleteByTag(ctx context.Context, tagID int64, limit int) error
	DeleteByFeature(ctx context.Context, featureID int64, limit int) error
}

type TransactionManager interface {
	RunSerializable(ctx context.Context, f func(ctxTX context.Context) error) error
	RunReadCommitted(ctx context.Context, f func(ctxTX context.Context) error) error
}

type BannerService struct {
	bannerRepository   BannerRepository
	tagRepository      TagRepository
	versionRepository  VersionRepository
	transactionManager TransactionManager

	bannerCache *expirable.LRU[string, string]
}

func NewBannerService(bannerRepository BannerRepository, manager TransactionManager,
	tagRepository TagRepository, versionRepository VersionRepository) *BannerService {
	return &BannerService{
		transactionManager: manager,
		bannerRepository:   bannerRepository,
		tagRepository:      tagRepository,
		versionRepository:  versionRepository,
		bannerCache:        expirable.NewLRU[string, string](30000, nil, time.Minute*5),
	}
}

func (s *BannerService) Create(ctx context.Context, banner model.Banner) (int64, error) {
	var id int64
	err := s.transactionManager.RunReadCommitted(ctx, func(ctxTX context.Context) error {

		var err error
		id, err = s.bannerRepository.Insert(ctxTX, *buildShemaBanner(&banner))
		if err != nil {
			return err
		}

		for _, tagID := range banner.TagIDs {
			err = s.tagRepository.AddTag(ctxTX, id, *tagID, banner.FeatureID)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return id, err
	}

	return id, nil
}

func (s *BannerService) DeleteByID(ctx context.Context, bannerID int64) error {
	return s.bannerRepository.DeleteByID(ctx, bannerID)
}

func (s *BannerService) GetByTagAndFeature(ctx context.Context, tagID int64, featureID int64, useLastRevision bool) (string, error) {

	bannerFromCache, ok := s.bannerCache.Get(strconv.FormatInt(tagID, 10) + ":" +
		strconv.FormatInt(featureID, 10))

	if !useLastRevision {
		if ok {
			return bannerFromCache, nil
		}
	}

	banner, err := s.bannerRepository.GetByTagAndFeature(ctx, tagID, featureID)
	if err != nil {
		return "", err
	}

	if !banner.IsActive {
		return "", repository.ErrObjectNotFound
	}

	s.bannerCache.Add(strconv.FormatInt(tagID, 10)+":"+
		strconv.FormatInt(featureID, 10), banner.Content)

	return banner.Content, err
}

func (s *BannerService) GetByTagAndFeatureAdmin(ctx context.Context, tagID int64, featureID int64) (model.Banner, error) {
	var result model.Banner
	err := s.transactionManager.RunReadCommitted(ctx, func(ctxTX context.Context) error {
		banner, err := s.bannerRepository.GetByTagAndFeature(ctxTX, tagID, featureID)
		if err != nil {
			return err
		}

		tags, err := s.tagRepository.ListTags(ctxTX, banner.ID)
		if err != nil {
			return err
		}

		feature, err := s.bannerRepository.GetFeature(ctxTX, banner.ID)
		if err != nil {
			return err
		}

		result = *buildModelBanner(banner)
		result.TagIDs = tags
		result.FeatureID = feature

		return nil
	})

	if err != nil {
		return model.Banner{}, err
	}

	return result, nil
}

func (s *BannerService) ListBannersByTag(ctx context.Context, tagID int64, offset int, limit int) ([]model.Banner, error) {

	var result []model.Banner
	err := s.transactionManager.RunReadCommitted(ctx, func(ctxTX context.Context) error {
		banners, err := s.bannerRepository.ListByTag(ctxTX, tagID, offset, limit)
		if err != nil {
			return err
		}

		for _, b := range banners {
			modelBanner := buildModelBanner(b)
			tags, err := s.tagRepository.ListTags(ctxTX, b.ID)
			if err != nil {
				return err
			}
			feature, err := s.bannerRepository.GetFeature(ctxTX, b.ID)
			if err != nil {
				return err
			}
			modelBanner.TagIDs = tags
			modelBanner.FeatureID = feature
			result = append(result, *modelBanner)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *BannerService) ListBanners(ctx context.Context, offset int, limit int) ([]model.Banner, error) {

	var result []model.Banner
	err := s.transactionManager.RunReadCommitted(ctx, func(ctxTX context.Context) error {
		banners, err := s.bannerRepository.List(ctxTX, offset, limit)
		if err != nil {
			return err
		}

		for _, b := range banners {
			modelBanner := buildModelBanner(b)
			tags, err := s.tagRepository.ListTags(ctxTX, b.ID)
			if err != nil {
				return err
			}
			feature, err := s.bannerRepository.GetFeature(ctxTX, b.ID)
			if err != nil {
				return err
			}
			modelBanner.TagIDs = tags
			modelBanner.FeatureID = feature
			result = append(result, *modelBanner)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *BannerService) ListBannersByFeature(ctx context.Context, featureID int64, offset int, limit int) ([]model.Banner, error) {
	var result []model.Banner
	err := s.transactionManager.RunReadCommitted(ctx, func(ctxTX context.Context) error {
		banners, err := s.bannerRepository.ListByFeature(ctxTX, featureID, offset, limit)
		if err != nil {
			return err
		}

		for _, b := range banners {
			modelBanner := buildModelBanner(b)
			tags, err := s.tagRepository.ListTags(ctxTX, b.ID)
			if err != nil {
				return err
			}
			feature, err := s.bannerRepository.GetFeature(ctxTX, b.ID)
			if err != nil {
				return err
			}
			modelBanner.TagIDs = tags
			modelBanner.FeatureID = feature
			result = append(result, *modelBanner)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *BannerService) Update(ctx context.Context, banner model.Banner) error {
	err := s.transactionManager.RunReadCommitted(ctx, func(ctxTX context.Context) error {
		oldBanner, err := s.bannerRepository.GetById(ctxTX, banner.ID)
		if err != nil {
			return err
		}

		err = s.versionRepository.AddVersion(ctxTX, banner.ID, oldBanner.Content)
		if err != nil {
			return err
		}
		err = s.versionRepository.RemoveOldVersions(ctxTX, banner.ID, 3)
		if err != nil {
			return err
		}

		err = s.bannerRepository.Update(ctxTX, *buildShemaBanner(&banner))
		if err != nil {
			return err
		}

		err = s.tagRepository.DeleteTags(ctxTX, banner.ID)
		if err != nil {
			return err
		}

		for _, tag := range banner.TagIDs {
			err = s.tagRepository.AddTag(ctxTX, banner.ID, *tag, banner.FeatureID)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *BannerService) DeleteByTag(ctx context.Context, tagID int64) error {

	go func() {
		for {
			err := s.bannerRepository.DeleteByTag(ctx, tagID, 10000)
			if err != nil {
				return
			}
		}
	}()

	return nil
}

func (s *BannerService) DeleteByFeature(ctx context.Context, featureID int64) error {

	go func() {
		for {
			err := s.bannerRepository.DeleteByFeature(ctx, featureID, 10000)
			if err != nil {
				return
			}
		}
	}()

	return nil
}

func (s *BannerService) GetBannerVersions(ctx context.Context, bannerID int64) ([]model.BannerVersion, error) {

	versions, err := s.versionRepository.GetBannerVersions(ctx, bannerID)
	if err != nil {
		return nil, err
	}

	result := make([]model.BannerVersion, 0)
	for _, v := range versions {
		result = append(result, *buildModelBannerVersion(v))
	}

	return result, nil
}

func buildShemaBanner(banner *model.Banner) *shema.Banner {
	return &shema.Banner{
		ID:        banner.ID,
		Content:   string(*banner.Content),
		IsActive:  banner.IsActive,
		CreatedAt: banner.CreatedAt,
		UpdatedAt: banner.UpdatedAt,
	}
}

func buildModelBanner(banner *shema.Banner) *model.Banner {
	c := json.RawMessage(banner.Content)
	return &model.Banner{
		ID:        banner.ID,
		Content:   &c,
		IsActive:  banner.IsActive,
		CreatedAt: banner.CreatedAt,
		UpdatedAt: banner.UpdatedAt,
	}
}

func buildModelBannerVersion(bannerVersion *shema.BannerVersion) *model.BannerVersion {
	c := json.RawMessage(bannerVersion.Banner)
	return &model.BannerVersion{
		ID:        bannerVersion.ID,
		Banner:    &c,
		BannerID:  bannerVersion.BannerID,
		CreatedAt: bannerVersion.CreatedAt,
	}
}
