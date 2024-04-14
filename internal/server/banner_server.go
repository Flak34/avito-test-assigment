package server

import (
	"avito-test-assigment/internal/model"
	"avito-test-assigment/internal/payload"
	"avito-test-assigment/internal/service"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"strconv"
)

type BannerService interface {
	GetByTagAndFeature(ctx context.Context, tagID int64, featureID int64, useLastRevision bool) (string, error)
	DeleteByID(ctx context.Context, bannerID int64) error
	Update(ctx context.Context, banner model.Banner) error
	Create(ctx context.Context, banner model.Banner) (int64, error)
	ListBannersByTag(ctx context.Context, tagID int64, offset int, limit int) ([]model.Banner, error)
	ListBannersByFeature(ctx context.Context, featureID int64, offset int, limit int) ([]model.Banner, error)
	ListBanners(ctx context.Context, offset int, limit int) ([]model.Banner, error)
	GetByTagAndFeatureAdmin(ctx context.Context, tagID int64, featureID int64) (model.Banner, error)
	DeleteByTag(ctx context.Context, tagID int64) error
	DeleteByFeature(ctx context.Context, featureID int64) error
	GetBannerVersions(ctx context.Context, bannerID int64) ([]model.BannerVersion, error)
}

func NewBannerServer(service BannerService) *BannerServer {
	return &BannerServer{service: service}
}

type BannerServer struct {
	service BannerService
}

func (server *BannerServer) GetByTagAndFeature(w http.ResponseWriter, req *http.Request) error {
	tagIdStr := req.URL.Query().Get("tag_id")
	featureIdStr := req.URL.Query().Get("feature_id")
	useLastRevisionStr := req.URL.Query().Get("use_last_revision")

	if tagIdStr == "" || featureIdStr == "" {
		return fmt.Errorf("%w: необходимо задать id тега и фичи", service.ErrIncorrectData)
	}

	tagID, err := strconv.ParseInt(tagIdStr, 10, 64)
	featureID, err2 := strconv.ParseInt(featureIdStr, 10, 64)
	if err != nil || err2 != nil {
		return fmt.Errorf("%w: id тега и фичи должны быть целыми числами", service.ErrIncorrectData)
	}

	useLastRevision := useLastRevisionStr == "true"

	//TODO проверка на админа
	//fromAdmin := false

	bannerContent, err := server.service.GetByTagAndFeature(req.Context(), tagID, featureID, useLastRevision)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(bannerContent))
	if err != nil {
		return err
	}
	return nil
}

