package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// These routes share the same session and origin guard as the existing API.
func (a *app) gardenRoutes(r *gin.Engine) {
	g := r.Group("/api/v1", a.requireUser)
	g.GET("/crops", a.catalog)
	g.GET("/garden-setup", a.getGardenSetup)
	g.PUT("/garden-setup", a.saveGardenSetup)
	g.GET("/plants", a.plants)
	g.POST("/plants", a.savePlant)
	g.PUT("/plants/:id", a.savePlant)
	g.GET("/tasks", a.tasks)
	g.POST("/tasks", a.createCustomTask)
	g.DELETE("/tasks/:id", a.deleteCustomTask)
	g.PUT("/tasks/:id", a.completeTask)
	g.POST("/uploads", a.upload)
	g.PUT("/profile/photo", a.updateProfilePhoto)
	g.GET("/posts", a.posts)
	g.POST("/posts", a.savePost)
	g.PUT("/posts/:id", a.savePost)
	g.DELETE("/posts/:id", a.deletePost)
	g.PUT("/posts/:id/status", a.saleStatus)
	g.PUT("/posts/:id/like", a.likePost)
	g.GET("/posts/:id/comments", a.comments)
	g.POST("/posts/:id/comments", a.addComment)
	g.DELETE("/comments/:id", a.deleteComment)
	g.GET("/people/:id", a.publicProfile)
	r.GET("/api/media/:name", a.requireUser, a.media)
}
func uid(c *gin.Context) int64 { return c.MustGet("userID").(int64) }
func pathID(c *gin.Context) (int64, bool) {
	n, e := strconv.ParseInt(c.Param("id"), 10, 64)
	if e != nil || n < 1 {
		writeError(c, 400, "invalid ID")
		return 0, false
	}
	return n, true
}
func bounded(s string, min, max int) bool {
	n := utf8.RuneCountInString(strings.TrimSpace(s))
	return n >= min && n <= max
}
func dbError(c *gin.Context, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		writeError(c, 404, "item not found")
	} else {
		var dbErr *mysql.MySQLError
		if errors.As(err, &dbErr) {
			// Never log SQL values, messages, payment references, or credentials.
			log.Printf("database failure method=%s route=%s mysql_code=%d", c.Request.Method, c.FullPath(), dbErr.Number)
			if dbErr.Number == 1146 || dbErr.Number == 1054 {
				writeError(c, 503, "Database setup is incomplete. Ask the administrator to run the database check and apply the missing migration.")
				return
			}
		} else {
			log.Printf("database failure method=%s route=%s type=%T", c.Request.Method, c.FullPath(), err)
		}
		writeError(c, 500, "Database operation failed. No change was confirmed. Please contact the administrator.")
	}
}
func (a *app) owns(c *gin.Context, table string, id int64) bool {
	// table is always a server-side constant, never request input.
	var n int64
	err := a.db.QueryRowContext(c, "SELECT id FROM "+table+" WHERE id=? AND user_id=?", id, uid(c)).Scan(&n)
	if err != nil {
		dbError(c, err)
		return false
	}
	return true
}
func (a *app) catalog(c *gin.Context) {
	rows, e := a.db.QueryContext(c, `SELECT c.id,c.name,COALESCE(JSON_UNQUOTE(JSON_EXTRACT(e.data,'$.bengaliName')),c.bengali_name,''),CASE WHEN e.crop_id IS NULL THEN c.verification_status WHEN JSON_UNQUOTE(JSON_EXTRACT(e.data,'$.coreVerified'))='true' THEN 'core-reviewed; check setup' ELSE 'evidence-incomplete' END,COALESCE(JSON_UNQUOTE(JSON_EXTRACT(e.data,'$.containerNote')),c.container_guidance) FROM crops c LEFT JOIN crop_evidence e ON e.crop_id=c.id ORDER BY c.name`)
	if e != nil {
		dbError(c, e)
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id int64
		var n, b, s, g string
		if e = rows.Scan(&id, &n, &b, &s, &g); e != nil {
			dbError(c, e)
			return
		}
		out = append(out, gin.H{"id": id, "name": n, "bengaliName": b, "verificationStatus": s, "containerGuidance": g})
	}
	if e = rows.Err(); e != nil {
		dbError(c, e)
		return
	}
	c.JSON(200, gin.H{"crops": out})
}

