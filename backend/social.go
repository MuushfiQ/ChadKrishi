package main

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

func (a *app) socialRoutes(r *gin.Engine) {
	g := r.Group("/api/v1", a.requireUser)
	g.PUT("/posts/:id/reaction", a.reactPost)
	g.POST("/posts/:id/buy", a.buy)
	g.GET("/orders", a.orders)
	g.PUT("/orders/:id", a.orderAction)
	g.GET("/admin/listings", a.adminListings)
	g.PUT("/admin/listings/:id/archive", a.archiveListing)
	g.GET("/conversations", a.conversations)
	g.POST("/conversations", a.startConversation)
	g.GET("/conversations/:id/messages", a.messages)
	g.POST("/conversations/:id/messages", a.sendMessage)
	g.PUT("/conversations/:id/read", a.readMessages)
	g.GET("/notifications", a.notifications)
	g.PUT("/notifications/read-all", a.readAllNotifications)
	g.PUT("/notifications/:id/read", a.readNotification)
}
func (a *app) reactPost(c *gin.Context) {
	id, ok := pathID(c)
	if !ok || !a.postExists(c, id) {
		return
	}
	var b struct {
		Reaction string `json:"reaction"`
	}
	if !decodeJSON(c, &b) {
		return
	}
	var e error
	switch b.Reaction {
	case "":
		_, e = a.db.ExecContext(c, "DELETE FROM post_likes WHERE post_id=? AND user_id=?", id, uid(c))
	case "like", "love", "angry", "sad":
		_, e = a.db.ExecContext(c, `INSERT INTO post_likes(post_id,user_id,reaction) VALUES(?,?,?) ON DUPLICATE KEY UPDATE reaction=VALUES(reaction)`, id, uid(c), b.Reaction)
	default:
		writeError(c, 400, "choose like, love, angry or sad")
		return
	}
	if e != nil {
		dbError(c, e)
		return
	}
	c.Status(204)
}

