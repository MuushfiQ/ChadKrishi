USE defaultdb;

ALTER TABLE users
  MODIFY gardening_space ENUM('rooftop', 'balcony') NULL,
  MODIFY sunlight_hours DECIMAL(4, 1) NULL,
  MODIFY rain_exposure BOOLEAN NULL;

DELETE FROM sessions WHERE expires_at <= UTC_TIMESTAMP() OR revoked_at IS NOT NULL;
