package shema

import "time"

type Banner struct {
	ID        int64      `db:"id"`
	Content   string     `db:"content"`
	IsActive  bool       `db:"is_active"`
	CreatedAt *time.Time `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
}

type BannerVersion struct {
	ID        int64     `db:"id"`
	Banner    string    `db:"banner"`
	BannerID  int64     `db:"banner_id"`
	CreatedAt time.Time `db:"created_at"`
}
