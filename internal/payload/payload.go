package payload

import (
	"avito-test-assigment/internal/model"
	"encoding/json"
)

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type CreateBannerResponse struct {
	BannerID int64 `json:"banner_id"`
}

type UpdateBannerRequest struct {
	ID        int64
	TagIDs    []*int64               `json:"tag_ids"`
	FeatureID int64                  `json:"feature_id"`
	Content   map[string]interface{} `json:"content"`
	IsActive  bool                   `json:"is_active"`
}

func (r *UpdateBannerRequest) BuildBanner() *model.Banner {
	content, _ := json.Marshal(&r.Content)
	c := json.RawMessage(content)
	return &model.Banner{
		ID:        r.ID,
		FeatureID: r.FeatureID,
		TagIDs:    r.TagIDs,
		Content:   &c,
		IsActive:  r.IsActive,
	}
}

type CreateBannerRequest struct {
	TagIDs    []*int64               `json:"tag_ids"`
	FeatureID int64                  `json:"feature_id"`
	Content   map[string]interface{} `json:"content"`
	IsActive  bool                   `json:"is_active"`
}

func (r *CreateBannerRequest) BuildBanner() *model.Banner {
	content, _ := json.Marshal(&r.Content)
	c := json.RawMessage(content)
	return &model.Banner{
		FeatureID: r.FeatureID,
		TagIDs:    r.TagIDs,
		Content:   &c,
		IsActive:  r.IsActive,
	}
}