func (server *BannerServer) DeleteByID(w http.ResponseWriter, req *http.Request) error {
	bannerID, err := strconv.ParseInt(chi.URLParam(req, "id"), 10, 64)
	if err != nil {
		return fmt.Errorf("%w: некорректный id баннера", service.ErrIncorrectData)
	}

	err = server.service.DeleteByID(req.Context(), bannerID)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (server *BannerServer) Update(w http.ResponseWriter, req *http.Request) error {
	bannerID, err := strconv.ParseInt(chi.URLParam(req, "id"), 10, 64)
	if err != nil {
		return fmt.Errorf("%w: некорректный id баннера", service.ErrIncorrectData)
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}

	var unm payload.UpdateBannerRequest
	err = json.Unmarshal(body, &unm)
	if err != nil {
		return fmt.Errorf("%w: %s", service.ErrIncorrectData, err.Error())
	}
	unm.ID = bannerID

	banner := unm.BuildBanner()
	err = server.service.Update(req.Context(), *banner)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func (server *BannerServer) Create(w http.ResponseWriter, req *http.Request) error {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}

	var unm payload.CreateBannerRequest
	err = json.Unmarshal(body, &unm)
	if err != nil {
		return fmt.Errorf("%w: %s", service.ErrIncorrectData, err.Error())
	}

	banner := unm.BuildBanner()
	id, err := server.service.Create(req.Context(), *banner)
	if err != nil {
		return err
	}

	response := payload.CreateBannerResponse{BannerID: id}
	responseBytes, err := json.Marshal(&response)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(responseBytes)
	return nil
}

func (server *BannerServer) ListBannersByTagOrFeature(w http.ResponseWriter, req *http.Request) error {
	tagIdStr := req.URL.Query().Get("tag_id")
	featureIdStr := req.URL.Query().Get("feature_id")
	offsetStr := req.URL.Query().Get("offset")
	limitStr := req.URL.Query().Get("limit")

	var offset, limit = 0, 10
	if offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err != nil || o < 0 {
			return fmt.Errorf("%w: offset должен быть целым неотрицательным числом", service.ErrIncorrectData)
		}
		offset = o
	}

	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil || l < 0 {
			return fmt.Errorf("%w: limit должен быть целым неотрицательным числом", service.ErrIncorrectData)
		}
		limit = l
	}

	tagID, err1 := strconv.ParseInt(tagIdStr, 10, 64)
	featureID, err2 := strconv.ParseInt(featureIdStr, 10, 64)

	var result []model.Banner
	var err error

	if tagIdStr != "" && featureIdStr != "" {
		if !(err1 == nil && err2 == nil) {
			return fmt.Errorf("%w: tag_id и feature_id должны быть целыми числами", service.ErrIncorrectData)
		}
		b, err := server.service.GetByTagAndFeatureAdmin(req.Context(), tagID, featureID)
		if err != nil {
			return err
		}
		result = append(result, b)
	} else if tagIdStr != "" {
		if err1 != nil {
			return fmt.Errorf("%w: tag_id должен быть целым числом", service.ErrIncorrectData)
		}
		result, err = server.service.ListBannersByTag(req.Context(), tagID, offset, limit)
		if err != nil {
			return err
		}
	} else if featureIdStr != "" {
		if err2 != nil {
			return fmt.Errorf("%w: feature_id должен быть целым числом", service.ErrIncorrectData)
		}
		result, err = server.service.ListBannersByFeature(req.Context(), featureID, offset, limit)
		if err != nil {
			return err
		}
	} else {
		result, err = server.service.ListBanners(req.Context(), offset, limit)
		if err != nil {
			return err
		}
	}

	if result == nil {
		result = make([]model.Banner, 0)
	}

	bytes, err := json.Marshal(&result)
	if err != nil {
		return err
	}

	_, err = w.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}

func (server *BannerServer) DeleteByTagOrFeature(w http.ResponseWriter, req *http.Request) error {
	tagIdStr := req.URL.Query().Get("tag_id")
	featureIdStr := req.URL.Query().Get("feature_id")

	tagID, err1 := strconv.ParseInt(tagIdStr, 10, 64)
	featureID, err2 := strconv.ParseInt(featureIdStr, 10, 64)

	if tagIdStr != "" && featureIdStr != "" {
		return fmt.Errorf("%w: в данном запроссе можно задать только один из фильтров: tagID или featureID",
			service.ErrIncorrectData)
	} else if tagIdStr != "" {
		if err1 != nil {
			return fmt.Errorf("%w: некорректный tagID",
				service.ErrIncorrectData)
		}
		err := server.service.DeleteByTag(context.TODO(), tagID)
		if err != nil {
			return err
		}
	} else if featureIdStr != "" {
		if err2 != nil {
			return fmt.Errorf("%w: некорректный featureID",
				service.ErrIncorrectData)
		}
		err := server.service.DeleteByFeature(context.TODO(), featureID)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("%w: в данном запроссе необходимо задать один из фильтров: tagID или featureID",
			service.ErrIncorrectData)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Запрос поступил в работу"))
	return nil
}

func (server *BannerServer) GetBannerVersions(w http.ResponseWriter, req *http.Request) error {
	bannerID, err := strconv.ParseInt(chi.URLParam(req, "id"), 10, 64)
	if err != nil {
		return fmt.Errorf("%w: некорректный id баннера", service.ErrIncorrectData)
	}

	versions, err := server.service.GetBannerVersions(req.Context(), bannerID)
	if err != nil {
		return err
	}

	bytes, err := json.Marshal(&versions)
	if err != nil {
		return err
	}

	_, err = w.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}