type plantInput struct {
	CropID            int64  `json:"cropId"`
	Nickname          string `json:"nickname"`
	PlantingDate      string `json:"plantingDate"`
	PropagationMethod string `json:"propagationMethod"`
	GrowthStage       string `json:"growthStage"`
	PhotoPath         string `json:"photoPath"`
	Note              string `json:"note"`
	Archived          bool   `json:"archived"`
}

func validatePlant(p plantInput, now time.Time) error {
	d, e := time.Parse("2006-01-02", p.PlantingDate)
	today := now.In(dhaka).Format("2006-01-02")
	if e != nil || d.Format("2006-01-02") > today || d.Year() < 1900 {
		return errors.New("choose a planting date from 1900 through today")
	}
	if p.CropID < 1 || !bounded(p.Nickname, 1, 100) || !bounded(p.Note, 0, 500) {
		return errors.New("choose a crop, a name (1–100 characters), and notes up to 500 characters")
	}
	if p.PropagationMethod != "seed" && p.PropagationMethod != "seedling" {
		return errors.New("choose seed or seedling")
	}
	switch p.GrowthStage {
	case "seedling", "growing", "flowering", "fruiting_harvest":
	default:
		return errors.New("invalid growth stage")
	}
	return nil
}
func (a *app) savePlant(c *gin.Context) {
	var p plantInput
	if !decodeJSON(c, &p) {
		return
	}
	if e := validatePlant(p, time.Now()); e != nil {
		writeError(c, 400, e.Error())
		return
	}
	id := int64(0)
	if c.Request.Method == http.MethodPut {
		var ok bool
		id, ok = pathID(c)
		if !ok || !a.owns(c, "user_plants", id) {
			return
		}
	}
	if !a.validImage(c, p.PhotoPath) {
		return
	}
	var crop int64
	if e := a.db.QueryRowContext(c, "SELECT id FROM crops WHERE id=?", p.CropID).Scan(&crop); e != nil {
		dbError(c, e)
		return
	}
	var archived any
	if p.Archived {
		archived = time.Now().UTC()
	}
	if id == 0 {
		res, e := a.db.ExecContext(c, `INSERT INTO user_plants(user_id,crop_id,nickname,planting_date,propagation_method,growth_stage,photo_path,note) VALUES(?,?,?,?,?,?,?,?)`, uid(c), p.CropID, strings.TrimSpace(p.Nickname), p.PlantingDate, p.PropagationMethod, p.GrowthStage, p.PhotoPath, strings.TrimSpace(p.Note))
		if e != nil {
			dbError(c, e)
			return
		}
		id, _ = res.LastInsertId()
		c.JSON(201, gin.H{"id": id})
		return
	}
	_, e := a.db.ExecContext(c, `UPDATE user_plants SET crop_id=?,nickname=?,planting_date=?,propagation_method=?,growth_stage=?,photo_path=?,note=?,archived_at=? WHERE id=? AND user_id=?`, p.CropID, strings.TrimSpace(p.Nickname), p.PlantingDate, p.PropagationMethod, p.GrowthStage, p.PhotoPath, strings.TrimSpace(p.Note), archived, id, uid(c))
	if e != nil {
		dbError(c, e)
		return
	}
	c.JSON(200, gin.H{"id": id})
}
func (a *app) plants(c *gin.Context) {
	rows, e := a.db.QueryContext(c, `SELECT p.id,p.crop_id,p.nickname,DATE_FORMAT(p.planting_date,'%Y-%m-%d'),p.propagation_method,p.growth_stage,COALESCE(p.photo_path,''),COALESCE(p.note,''),p.archived_at IS NOT NULL,c.name,COALESCE(JSON_UNQUOTE(JSON_EXTRACT(e.data,'$.bengaliName')),c.bengali_name,''),CASE WHEN e.crop_id IS NULL THEN c.verification_status WHEN JSON_UNQUOTE(JSON_EXTRACT(e.data,'$.coreVerified'))='true' THEN 'core-reviewed; check setup' ELSE 'evidence-incomplete' END,COALESCE(JSON_UNQUOTE(JSON_EXTRACT(e.data,'$.careNote')),c.care_guidance) FROM user_plants p JOIN crops c ON c.id=p.crop_id LEFT JOIN crop_evidence e ON e.crop_id=c.id WHERE p.user_id=? ORDER BY p.archived_at IS NOT NULL,p.created_at DESC,p.id DESC`, uid(c))
	if e != nil {
		dbError(c, e)
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id, crop int64
		var n, d, m, s, ph, note, cn, b, ver, care string
		var archived bool
		if e = rows.Scan(&id, &crop, &n, &d, &m, &s, &ph, &note, &archived, &cn, &b, &ver, &care); e != nil {
			dbError(c, e)
			return
		}
		out = append(out, gin.H{"id": id, "cropId": crop, "nickname": n, "plantingDate": d, "propagationMethod": m, "growthStage": s, "photoPath": ph, "note": note, "archived": archived, "cropName": cn, "bengaliName": b, "verificationStatus": ver, "careGuidance": care})
	}
	if e = rows.Err(); e != nil {
		dbError(c, e)
		return
	}
	c.JSON(200, gin.H{"plants": out})
}

