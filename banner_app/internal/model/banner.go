package model

import (
	"encoding/json"
	"time"
)

type Banner struct {
	ID        int64            `json:"id"`
	Content   *json.RawMessage `json:"content"`
	IsActive  bool             `json:"is_active"`
	TagIDs    []*int64         `json:"tag_ids"`
	FeatureID int64            `json:"feature_id"`
	CreatedAt *time.Time       `json:"created_at"`
	UpdatedAt *time.Time       `json:"updated_at"`
}

type UserBanner struct {
	Content  string
	IsActive bool
}
