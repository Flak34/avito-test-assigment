CREATE TABLE IF NOT EXISTS pickup_points (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    name VARCHAR(50) NOT NULL,
    address VARCHAR(50) NOt NULL,
    phone_number VARCHAR(50) NOT NULL
);