package main

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"strconv"
)

// Notification insertion shares the transaction with the triggering action.
// A failed notification insert rolls back the order/message, preventing silent loss.
func addNotification(c *gin.Context, tx *sql.Tx, recipient int64, kind, title, link string, conversation, message any) error {
	_, err := tx.ExecContext(c, `INSERT INTO notifications(user_id,actor_id,kind,title,link,conversation_id,message_id) VALUES(?,?,?,?,?,?,?)`, recipient, uid(c), kind, title, link, conversation, message)
	return err
}

func (a *app) notifications(c *gin.Context) {
	limit, before, ok := pageLimit(c)
	if !ok {
		return
	}
	var unread, latest int64
	if err := a.db.QueryRowContext(c, `SELECT COUNT(CASE WHEN read_at IS NULL THEN 1 END),COALESCE(MAX(id),0) FROM notifications WHERE user_id=?`, uid(c)).Scan(&unread, &latest); err != nil {
		dbError(c, err)
		return
	}
	rows, err := a.db.QueryContext(c, `SELECT n.id,n.kind,n.title,n.link,n.read_at IS NOT NULL,DATE_FORMAT(n.created_at,'%Y-%m-%dT%H:%i:%sZ'),u.name,COALESCE(u.profile_photo,'') FROM notifications n JOIN users u ON u.id=n.actor_id WHERE n.user_id=? AND (?=0 OR n.id<?) ORDER BY n.id DESC LIMIT ?`, uid(c), before, before, limit+1)
	if err != nil {
		dbError(c, err)
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id int64
		var kind, title, link, date, name, photo string
		var read bool
		if err = rows.Scan(&id, &kind, &title, &link, &read, &date, &name, &photo); err != nil {
			dbError(c, err)
			return
		}
		out = append(out, gin.H{"id": id, "kind": kind, "title": title, "link": link, "read": read, "createdAt": date, "actorName": name, "actorPhoto": photo})
	}
	if err = rows.Err(); err != nil {
		dbError(c, err)
		return
	}
	var next any
	if len(out) > limit {
		out = out[:limit]
		next = out[len(out)-1]["id"]
	}
	c.JSON(200, gin.H{"notifications": out, "unreadCount": unread, "latestId": latest, "nextCursor": next})
}

func (a *app) readNotification(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	var owned int64
	if err := a.db.QueryRowContext(c, "SELECT id FROM notifications WHERE id=? AND user_id=?", id, uid(c)).Scan(&owned); err != nil {
		dbError(c, err)
		return
	}
	if _, err := a.db.ExecContext(c, "UPDATE notifications SET read_at=COALESCE(read_at,UTC_TIMESTAMP()) WHERE id=? AND user_id=?", id, uid(c)); err != nil {
		dbError(c, err)
		return
	}
	c.Status(204)
}

func (a *app) readAllNotifications(c *gin.Context) {
	var b struct {
		Through int64 `json:"through"`
	}
	if !decodeJSON(c, &b) {
		return
	}
	if b.Through < 1 {
		writeError(c, 400, "last displayed notification is required")
		return
	}
	if _, err := a.db.ExecContext(c, "UPDATE notifications SET read_at=UTC_TIMESTAMP() WHERE user_id=? AND id<=? AND read_at IS NULL", uid(c), b.Through); err != nil {
		dbError(c, err)
		return
	}
	c.Status(204)
}

func orderNotification(action string, id int64) string {
	prefix := "Order #" + strconv.FormatInt(id, 10) + ": "
	switch action {
	case "paid":
		return prefix + "buyer marked Paid. Please check your payment."
	case "received":
		return prefix + "seller confirmed payment. Your purchase is complete."
	case "not_received":
		return prefix + "seller has not received payment. Please check."
	case "cancel":
		return prefix + "reservation cancelled."
	}
	return prefix + "updated."
}