// Lock listing before order everywhere: competing buyers cannot reserve the same lot.
func (a *app) buy(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	tx, e := a.db.BeginTx(c, nil)
	if e != nil {
		dbError(c, e)
		return
	}
	defer tx.Rollback()
	var seller int64
	var title, typ, status, amount string
	var qty int
	var archived bool
	e = tx.QueryRowContext(c, `SELECT user_id,COALESCE(title,''),type,COALESCE(sale_status,''),COALESCE(price_bdt,0),COALESCE(quantity,0),archived_at IS NOT NULL FROM posts WHERE id=? FOR UPDATE`, id).Scan(&seller, &title, &typ, &status, &amount, &qty, &archived)
	if e != nil {
		dbError(c, e)
		return
	}
	if seller == uid(c) {
		writeError(c, 400, "you cannot buy your own listing")
		return
	}
	if typ != "sale" || archived {
		writeError(c, 404, "listing not found")
		return
	}
	if status != "available" {
		writeError(c, 409, "this listing is already reserved or sold")
		return
	}
	res, e := tx.ExecContext(c, `INSERT INTO orders(post_id,buyer_id,seller_id,title,amount_bdt,quantity) VALUES(?,?,?,?,?,?)`, id, uid(c), seller, title, amount, qty)
	if e != nil {
		dbError(c, e)
		return
	}
	oid, _ := res.LastInsertId()
	if _, e = tx.ExecContext(c, `UPDATE posts SET sale_status='reserved' WHERE id=?`, id); e != nil {
		dbError(c, e)
		return
	}
	if _, e = tx.ExecContext(c, `INSERT INTO order_events(order_id,actor_id,action) VALUES(?,?,'buy')`, oid, uid(c)); e != nil {
		dbError(c, e)
		return
	}
	if e = addNotification(c, tx, seller, "purchase", "New purchase request: "+title, "/orders?order="+strconv.FormatInt(oid, 10), nil, nil); e != nil {
		dbError(c, e)
		return
	}
	if e = tx.Commit(); e != nil {
		dbError(c, e)
		return
	}
	c.JSON(201, gin.H{"id": oid})
}
func nextOrderState(status, action string, buyer, seller bool) (string, bool) {
	switch action {
	case "paid":
		if buyer && status == "waiting_payment" {
			return "buyer_paid", true
		}
	case "received":
		if seller && status == "buyer_paid" {
			return "completed", true
		}
	case "cancel":
		if (buyer || seller) && status == "waiting_payment" {
			return "cancelled", true
		}
	case "not_received":
		if seller && status == "buyer_paid" {
			return "waiting_payment", true
		}
	}
	return "", false
}
func (a *app) orderAction(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	var b struct {
		Action    string `json:"action"`
		Method    string `json:"method"`
		Reference string `json:"reference"`
	}
	if !decodeJSON(c, &b) {
		return
	}
	if b.Action == "paid" && (b.Method != "cash" && b.Method != "bkash" && b.Method != "other" || !bounded(b.Reference, 0, 160)) {
		writeError(c, 400, "choose a payment method and reference up to 160 characters")
		return
	}
	var pid int64
	e := a.db.QueryRowContext(c, `SELECT post_id FROM orders WHERE id=? AND (buyer_id=? OR seller_id=?)`, id, uid(c), uid(c)).Scan(&pid)
	if e != nil {
		dbError(c, e)
		return
	}
	tx, e := a.db.BeginTx(c, nil)
	if e != nil {
		dbError(c, e)
		return
	}
	defer tx.Rollback()
	var locked int64
	if e = tx.QueryRowContext(c, "SELECT id FROM posts WHERE id=? FOR UPDATE", pid).Scan(&locked); e != nil {
		dbError(c, e)
		return
	}
	var buyer, seller int64
	var status string
	if e = tx.QueryRowContext(c, "SELECT buyer_id,seller_id,status FROM orders WHERE id=? FOR UPDATE", id).Scan(&buyer, &seller, &status); e != nil {
		dbError(c, e)
		return
	}
	next, allowed := nextOrderState(status, b.Action, buyer == uid(c), seller == uid(c))
	if !allowed {
		writeError(c, 409, "action not allowed for your role or current order state; refresh history")
		return
	}
	switch b.Action {
	case "paid":
		_, e = tx.ExecContext(c, `UPDATE orders SET status=?,payment_method=?,payment_reference=?,buyer_paid_at=UTC_TIMESTAMP() WHERE id=?`, next, b.Method, strings.TrimSpace(b.Reference), id)
	case "received":
		_, e = tx.ExecContext(c, `UPDATE orders SET status=?,received_at=UTC_TIMESTAMP() WHERE id=?`, next, id)
	default:
		_, e = tx.ExecContext(c, "UPDATE orders SET status=? WHERE id=?", next, id)
	}
	if e != nil {
		dbError(c, e)
		return
	}
	sale := "reserved"
	if next == "completed" {
		sale = "sold"
	}
	if next == "cancelled" {
		sale = "available"
	}
	if _, e = tx.ExecContext(c, "UPDATE posts SET sale_status=? WHERE id=?", sale, pid); e != nil {
		dbError(c, e)
		return
	}
	if _, e = tx.ExecContext(c, "INSERT INTO order_events(order_id,actor_id,action) VALUES(?,?,?)", id, uid(c), b.Action); e != nil {
		dbError(c, e)
		return
	}
	recipient := buyer
	if uid(c) == buyer {
		recipient = seller
	}
	if e = addNotification(c, tx, recipient, "order_"+b.Action, orderNotification(b.Action, id), "/orders?order="+strconv.FormatInt(id, 10), nil, nil); e != nil {
		dbError(c, e)
		return
	}
	if e = tx.Commit(); e != nil {
		dbError(c, e)
		return
	}
	c.JSON(200, gin.H{"status": next})
}
func (a *app) orders(c *gin.Context) {
	limit, before, ok := pageLimit(c)
	if !ok {
		return
	}
	var selected int64
	if c.Query("id") != "" {
		var err error
		selected, err = strconv.ParseInt(c.Query("id"), 10, 64)
		if err != nil || selected < 1 {
			writeError(c, 400, "invalid order ID")
			return
		}
	}
	rows, e := a.db.QueryContext(c, `SELECT o.id,o.post_id,o.buyer_id,o.seller_id,o.title,o.amount_bdt,o.quantity,o.status,COALESCE(o.payment_method,''),o.payment_reference,DATE_FORMAT(o.created_at,'%Y-%m-%dT%H:%i:%sZ'),b.name,s.name,p.archived_at IS NOT NULL FROM orders o JOIN users b ON b.id=o.buyer_id JOIN users s ON s.id=o.seller_id JOIN posts p ON p.id=o.post_id WHERE (o.buyer_id=? OR o.seller_id=?) AND (?=0 OR o.id=?) AND (?=0 OR o.id<?) ORDER BY o.id DESC LIMIT ?`, uid(c), uid(c), selected, selected, before, before, limit+1)
	if e != nil {
		dbError(c, e)
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id, pid, buyer, seller int64
		var title, amount, status, method, ref, date, bn, sn string
		var qty int
		var archived bool
		if e = rows.Scan(&id, &pid, &buyer, &seller, &title, &amount, &qty, &status, &method, &ref, &date, &bn, &sn, &archived); e != nil {
			dbError(c, e)
			return
		}
		out = append(out, gin.H{"id": id, "postId": pid, "buyerId": buyer, "sellerId": seller, "title": title, "amountBdt": amount, "quantity": qty, "status": status, "method": method, "reference": ref, "createdAt": date, "buyerName": bn, "sellerName": sn, "isBuyer": buyer == uid(c), "archived": archived})
	}
	if e = rows.Err(); e != nil {
		dbError(c, e)
		return
	}
	var next any
	if len(out) > limit {
		out = out[:limit]
		next = out[len(out)-1]["id"]
	}
	c.JSON(200, gin.H{"orders": out, "nextCursor": next})
}
func (a *app) requireAdmin(c *gin.Context) bool {
	var admin bool
	e := a.db.QueryRowContext(c, "SELECT is_admin FROM users WHERE id=?", uid(c)).Scan(&admin)
	if e != nil {
		dbError(c, e)
		return false
	}
	if !admin {
		writeError(c, 403, "administrator access required")
		return false
	}
	return true
}
func (a *app) adminListings(c *gin.Context) {
	if !a.requireAdmin(c) {
		return
	}
	limit, before, ok := pageLimit(c)
	if !ok {
		return
	}
	rows, e := a.db.QueryContext(c, `SELECT p.id,COALESCE(p.title,''),u.name,p.archived_at IS NOT NULL FROM posts p JOIN users u ON u.id=p.user_id WHERE p.type='sale' AND p.sale_status='sold' AND (?=0 OR p.id<?) ORDER BY p.id DESC LIMIT ?`, before, before, limit+1)
	if e != nil {
		dbError(c, e)
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id int64
		var title, name string
		var archived bool
		if e = rows.Scan(&id, &title, &name, &archived); e != nil {
			dbError(c, e)
			return
		}
		out = append(out, gin.H{"id": id, "title": title, "sellerName": name, "archived": archived})
	}
	if e = rows.Err(); e != nil {
		dbError(c, e)
		return
	}
	var next any
	if len(out) > limit {
		out = out[:limit]
		next = out[len(out)-1]["id"]
	}
	c.JSON(200, gin.H{"listings": out, "nextCursor": next})
}
func (a *app) archiveListing(c *gin.Context) {
	if !a.requireAdmin(c) {
		return
	}
	id, ok := pathID(c)
	if !ok {
		return
	}
	_, e := a.db.ExecContext(c, `UPDATE posts SET archived_at=COALESCE(archived_at,UTC_TIMESTAMP()) WHERE id=? AND type='sale' AND sale_status='sold'`, id)
	if e != nil {
		dbError(c, e)
		return
	}
	var archived bool
	e = a.db.QueryRowContext(c, "SELECT archived_at IS NOT NULL FROM posts WHERE id=? AND sale_status='sold'", id).Scan(&archived)
	if e != nil || !archived {
		writeError(c, 409, "only sold listings can be archived")
		return
	}
	c.Status(204)
}
func (a *app) startConversation(c *gin.Context) {
	var b struct {
		UserID int64 `json:"userId"`
	}
	if !decodeJSON(c, &b) {
		return
	}
	if b.UserID < 1 || b.UserID == uid(c) {
		writeError(c, 400, "choose another member")
		return
	}
	var n int64
	if e := a.db.QueryRowContext(c, "SELECT id FROM users WHERE id=?", b.UserID).Scan(&n); e != nil {
		dbError(c, e)
		return
	}
	low, high := uid(c), b.UserID
	if low > high {
		low, high = high, low
	}
	_, e := a.db.ExecContext(c, `INSERT INTO conversations(user_low,user_high) VALUES(?,?) ON DUPLICATE KEY UPDATE id=id`, low, high)
	if e != nil {
		dbError(c, e)
		return
	}
	if e = a.db.QueryRowContext(c, "SELECT id FROM conversations WHERE user_low=? AND user_high=?", low, high).Scan(&n); e != nil {
		dbError(c, e)
		return
	}
	c.JSON(200, gin.H{"id": n})
}
func (a *app) threadAccess(c *gin.Context, id int64) bool {
	var n int64
	e := a.db.QueryRowContext(c, "SELECT id FROM conversations WHERE id=? AND (user_low=? OR user_high=?)", id, uid(c), uid(c)).Scan(&n)
	if e != nil {
		dbError(c, e)
		return false
	}
	return true
}
func (a *app) conversations(c *gin.Context) {
	limit, before, ok := pageLimit(c)
	if !ok {
		return
	}
	rows, e := a.db.QueryContext(c, `SELECT v.id,u.id,u.name,COALESCE(u.profile_photo,''),COALESCE((SELECT content FROM messages WHERE conversation_id=v.id ORDER BY id DESC LIMIT 1),''),(SELECT COUNT(*) FROM messages WHERE conversation_id=v.id AND sender_id<>? AND read_at IS NULL) FROM conversations v JOIN users u ON u.id=IF(v.user_low=?,v.user_high,v.user_low) WHERE (v.user_low=? OR v.user_high=?) AND (?=0 OR v.id<?) ORDER BY v.id DESC LIMIT ?`, uid(c), uid(c), uid(c), uid(c), before, before, limit+1)
	if e != nil {
		dbError(c, e)
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id, user, unread int64
		var name, photo, last string
		if e = rows.Scan(&id, &user, &name, &photo, &last, &unread); e != nil {
			dbError(c, e)
			return
		}
		out = append(out, gin.H{"id": id, "userId": user, "name": name, "profilePhoto": photo, "lastMessage": last, "unread": unread})
	}
	if e = rows.Err(); e != nil {
		dbError(c, e)
		return
	}
	var next any
	if len(out) > limit {
		out = out[:limit]
		next = out[len(out)-1]["id"]
	}
	c.JSON(200, gin.H{"conversations": out, "nextCursor": next})
}
func (a *app) messages(c *gin.Context) {
	id, ok := pathID(c)
	if !ok || !a.threadAccess(c, id) {
		return
	}
	before := int64(0)
	if c.Query("before") != "" {
		var e error
		before, e = strconv.ParseInt(c.Query("before"), 10, 64)
		if e != nil || before < 1 {
			writeError(c, 400, "invalid cursor")
			return
		}
	}
	rows, e := a.db.QueryContext(c, `SELECT id,sender_id,content,DATE_FORMAT(created_at,'%Y-%m-%dT%H:%i:%sZ'),read_at IS NOT NULL FROM messages WHERE conversation_id=? AND (?=0 OR id<?) ORDER BY id DESC LIMIT 51`, id, before, before)
	if e != nil {
		dbError(c, e)
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var mid, sender int64
		var txt, date string
		var read bool
		if e = rows.Scan(&mid, &sender, &txt, &date, &read); e != nil {
			dbError(c, e)
			return
		}
		out = append(out, gin.H{"id": mid, "content": txt, "mine": sender == uid(c), "createdAt": date, "read": read})
	}
	if e = rows.Err(); e != nil {
		dbError(c, e)
		return
	}
	var next any
	if len(out) > 50 {
		out = out[:50]
		next = out[len(out)-1]["id"]
	}
	c.JSON(200, gin.H{"messages": out, "nextCursor": next})
}
func (a *app) sendMessage(c *gin.Context) {
	id, ok := pathID(c)
	if !ok || !a.threadAccess(c, id) {
		return
	}
	var b struct {
		Content string `json:"content"`
	}
	if !decodeJSON(c, &b) {
		return
	}
	if !bounded(b.Content, 1, 2000) {
		writeError(c, 400, "message needs 1–2000 characters")
		return
	}
	tx, e := a.db.BeginTx(c, nil)
	if e != nil {
		dbError(c, e)
		return
	}
	defer tx.Rollback()
	var recipient int64
	if e = tx.QueryRowContext(c, "SELECT IF(user_low=?,user_high,user_low) FROM conversations WHERE id=? AND (user_low=? OR user_high=?)", uid(c), id, uid(c), uid(c)).Scan(&recipient); e != nil {
		dbError(c, e)
		return
	}
	res, e := tx.ExecContext(c, "INSERT INTO messages(conversation_id,sender_id,content) VALUES(?,?,?)", id, uid(c), strings.TrimSpace(b.Content))
	if e != nil {
		dbError(c, e)
		return
	}
	mid, _ := res.LastInsertId()
	if e = addNotification(c, tx, recipient, "message", "Sent you a new message", "/messages?thread="+strconv.FormatInt(id, 10), id, mid); e != nil {
		dbError(c, e)
		return
	}
	if e = tx.Commit(); e != nil {
		dbError(c, e)
		return
	}

	c.JSON(201, gin.H{"id": mid})
}
func (a *app) readMessages(c *gin.Context) {
	id, ok := pathID(c)
	if !ok || !a.threadAccess(c, id) {
		return
	}
	var b struct {
		Through int64 `json:"through"`
	}
	if !decodeJSON(c, &b) {
		return
	}
	if b.Through < 1 {
		writeError(c, 400, "last displayed message is required")
		return
	}
	tx, e := a.db.BeginTx(c, nil)
	if e != nil {
		dbError(c, e)
		return
	}
	defer tx.Rollback()
	if _, e = tx.ExecContext(c, `UPDATE messages SET read_at=UTC_TIMESTAMP() WHERE conversation_id=? AND sender_id<>? AND id<=? AND read_at IS NULL`, id, uid(c), b.Through); e != nil {
		dbError(c, e)
		return
	}
	if _, e = tx.ExecContext(c, `UPDATE notifications SET read_at=UTC_TIMESTAMP() WHERE user_id=? AND conversation_id=? AND message_id<=? AND read_at IS NULL`, uid(c), id, b.Through); e != nil {
		dbError(c, e)
		return
	}
	if e = tx.Commit(); e != nil {
		dbError(c, e)
		return
	}

	c.Status(204)
}
