-- Apply ONCE after 001–006. Back up the database first.
USE defaultdb;
ALTER TABLE users ADD COLUMN is_admin BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE posts MODIFY sale_status ENUM('available','reserved','sold') DEFAULT NULL, ADD COLUMN archived_at DATETIME NULL;
ALTER TABLE post_likes ADD COLUMN reaction ENUM('like','love','angry','sad') NOT NULL DEFAULT 'like';
CREATE TABLE orders (
 id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT, post_id BIGINT UNSIGNED NOT NULL,
 buyer_id BIGINT UNSIGNED NOT NULL, seller_id BIGINT UNSIGNED NOT NULL,
 title VARCHAR(160) NOT NULL, amount_bdt DECIMAL(10,2) NOT NULL, quantity INT UNSIGNED NOT NULL,
 status ENUM('waiting_payment','buyer_paid','completed','cancelled') NOT NULL DEFAULT 'waiting_payment',
 payment_method ENUM('cash','bkash','other') NULL, payment_reference VARCHAR(160) NOT NULL DEFAULT '',
 buyer_paid_at DATETIME NULL, received_at DATETIME NULL,
 created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
 updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
 FOREIGN KEY(post_id) REFERENCES posts(id), FOREIGN KEY(buyer_id) REFERENCES users(id), FOREIGN KEY(seller_id) REFERENCES users(id),
 INDEX idx_orders_buyer(buyer_id,id), INDEX idx_orders_seller(seller_id,id), INDEX idx_orders_post(post_id,status)
) ENGINE=InnoDB;
CREATE TABLE order_events (
 id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT, order_id BIGINT UNSIGNED NOT NULL, actor_id BIGINT UNSIGNED NOT NULL,
 action VARCHAR(40) NOT NULL, created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
 FOREIGN KEY(order_id) REFERENCES orders(id), FOREIGN KEY(actor_id) REFERENCES users(id)
) ENGINE=InnoDB;
CREATE TABLE conversations (
 id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT, user_low BIGINT UNSIGNED NOT NULL, user_high BIGINT UNSIGNED NOT NULL,
 created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, UNIQUE KEY uq_participants(user_low,user_high),
 FOREIGN KEY(user_low) REFERENCES users(id), FOREIGN KEY(user_high) REFERENCES users(id)
) ENGINE=InnoDB;
CREATE TABLE messages (
 id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT, conversation_id BIGINT UNSIGNED NOT NULL, sender_id BIGINT UNSIGNED NOT NULL,
 content VARCHAR(2000) NOT NULL, read_at DATETIME NULL, created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
 FOREIGN KEY(conversation_id) REFERENCES conversations(id), FOREIGN KEY(sender_id) REFERENCES users(id), INDEX idx_messages_thread(conversation_id,id)
) ENGINE=InnoDB;