var dhaka = time.FixedZone("Asia/Dhaka", 6*60*60)

type careRule struct{ key, title, reason string }

// Observation prompts, not dosing instructions or disease diagnoses.
func routineRules(stage string) []careRule {
	r := []careRule{{"soil_check", "Check soil moisture and drainage", "Observe the pot first. Water only if the growing medium needs it; make sure drainage holes are clear."}, {"leaf_check", "Look at leaves and stems", "Record visible changes or pests. This checklist cannot diagnose plant disease."}}
	switch stage {
	case "seedling":
		r = append(r, careRule{"seedling_check", "Check seedling support and exposure", "Young plants need observation. Adjust shelter if the plant shows stress; do not infer growth from age alone."})
	case "flowering", "fruiting_harvest":
		r = append(r, careRule{"support_check", "Inspect supports and developing flowers or fruit", "Check that branches and supports are secure. Update the recorded stage to match what you see."})
	}
	return r
}
func (a *app) tasks(c *gin.Context) {
	date := c.DefaultQuery("date", time.Now().In(dhaka).Format("2006-01-02"))
	var e error
	today := time.Now().In(dhaka).Format("2006-01-02")
	if !validTaskDate(date, time.Now()) {
		writeError(c, 400, "choose a date within one year of today")
		return
	}
	var weather weatherResponse
	var rain bool
	if date == today {
		var district sql.NullString
		var exposed sql.NullBool
		if e = a.db.QueryRowContext(c, "SELECT district,rain_exposure FROM users WHERE id=?", uid(c)).Scan(&district, &exposed); e != nil {
			dbError(c, e)
			return
		}
		rain = exposed.Valid && exposed.Bool
		if loc, ok := locationForDistrict(district.String); ok && a.weather != nil {
			weather = a.weather.get(c, loc)
		}
		// Close the read cursor before inserts, also when the DB pool has one connection.
		rows, err := a.db.QueryContext(c, `SELECT p.id,p.growth_stage,COALESCE(JSON_UNQUOTE(JSON_EXTRACT(e.data,'$.careNote')),'') FROM user_plants p LEFT JOIN crop_evidence e ON e.crop_id=p.crop_id WHERE p.user_id=? AND p.archived_at IS NULL AND p.planting_date<=?`, uid(c), today)
		if err != nil {
			dbError(c, err)
			return
		}
		type item struct {
			id    int64
			stage string
			care  string
		}
		ps := []item{}
		for rows.Next() {
			var p item
			if err = rows.Scan(&p.id, &p.stage, &p.care); err != nil {
				rows.Close()
				dbError(c, err)
				return
			}
			ps = append(ps, p)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			dbError(c, err)
			return
		}
		for _, p := range ps {
			rules := routineRules(p.stage)
			if p.care != "" {
				rules = append(rules, careRule{"crop_observation", "Review this crop’s care observations", p.care})
			}
			if weather.Status == "available" && len(weather.Forecast) > 0 && weather.Forecast[0].Date == today && rain && weather.Forecast[0].PrecipitationSum >= 10 {
				rules = append(rules, careRule{"rain_check", "Check exposed pots before forecast rain", "At least 10 mm is forecast today. App precaution threshold: check drainage and shelter; actual rainfall at the pot is unknown."})
			}
			if weather.Status == "available" && weather.Current != nil && weather.Current.WindSpeed >= 40 {
				rules = append(rules, careRule{"wind_check", "Check pots and supports in wind", "Area wind is at least 40 km/h. App precaution threshold: secure loose pots and supports if safe."})
			}
			for _, r := range rules {
				_, err = a.db.ExecContext(c, `INSERT INTO care_tasks(plant_id,rule_key,title,reason,due_date) VALUES(?,?,?,?,?) ON DUPLICATE KEY UPDATE id=id`, p.id, r.key, r.title, r.reason, today)
				if err != nil {
					dbError(c, err)
					return
				}
			}
		}
	}
	rows, e := a.db.QueryContext(c, `SELECT t.id,t.plant_id,p.nickname,t.title,t.reason,t.completed_at IS NOT NULL,LEFT(t.rule_key,7)='custom_',cr.name FROM care_tasks t JOIN user_plants p ON p.id=t.plant_id JOIN crops cr ON cr.id=p.crop_id WHERE p.user_id=? AND t.due_date=? AND (p.archived_at IS NULL OR ?<?) ORDER BY p.id,t.id`, uid(c), date, date, today)
	if e != nil {
		dbError(c, e)
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id, pid int64
		var n, t, r, crop string
		var done, custom bool
		if e = rows.Scan(&id, &pid, &n, &t, &r, &done, &custom, &crop); e != nil {
			dbError(c, e)
			return
		}
		out = append(out, gin.H{"id": id, "plantId": pid, "plantName": n, "title": t, "reason": r, "completed": done, "custom": custom, "cropName": crop})
	}
	if e = rows.Err(); e != nil {
		dbError(c, e)
		return
	}
	c.JSON(200, gin.H{"date": date, "tasks": out, "guidance": "General observation checklist; not crop-specific dosing. Weather precautions use app thresholds, not official alerts.", "weatherStatus": weather.Status})
}
func (a *app) completeTask(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	var body struct {
		Completed *bool `json:"completed"`
	}
	if !decodeJSON(c, &body) {
		return
	}
	if body.Completed == nil {
		writeError(c, 400, "completed is required")
		return
	}
	var task int64
	e := a.db.QueryRowContext(c, `SELECT t.id FROM care_tasks t JOIN user_plants p ON p.id=t.plant_id WHERE t.id=? AND p.user_id=?`, id, uid(c)).Scan(&task)
	if e != nil {
		dbError(c, e)
		return
	}
	var at any
	if *body.Completed {
		at = time.Now().UTC()
	}
	_, e = a.db.ExecContext(c, "UPDATE care_tasks SET completed_at=? WHERE id=?", at, id)
	if e != nil {
		dbError(c, e)
		return
	}
	c.Status(204)
}

