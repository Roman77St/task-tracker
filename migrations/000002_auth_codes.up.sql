CREATE TABLE auth_codes (
    user_id BIGINT PRIMARY KEY,
    code TEXT NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);