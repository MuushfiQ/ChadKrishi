package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// This explicit upgrade only creates the missing table. It never drops data
// or guesses how to alter an incompatible existing schema.
const notificationDDL = `CREATE TABLE IF NOT EXISTS notifications (
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
) ENGINE=InnoDB;`

func repairNotifications(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := db.ExecContext(ctx, notificationDDL); err != nil {
		return fmt.Errorf("notification upgrade failed (requires migrations 001-007 and CREATE permission): %w", err)
	}
	return checkSchema(db)
}

func checkSchema(db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("DATABASE_DSN is not configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}
	checks := []struct{ table, columns string }{
		{"crop_evidence", "slug,crop_id,data"},
		{"garden_setup", "user_id,depth_cm,volume_l,drainage,support,method"},
		{"users", "id,is_admin,profile_photo"},
		{"posts", "id,sale_status,archived_at"},
		{"post_likes", "reaction"},
		{"orders", "id,post_id,buyer_id,seller_id,status,payment_method,payment_reference,buyer_paid_at,received_at"},
		{"order_events", "id,order_id,actor_id,action"},
		{"conversations", "id,user_low,user_high"},
		{"messages", "id,conversation_id,sender_id,content,read_at,created_at"},
		{"notifications", "id,user_id,actor_id,kind,title,link,conversation_id,message_id,read_at,created_at"},
	}
	for _, check := range checks {
		rows, err := db.QueryContext(ctx, "SELECT "+check.columns+" FROM "+check.table+" LIMIT 0")
		if err != nil {
			return fmt.Errorf("schema check failed for %s: %w. Apply the missing migration; for missing notifications run: go run . -repair-notifications; for crop_evidence/garden_setup run: go run . -upgrade-crops", check.table, err)
		}
		rows.Close()
		var engine string
		if err := db.QueryRowContext(ctx, "SELECT ENGINE FROM information_schema.TABLES WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME=?", check.table).Scan(&engine); err != nil {
			return err
		}
		if engine != "InnoDB" {
			return fmt.Errorf("table %s must use InnoDB for transactional safety", check.table)
		}
	}
	for _, column := range []struct {
		table, name string
		values      []string
	}{
		{"posts", "sale_status", []string{"available", "reserved", "sold"}},
		{"orders", "status", []string{"waiting_payment", "buyer_paid", "completed", "cancelled"}},
	} {
		var typ string
		if err := db.QueryRowContext(ctx, "SELECT COLUMN_TYPE FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME=? AND COLUMN_NAME=?", column.table, column.name).Scan(&typ); err != nil {
			return err
		}
		for _, value := range column.values {
			if !strings.Contains(typ, "'"+value+"'") {
				return fmt.Errorf("incompatible %s.%s: missing state %s; review migration 007 against your schema before altering", column.table, column.name, value)
			}
		}
	}

	return nil
}