func uploadDir() string              { return envOr("UPLOAD_DIR", "uploads") }
func uploadPrefix(user int64) string { return strconv.FormatInt(user, 10) + "-" }
func (a *app) validImage(c *gin.Context, p string) bool {
	if p == "" {
		return true
	}
	name := strings.TrimPrefix(p, "/api/media/")
	if p != "/api/media/"+name || name != filepath.Base(name) || strings.ContainsAny(name, "/\\") || !strings.HasPrefix(name, uploadPrefix(uid(c))) {
		writeError(c, 400, "select an image uploaded by your account")
		return false
	}
	if _, e := os.Stat(filepath.Join(uploadDir(), name)); e != nil {
		writeError(c, 400, "uploaded image was not found")
		return false
	}
	return true
}
func (a *app) upload(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 5*1024*1024+65536)
	if e := c.Request.ParseMultipartForm(5 * 1024 * 1024); e != nil {
		writeError(c, 400, "choose a JPEG or PNG image up to 5 MB")
		return
	}
	defer c.Request.MultipartForm.RemoveAll()
	f, h, e := c.Request.FormFile("image")
	if e != nil {
		writeError(c, 400, "image file is required")
		return
	}
	defer f.Close()
	if h.Size > 5*1024*1024 {
		writeError(c, 400, "image must be at most 5 MB")
		return
	}
	cfg, format, e := image.DecodeConfig(f)
	if e != nil || (format != "jpeg" && format != "png") || cfg.Width < 1 || cfg.Height < 1 || int64(cfg.Width)*int64(cfg.Height) > 20000000 {
		writeError(c, 400, "use a valid JPEG/PNG of at most 20 megapixels")
		return
	}
	if _, e = f.Seek(0, 0); e != nil {
		writeError(c, 400, "could not read image")
		return
	}
	decoded, _, e := image.Decode(f)
	if e != nil {
		writeError(c, 400, "image is damaged")
		return
	}
	f.Seek(0, 0)
	b := make([]byte, 16)
	if _, e = rand.Read(b); e != nil {
		writeError(c, 500, "could not name image")
		return
	}
	ext := ".jpg"
	if format == "png" {
		ext = ".png"
	}
	name := uploadPrefix(uid(c)) + hex.EncodeToString(b) + ext
	if e = os.MkdirAll(uploadDir(), 0700); e != nil {
		writeError(c, 500, "upload storage unavailable")
		return
	}
	dest, e := os.OpenFile(filepath.Join(uploadDir(), name), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		writeError(c, 500, "could not save image")
		return
	}
	if format == "png" {
		e = png.Encode(dest, decoded)
	} else {
		e = jpeg.Encode(dest, decoded, &jpeg.Options{Quality: 90})
	}
	ce := dest.Close()
	if e != nil || ce != nil {
		os.Remove(filepath.Join(uploadDir(), name))
		writeError(c, 500, "could not save image")
		return
	}
	c.JSON(201, gin.H{"path": "/api/media/" + name})
}
func (a *app) media(c *gin.Context) {
	n := c.Param("name")
	if n != filepath.Base(n) || strings.ContainsAny(n, "/\\") || (filepath.Ext(n) != ".jpg" && filepath.Ext(n) != ".png") {
		c.Status(404)
		return
	}
	// Orphan uploads and private plant photos cannot be fetched by other members.
	if !strings.HasPrefix(n, uploadPrefix(uid(c))) {
		var count int
		e := a.db.QueryRowContext(c, `SELECT (SELECT COUNT(*) FROM posts WHERE archived_at IS NULL AND image_path=?) + (SELECT COUNT(*) FROM users WHERE profile_photo=?)`, "/api/media/"+n, "/api/media/"+n).Scan(&count)
		if e != nil {
			dbError(c, e)
			return
		}
		if count == 0 {
			c.Status(404)
			return
		}
	}
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Cache-Control", "private, no-store")
	c.File(filepath.Join(uploadDir(), n))
}

