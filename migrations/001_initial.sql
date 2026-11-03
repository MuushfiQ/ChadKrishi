CREATE DATABASE IF NOT EXISTS chadkrishi
  CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE defaultdb;

CREATE TABLE IF NOT EXISTS users (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(100) NOT NULL,
  email VARCHAR(254) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  profile_photo VARCHAR(255),
  district VARCHAR(80),
  latitude DECIMAL(9, 6),
  longitude DECIMAL(9, 6),
  gardening_space ENUM('rooftop', 'balcony'),
  sunlight_hours DECIMAL(4, 1),
  rain_exposure BOOLEAN,
  bio VARCHAR(500),
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS sessions (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  token_hash CHAR(64) NOT NULL UNIQUE,
  expires_at DATETIME NOT NULL,
  revoked_at DATETIME,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_sessions_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  INDEX idx_sessions_active (token_hash, expires_at, revoked_at)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS crops (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(100) NOT NULL UNIQUE,
  description TEXT NOT NULL,
  image_path VARCHAR(255),
  planting_months JSON NOT NULL,
  min_sunlight_hours DECIMAL(4, 1) NOT NULL,
  max_sunlight_hours DECIMAL(4, 1) NOT NULL,
  suitable_spaces JSON NOT NULL,
  container_guidance TEXT NOT NULL,
  care_guidance TEXT NOT NULL,
  source_reference VARCHAR(500) NOT NULL,
  verification_status ENUM('demo', 'reviewed') NOT NULL DEFAULT 'demo'
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS user_plants (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  crop_id BIGINT UNSIGNED NOT NULL,
  nickname VARCHAR(100) NOT NULL,
  planting_date DATE NOT NULL,
  propagation_method ENUM('seed', 'seedling') NOT NULL,
  growth_stage ENUM('seedling', 'growing', 'flowering', 'fruiting_harvest') NOT NULL,
  photo_path VARCHAR(255),
  note VARCHAR(500),
  archived_at DATETIME,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  CONSTRAINT fk_plants_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_plants_crop FOREIGN KEY (crop_id) REFERENCES crops(id),
  INDEX idx_plants_user_active (user_id, archived_at)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS care_tasks (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  plant_id BIGINT UNSIGNED NOT NULL,
  rule_key VARCHAR(80) NOT NULL,
  title VARCHAR(160) NOT NULL,
  reason VARCHAR(300) NOT NULL,
  due_date DATE NOT NULL,
  completed_at DATETIME,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_tasks_plant FOREIGN KEY (plant_id) REFERENCES user_plants(id) ON DELETE CASCADE,
  UNIQUE KEY uq_task_identity (plant_id, rule_key, due_date)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS posts (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  plant_id BIGINT UNSIGNED,
  type ENUM('update', 'question', 'blog', 'sale') NOT NULL,
  title VARCHAR(160),
  content TEXT NOT NULL,
  image_path VARCHAR(255),
  price_bdt DECIMAL(10, 2),
  quantity INT UNSIGNED,
  contact_method VARCHAR(160),
  sale_status ENUM('available', 'sold') DEFAULT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  CONSTRAINT fk_posts_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_posts_plant FOREIGN KEY (plant_id) REFERENCES user_plants(id) ON DELETE SET NULL,
  INDEX idx_posts_feed (created_at, id),
  INDEX idx_posts_sale (type, sale_status)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS comments (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  post_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  content VARCHAR(1000) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_comments_post FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
  CONSTRAINT fk_comments_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS post_likes (
  post_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (post_id, user_id),
  CONSTRAINT fk_likes_post FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
  CONSTRAINT fk_likes_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB;

