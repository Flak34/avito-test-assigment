CREATE TABLE banner (
   id BIGSERIAL PRIMARY KEY,
   content JSONB NOT NULL,
   is_active BOOLEAN NOT NULL,
   created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
   updated_at TIMESTAMP WITH TIME ZONE
);


CREATE TABLE banner_feature_tag (
    banner_id BIGINT REFERENCES banner(id) ON DELETE CASCADE NOT NULL,
    tag_id BIGINT,
    feature_id BIGINT,
    PRIMARY KEY (tag_id, feature_id)
);
CREATE INDEX banner_feature_tag_banner_id_idx on banner_feature_tag(banner_id);
CREATE INDEX banner_feature_tag_feature_id_idx on banner_feature_tag(feature_id);


CREATE TABLE banner_version (
    id BIGSERIAL PRIMARY KEY,
    banner_id BIGINT REFERENCES banner(id) ON DELETE CASCADE NOT NULL,
    banner JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX banner_version_banner_id_idx on banner_version(banner_id);