// A listing is a community post, so both screens always show the same sale state.
type postInput struct {
	Type          string   `json:"type"`
	Title         string   `json:"title"`
	Content       string   `json:"content"`
	ImagePath     string   `json:"imagePath"`
	PlantID       *int64   `json:"plantId"`
	PriceBDT      *float64 `json:"priceBdt"`
	Quantity      *int     `json:"quantity"`
	ContactMethod string   `json:"contactMethod"`
}

func validatePost(p postInput) error {
	switch p.Type {
	case "update", "question", "blog", "sale":
	default:
		return errors.New("invalid post type")
	}
	if !bounded(p.Content, 1, 5000) || !bounded(p.Title, 0, 160) || !bounded(p.ContactMethod, 0, 160) {
		return errors.New("content is required (up to 5000 characters); title/contact up to 160")
	}
	if p.Type == "sale" && (p.PriceBDT == nil || *p.PriceBDT < 0 || *p.PriceBDT > 99999999.99 || p.Quantity == nil || *p.Quantity < 1 || *p.Quantity > 1000000 || !bounded(p.Title, 1, 160)) {
		return errors.New("sale needs a title, nonnegative BDT price and quantity 1–1000000")
	}
	return nil
}
func (a *app) savePost(c *gin.Context) {
	var p postInput
	if !decodeJSON(c, &p) {
		return
	}
	if e := validatePost(p); e != nil {
		writeError(c, 400, e.Error())
		return
	}
	id := int64(0)
	var editTx *sql.Tx
	if c.Request.Method == http.MethodPut {
		var ok bool
		id, ok = pathID(c)
		if !ok || !a.owns(c, "posts", id) {
			return
		}
		var err error
		editTx, err = a.db.BeginTx(c, nil)
		if err != nil {
			dbError(c, err)
			return
		}
		defer editTx.Rollback()
		var typ, saleState string
		if e := editTx.QueryRowContext(c, "SELECT type,COALESCE(sale_status,'') FROM posts WHERE id=? AND archived_at IS NULL FOR UPDATE", id).Scan(&typ, &saleState); e != nil {
			dbError(c, e)
			return
		}
		var locked int
		if e := editTx.QueryRowContext(c, "SELECT COUNT(*) FROM orders WHERE post_id=? AND status<>'cancelled'", id).Scan(&locked); e != nil {
			dbError(c, e)
			return
		}
		if locked > 0 || (typ == "sale" && saleState != "available") {
			writeError(c, 409, "reserved or completed listings cannot be edited")
			return
		}
		if typ != p.Type {
			writeError(c, 400, "post type cannot be changed after publishing")
			return
		}
	}
	if !a.validImage(c, p.ImagePath) {
		return
	}
	var district sql.NullString
	if e := a.db.QueryRowContext(c, "SELECT district FROM users WHERE id=?", uid(c)).Scan(&district); e != nil {
		dbError(c, e)
		return
	}
	if p.Type == "sale" && !district.Valid {
		writeError(c, 422, "set your general area in My profile before selling")
		return
	}
	if p.PlantID != nil {
		if !a.owns(c, "user_plants", *p.PlantID) {
			return
		}
		var archived bool
		if e := a.db.QueryRowContext(c, "SELECT archived_at IS NOT NULL FROM user_plants WHERE id=?", *p.PlantID).Scan(&archived); e != nil {
			dbError(c, e)
			return
		}
		if archived && id == 0 {
			writeError(c, 400, "restore this plant before creating a post")
			return
		}
	}
	var status any
	if p.Type == "sale" {
		status = "available"
	} else {
		p.PriceBDT = nil
		p.Quantity = nil
		p.ContactMethod = ""
	}
	args := []any{strings.TrimSpace(p.Title), strings.TrimSpace(p.Content), p.ImagePath, p.PlantID, p.PriceBDT, p.Quantity, strings.TrimSpace(p.ContactMethod)}
	if id == 0 {
		args = append(args, uid(c), p.Type, status)
		res, e := a.db.ExecContext(c, `INSERT INTO posts(title,content,image_path,plant_id,price_bdt,quantity,contact_method,user_id,type,sale_status) VALUES(?,?,?,?,?,?,?,?,?,?)`, args...)
		if e != nil {
			dbError(c, e)
			return
		}
		id, _ = res.LastInsertId()
		c.JSON(201, gin.H{"id": id})
		return
	}
	args = append(args, id, uid(c))
	_, e := editTx.ExecContext(c, `UPDATE posts SET title=?,content=?,image_path=?,plant_id=?,price_bdt=?,quantity=?,contact_method=? WHERE id=? AND user_id=?`, args...)
	if e != nil {
		dbError(c, e)
		return
	}
	if e = editTx.Commit(); e != nil {
		dbError(c, e)
		return
	}
	c.JSON(200, gin.H{"id": id})
}
func pageLimit(c *gin.Context) (int, int64, bool) {
	before := int64(0)
	var e error
	if c.Query("before") != "" {
		before, e = strconv.ParseInt(c.Query("before"), 10, 64)
		if e != nil || before < 1 {
			writeError(c, 400, "invalid cursor")
			return 0, 0, false
		}
	}
	return 20, before, true
}
func (a *app) posts(c *gin.Context) {
	limit, before, ok := pageLimit(c)
	if !ok {
		return
	}
	where := " WHERE p.archived_at IS NULL"
	args := []any{uid(c)}
	if before > 0 {
		where += " AND p.id<?"
		args = append(args, before)
	}
	typ := c.Query("type")
	if typ != "" {
		switch typ {
		case "update", "question", "blog", "sale":
		default:
			writeError(c, 400, "invalid post filter")
			return
		}
		where += " AND p.type=?"
		args = append(args, typ)
	}
	if c.Query("status") != "" {
		if c.Query("status") != "available" && c.Query("status") != "sold" && c.Query("status") != "reserved" {
			writeError(c, 400, "invalid sale status")
			return
		}
		where += " AND p.sale_status=?"
		args = append(args, c.Query("status"))
	}
	if c.Query("area") != "" {
		where += " AND u.district=?"
		args = append(args, c.Query("area"))
	}
	if c.Query("q") != "" {
		where += " AND (p.title LIKE ? OR p.content LIKE ?)"
		q := "%" + c.Query("q") + "%"
		args = append(args, q, q)
	}
	if c.Query("author") != "" {
		author, e := strconv.ParseInt(c.Query("author"), 10, 64)
		if e != nil || author < 1 {
			writeError(c, 400, "invalid author")
			return
		}
		where += " AND p.user_id=?"
		args = append(args, author)
	}
	args = append(args, limit+1)
	rows, e := a.db.QueryContext(c, `SELECT p.id,p.user_id,u.name,COALESCE(u.profile_photo,''),COALESCE(u.district,''),p.type,COALESCE(p.title,''),p.content,COALESCE(p.image_path,''),p.plant_id,p.price_bdt,p.quantity,COALESCE(p.contact_method,''),COALESCE(p.sale_status,''),DATE_FORMAT(p.created_at,'%Y-%m-%dT%H:%i:%sZ'),(SELECT COUNT(*) FROM post_likes l WHERE l.post_id=p.id),COALESCE((SELECT reaction FROM post_likes l WHERE l.post_id=p.id AND l.user_id=?),''),(SELECT COUNT(*) FROM comments cm WHERE cm.post_id=p.id),COALESCE(cr.name,''),(SELECT COUNT(*) FROM post_likes WHERE post_id=p.id AND reaction='like'),(SELECT COUNT(*) FROM post_likes WHERE post_id=p.id AND reaction='love'),(SELECT COUNT(*) FROM post_likes WHERE post_id=p.id AND reaction='angry'),(SELECT COUNT(*) FROM post_likes WHERE post_id=p.id AND reaction='sad'),(SELECT COUNT(*) FROM orders WHERE post_id=p.id) FROM posts p JOIN users u ON u.id=p.user_id LEFT JOIN user_plants pl ON pl.id=p.plant_id LEFT JOIN crops cr ON cr.id=pl.crop_id`+where+` ORDER BY p.id DESC LIMIT ?`, args...)
	if e != nil {
		dbError(c, e)
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id, author, likes, comments int64
		var name, photo, area, typ, title, content, im, contact, status, created, crop string
		var pid sql.NullInt64
		var price sql.NullFloat64
		var qty sql.NullInt64
		var reaction string
		var likeN, loveN, angryN, sadN, orderCount int
		if e = rows.Scan(&id, &author, &name, &photo, &area, &typ, &title, &content, &im, &pid, &price, &qty, &contact, &status, &created, &likes, &reaction, &comments, &crop, &likeN, &loveN, &angryN, &sadN, &orderCount); e != nil {
			dbError(c, e)
			return
		}
		var plantID, pr, quantity any
		if pid.Valid {
			plantID = pid.Int64
		}
		if price.Valid {
			pr = price.Float64
		}
		if qty.Valid {
			quantity = qty.Int64
		}
		out = append(out, gin.H{"id": id, "authorId": author, "authorName": name, "authorPhoto": photo, "area": area, "type": typ, "title": title, "content": content, "imagePath": im, "plantId": plantID, "cropName": crop, "priceBdt": pr, "quantity": quantity, "contactMethod": contact, "saleStatus": status, "createdAt": created, "likeCount": likes, "hasOrders": orderCount > 0, "liked": reaction != "", "reaction": reaction, "reactions": gin.H{"like": likeN, "love": loveN, "angry": angryN, "sad": sadN}, "commentCount": comments, "isOwner": author == uid(c)})
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
	c.JSON(200, gin.H{"posts": out, "nextCursor": next})
}
func (a *app) deletePost(c *gin.Context) {
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
	var n int64
	if e = tx.QueryRowContext(c, "SELECT id FROM posts WHERE id=? AND user_id=? FOR UPDATE", id, uid(c)).Scan(&n); e != nil {
		dbError(c, e)
		return
	}
	var count int
	if e = tx.QueryRowContext(c, "SELECT COUNT(*) FROM orders WHERE post_id=?", id).Scan(&count); e != nil {
		dbError(c, e)
		return
	}
	if count > 0 {
		writeError(c, 409, "order history must be preserved; sold listings can be archived by an admin")
		return
	}
	if _, e = tx.ExecContext(c, "DELETE FROM posts WHERE id=?", id); e != nil {
		dbError(c, e)
		return
	}
	if e = tx.Commit(); e != nil {
		dbError(c, e)
		return
	}
	c.Status(204)
}
func (a *app) saleStatus(c *gin.Context) {
	writeError(c, 409, "use order payment confirmation to mark a listing sold")
}
func (a *app) postExists(c *gin.Context, id int64) bool {
	var n int64
	e := a.db.QueryRowContext(c, "SELECT id FROM posts WHERE id=? AND archived_at IS NULL", id).Scan(&n)
	if e != nil {
		dbError(c, e)
		return false
	}
	return true
}
func (a *app) likePost(c *gin.Context) {
	id, ok := pathID(c)
	if !ok || !a.postExists(c, id) {
		return
	}
	var b struct {
		Liked *bool `json:"liked"`
	}
	if !decodeJSON(c, &b) {
		return
	}
	if b.Liked == nil {
		writeError(c, 400, "liked is required")
		return
	}
	var e error
	if *b.Liked {
		_, e = a.db.ExecContext(c, "INSERT INTO post_likes(post_id,user_id) VALUES(?,?) ON DUPLICATE KEY UPDATE post_id=post_id", id, uid(c))
	} else {
		_, e = a.db.ExecContext(c, "DELETE FROM post_likes WHERE post_id=? AND user_id=?", id, uid(c))
	}
	if e != nil {
		dbError(c, e)
		return
	}
	c.Status(204)
}
func (a *app) comments(c *gin.Context) {
	id, ok := pathID(c)
	if !ok || !a.postExists(c, id) {
		return
	}
	limit, before, ok := pageLimit(c)
	if !ok {
		return
	}
	q := `SELECT c.id,c.user_id,u.name,COALESCE(u.profile_photo,''),c.content,DATE_FORMAT(c.created_at,'%Y-%m-%dT%H:%i:%sZ') FROM comments c JOIN users u ON u.id=c.user_id WHERE c.post_id=?`
	args := []any{id}
	if before > 0 {
		q += " AND c.id<?"
		args = append(args, before)
	}
	args = append(args, limit+1)
	rows, e := a.db.QueryContext(c, q+" ORDER BY c.id DESC LIMIT ?", args...)
	if e != nil {
		dbError(c, e)
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var cid, user int64
		var n, photo, txt, date string
		if e = rows.Scan(&cid, &user, &n, &photo, &txt, &date); e != nil {
			dbError(c, e)
			return
		}
		out = append(out, gin.H{"id": cid, "authorId": user, "authorName": n, "authorPhoto": photo, "content": txt, "createdAt": date, "isOwner": user == uid(c)})
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
	c.JSON(200, gin.H{"comments": out, "nextCursor": next})
}
func (a *app) addComment(c *gin.Context) {
	id, ok := pathID(c)
	if !ok || !a.postExists(c, id) {
		return
	}
	var b struct {
		Content string `json:"content"`
	}
	if !decodeJSON(c, &b) {
		return
	}
	if !bounded(b.Content, 1, 1000) {
		writeError(c, 400, "comment needs 1–1000 characters")
		return
	}
	res, e := a.db.ExecContext(c, "INSERT INTO comments(post_id,user_id,content) VALUES(?,?,?)", id, uid(c), strings.TrimSpace(b.Content))
	if e != nil {
		dbError(c, e)
		return
	}
	n, _ := res.LastInsertId()
	c.JSON(201, gin.H{"id": n})
}
func (a *app) deleteComment(c *gin.Context) {
	id, ok := pathID(c)
	if !ok || !a.owns(c, "comments", id) {
		return
	}
	_, e := a.db.ExecContext(c, "DELETE FROM comments WHERE id=? AND user_id=?", id, uid(c))
	if e != nil {
		dbError(c, e)
		return
	}
	c.Status(204)
}
func (a *app) publicProfile(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	var name string
	var area, bio, photo sql.NullString
	var count int
	e := a.db.QueryRowContext(c, `SELECT name,district,bio,profile_photo,
 (SELECT COUNT(*) FROM posts WHERE user_id=users.id AND archived_at IS NULL)
 FROM users WHERE id=?`, id).Scan(&name, &area, &bio, &photo, &count)
	if e != nil {
		dbError(c, e)
		return
	}
	c.JSON(200, gin.H{"person": gin.H{"id": id, "name": name, "area": area.String, "bio": bio.String, "profilePhoto": photo.String, "postCount": count, "isSelf": id == uid(c)}})
}
