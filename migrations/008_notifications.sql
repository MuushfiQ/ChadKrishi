-- Apply after migrations 001–007, to the existing database.
USE defaultdb;
CREATE TABLE IF NOT EXISTS notifications (
 id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
 user_id BIGINT UNSIGNED NOT NULL,
 actor_id BIGINT UNSIGNED NOT NULL,
 kind VARCHAR(40) NOT NULL,
 title VARCHAR(255) NOT NULL,
 link VARCHAR(255) NOT NULL,
 conversation_id BIGINT UNSIGNED NULL,
 message_id BIGINT UNSIGNED NULL,
 read_at DATETIME NULL,
 created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
 FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
 FOREIGN KEY(actor_id) REFERENCES users(id) ON DELETE CASCADE,
 FOREIGN KEY(conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
 FOREIGN KEY(message_id) REFERENCES messages(id) ON DELETE CASCADE,
 UNIQUE KEY uq_message_recipient(user_id,message_id),
 INDEX idx_notification_feed(user_id,id),
 INDEX idx_notification_unread(user_id,read_at,id)
) ENGINE=InnoDB;
