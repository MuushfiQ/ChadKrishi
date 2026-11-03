package main

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"strings"
	"time"
)

func validTaskDate(date string, now time.Time) bool {
	parsed, err := time.ParseInLocation("2006-01-02", date, dhaka)
	if err != nil {
		return false
	}
	day := now.In(dhaka).Format("2006-01-02")
	today, _ := time.ParseInLocation("2006-01-02", day, dhaka)
	return !parsed.Before(today.AddDate(-1, 0, 0)) && !parsed.After(today.AddDate(1, 0, 0))
}

func (a *app) createCustomTask(c *gin.Context) {
	var b struct {
		PlantID int64  `json:"plantId"`
		Title   string `json:"title"`
		Reason  string `json:"reason"`
		Date    string `json:"date"`
	}
	if !decodeJSON(c, &b) {
		return
	}
	if b.PlantID < 1 || !bounded(b.Title, 1, 160) || !bounded(b.Reason, 0, 300) || !validTaskDate(b.Date, time.Now()) {
		writeError(c, 400, "choose a plant, a title of 1–160 characters, notes up to 300 characters and a date within one year")
		return
	}
	tx, err := a.db.BeginTx(c, nil)
	if err != nil {
		dbError(c, err)
		return
	}
	defer tx.Rollback()
	var plantingDate string
	if err = tx.QueryRowContext(c, `SELECT DATE_FORMAT(planting_date,'%Y-%m-%d') FROM user_plants WHERE id=? AND user_id=? AND archived_at IS NULL FOR UPDATE`, b.PlantID, uid(c)).Scan(&plantingDate); err != nil {
		dbError(c, err)
		return
	}
	if b.Date < plantingDate {
		writeError(c, 400, "task date cannot be before the planting date")
		return
	}
	token := make([]byte, 16)
	if _, err = rand.Read(token); err != nil {
		writeError(c, 500, "could not create task")
		return
	}
	result, err := tx.ExecContext(c, `INSERT INTO care_tasks(plant_id,rule_key,title,reason,due_date) VALUES(?,?,?,?,?)`, b.PlantID, "custom_"+hex.EncodeToString(token), strings.TrimSpace(b.Title), strings.TrimSpace(b.Reason), b.Date)
	if err != nil {
		dbError(c, err)
		return
	}
	id, _ := result.LastInsertId()
	if err = tx.Commit(); err != nil {
		dbError(c, err)
		return
	}
	c.JSON(201, gin.H{"id": id})
}

func (a *app) deleteCustomTask(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	result, err := a.db.ExecContext(c, `DELETE t FROM care_tasks t JOIN user_plants p ON p.id=t.plant_id WHERE t.id=? AND p.user_id=? AND LEFT(t.rule_key,7)='custom_'`, id, uid(c))
	if err != nil {
		dbError(c, err)
		return
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		writeError(c, 404, "custom task not found")
		return
	}
	c.Status(204)
}
