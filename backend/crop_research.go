package main

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"math"
	"time"
)

//go:embed crop_catalog.json
var cropCatalogJSON []byte

type researchCrop struct {
	ID                 int64               `json:"id"`
	Slug               string              `json:"slug"`
	LegacyName         string              `json:"legacyName"`
	Name               string              `json:"name"`
	BengaliName        string              `json:"bengaliName"`
	ScientificName     string              `json:"scientificName"`
	VarietyNote        string              `json:"varietyNote"`
	SeedMonths         []int               `json:"seedMonths"`
	TransplantMonths   []int               `json:"transplantMonths"`
	MinSunHours        *float64            `json:"minSunHours"`
	PreferredSunHours  []float64           `json:"preferredSunHours"`
	MinDepthCm         *float64            `json:"minDepthCm"`
	MinVolumeL         *float64            `json:"minVolumeL"`
	ContainerSupported *bool               `json:"containerSupported"`
	SupportRequired    *bool               `json:"supportRequired"`
	CoreVerified       bool                `json:"coreVerified"`
	SeasonNote         string              `json:"seasonNote"`
	ContainerNote      string              `json:"containerNote"`
	CareNote           string              `json:"careNote"`
	RegionNote         string              `json:"regionNote"`
	Districts          []string            `json:"districts"`
	SourceIDs          []string            `json:"sourceIds"`
	FieldSources       map[string][]string `json:"fieldSources"`
	CheckedAt          string              `json:"checkedAt"`
	ReportCheckedAt    string              `json:"reportCheckedAt"`
	ReportHarvest      json.RawMessage     `json:"reportHarvest"`
	TemperatureRange   json.RawMessage     `json:"temperatureRange"`
	ReviewNote         string              `json:"reviewNote"`
	ImagePath          *string             `json:"imagePath"`
	Status             string              `json:"status"`
	Reasons            []string            `json:"reasons"`
	Conditions         []string            `json:"conditions"`
}
type researchDocument struct {
	Version    int                        `json:"version"`
	Provenance string                     `json:"provenance"`
	Sources    map[string]json.RawMessage `json:"sources"`
	Crops      []researchCrop             `json:"crops"`
}
type gardenSetup struct {
	Depth    *float64 `json:"depthCm"`
	Volume   *float64 `json:"volumeL"`
	Drainage *bool    `json:"drainage"`
	Support  *bool    `json:"support"`
	Method   string   `json:"method"`
}

const cropEvidenceDDL = `CREATE TABLE IF NOT EXISTS crop_evidence (
 slug VARCHAR(60) PRIMARY KEY, crop_id BIGINT UNSIGNED NOT NULL UNIQUE,
 data JSON NOT NULL, FOREIGN KEY(crop_id) REFERENCES crops(id)
) ENGINE=InnoDB`
const gardenSetupDDL = `CREATE TABLE IF NOT EXISTS garden_setup (
 user_id BIGINT UNSIGNED PRIMARY KEY,
 depth_cm DECIMAL(7,2) NULL, volume_l DECIMAL(8,3) NULL,
 drainage BOOLEAN NULL, support BOOLEAN NULL,
 method ENUM('seed','seedling') NOT NULL DEFAULT 'seed',
 FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB`

func upgradeCrops(db *sql.DB) error {
	var doc researchDocument
	if err := json.Unmarshal(cropCatalogJSON, &doc); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	for _, ddl := range []string{cropEvidenceDDL, gardenSetupDDL} {
		if _, err := db.ExecContext(ctx, ddl); err != nil {
			return err
		}
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, crop := range doc.Crops {
		// Existing rows/IDs and user plant references are preserved, including user edits.
		_, err = tx.ExecContext(ctx, `INSERT INTO crops(name,bengali_name,description,planting_months,min_sunlight_hours,max_sunlight_hours,suitable_spaces,container_guidance,care_guidance,source_reference,verification_status) VALUES(?,?,?,'[]',0,24,'["rooftop","balcony"]',?,?,?,'source-pending') ON DUPLICATE KEY UPDATE id=id`, crop.LegacyName, crop.BengaliName, crop.VarietyNote, crop.ContainerNote, crop.CareNote, "Research catalog: see crop_evidence and field source links")
		if err != nil {
			return err
		}
		var id int64
		if err = tx.QueryRowContext(ctx, "SELECT id FROM crops WHERE name=?", crop.LegacyName).Scan(&id); err != nil {
			return err
		}
		b, err := json.Marshal(crop)
		if err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO crop_evidence(slug,crop_id,data) VALUES(?,?,?) ON DUPLICATE KEY UPDATE data=VALUES(data)`, crop.Slug, id, string(b)); err != nil {
			return err
		}
	}
	return tx.Commit()
}
func validGardenSetup(s gardenSetup) bool {
	if s.Method != "seed" && s.Method != "seedling" {
		return false
	}
	for i, v := range []*float64{s.Depth, s.Volume} {
		minimum := []float64{0.01, 0.001}[i]
		if v != nil && (math.IsNaN(*v) || math.IsInf(*v, 0) || *v < minimum || *v > 1000) {
			return false
		}
	}
	return true
}
func (a *app) loadGardenSetup(ctx context.Context, user int64) (gardenSetup, error) {
	s := gardenSetup{Method: "seed"}
	err := a.db.QueryRowContext(ctx, "SELECT depth_cm,volume_l,drainage,support,method FROM garden_setup WHERE user_id=?", user).Scan(&s.Depth, &s.Volume, &s.Drainage, &s.Support, &s.Method)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return s, err
}
func (a *app) getGardenSetup(c *gin.Context) {
	s, err := a.loadGardenSetup(c, uid(c))
	if err != nil {
		dbError(c, err)
		return
	}
	c.JSON(200, s)
}
func (a *app) saveGardenSetup(c *gin.Context) {
	var s gardenSetup
	if !decodeJSON(c, &s) {
		return
	}
	if !validGardenSetup(s) {
		writeError(c, 400, "Use seed or seedling; depth must be blank or 0.01–1000 cm, volume blank or 0.001–1000 litres.")
		return
	}
	// Match the database precision so a subsequent GET returns the saved values.
	if s.Depth != nil {
		v := math.Round(*s.Depth*100) / 100
		s.Depth = &v
	}
	if s.Volume != nil {
		v := math.Round(*s.Volume*1000) / 1000
		s.Volume = &v
	}
	_, err := a.db.ExecContext(c, `INSERT INTO garden_setup(user_id,depth_cm,volume_l,drainage,support,method) VALUES(?,?,?,?,?,?) ON DUPLICATE KEY UPDATE depth_cm=VALUES(depth_cm),volume_l=VALUES(volume_l),drainage=VALUES(drainage),support=VALUES(support),method=VALUES(method)`, uid(c), s.Depth, s.Volume, s.Drainage, s.Support, s.Method)
	if err != nil {
		dbError(c, err)
		return
	}
	c.JSON(200, s)
}
func assessCrop(c researchCrop, month int, space string, sun float64, s gardenSetup) researchCrop {
	c.Reasons = []string{}
	c.Conditions = []string{c.RegionNote, "Month-level guidance: check the exact planting window and seed packet before planting.", "Forecast is separate from seasonal suitability; no sensors measure this pot."}
	c.Status = "candidate"
	if !c.CoreVerified || c.ContainerSupported == nil || !*c.ContainerSupported || c.MinSunHours == nil {
		c.Status = "needs_evidence"
		c.Reasons = append(c.Reasons, c.ReviewNote)
		return c
	}
	months := c.SeedMonths
	if s.Method == "seedling" {
		months = c.TransplantMonths
	}
	if len(months) == 0 {
		c.Status = "needs_evidence"
		c.Reasons = append(c.Reasons, "No verified planting window for the selected seed/transplant method.")
		return c
	}
	inSeason := false
	for _, m := range months {
		if month == m {
			inSeason = true
		}
	}
	if !inSeason {
		c.Status = "out_of_season"
		c.Reasons = append(c.Reasons, "Current Dhaka calendar month is outside this method's documented window.")
	}
	if (space != "rooftop" && space != "balcony") || sun < *c.MinSunHours {
		c.Status = "not_fit"
		c.Reasons = append(c.Reasons, "Direct sunlight or growing space does not meet this guidance.")
	}
	// Preferred sunlight maximum is NOT a hard upper suitability bound.
	for _, v := range []struct {
		got, want *float64
		label     string
	}{{s.Depth, c.MinDepthCm, "depth (cm)"}, {s.Volume, c.MinVolumeL, "volume (litres)"}} {
		if v.want == nil {
			c.Conditions = append(c.Conditions, "Crop-specific container "+v.label+" is unknown.")
			if c.Status == "candidate" {
				c.Status = "conditional"
			}
		} else if v.got == nil {
			c.Conditions = append(c.Conditions, "Enter your pot "+v.label+" to check it.")
			if c.Status == "candidate" || c.Status == "conditional" {
				c.Status = "needs_setup"
			}
		} else if *v.got+0.0001 < *v.want {
			c.Status = "not_fit"
			c.Reasons = append(c.Reasons, fmt.Sprintf("Pot %s is below the selected guidance (%.3g).", v.label, *v.want))
		}
	}
	if s.Drainage == nil {
		c.Conditions = append(c.Conditions, "Confirm that the container drains.")
		if c.Status == "candidate" || c.Status == "conditional" {
			c.Status = "needs_setup"
		}
	} else if !*s.Drainage {
		c.Status = "not_fit"
		c.Reasons = append(c.Reasons, "Provide drainage before planting.")
	}
	if c.SupportRequired != nil && *c.SupportRequired && (s.Support == nil || !*s.Support) {
		c.Conditions = append(c.Conditions, "Provide appropriate stakes/trellis for the selected variety.")
		if c.Status == "candidate" {
			c.Status = "conditional"
		}
	}
	if len(c.Reasons) == 0 {
		c.Reasons = append(c.Reasons, "Season and sunlight align; review container, cultivar and local conditions.")
	}
	return c
}
func (a *app) researchRecommendations(c *gin.Context) {
	user, err := a.loadProfile(c, uid(c))
	if err != nil {
		dbError(c, err)
		return
	}
	if user.District == nil || user.GardeningSpace == nil || user.SunlightHours == nil {
		writeError(c, 422, "complete your gardening profile first")
		return
	}
	loc, ok := locationForDistrict(*user.District)
	if !ok {
		writeError(c, 422, "unsupported weather location")
		return
	}
	setup, err := a.loadGardenSetup(c, uid(c))
	if err != nil {
		dbError(c, err)
		return
	}
	rows, err := a.db.QueryContext(c, "SELECT crop_id,data FROM crop_evidence ORDER BY slug")
	if err != nil {
		dbError(c, err)
		return
	}
	all := []researchCrop{}
	matched := []researchCrop{}
	now := time.Now().In(dhaka)
	for rows.Next() {
		var crop researchCrop
		var id int64
		var b []byte
		if err = rows.Scan(&id, &b); err != nil {
			break
		}
		if err = json.Unmarshal(b, &crop); err != nil {
			break
		}
		crop.ID = id
		crop = assessCrop(crop, int(now.Month()), *user.GardeningSpace, *user.SunlightHours, setup)
		all = append(all, crop)
		if crop.Status == "candidate" || crop.Status == "conditional" {
			matched = append(matched, crop)
		}
	}
	if err == nil {
		err = rows.Err()
	}
	rows.Close()
	if err != nil {
		dbError(c, err)
		return
	}
	var doc researchDocument
	if err = json.Unmarshal(cropCatalogJSON, &doc); err != nil {
		dbError(c, err)
		return
	}
	weather := weatherResponse{Status: "unavailable", Message: "Weather provider is not configured."}
	if a.weather != nil {
		weather = a.weather.get(c, loc)
	}
	c.JSON(200, gin.H{"location": *user.District, "month": int(now.Month()), "date": now.Format("2006-01-02"), "weather": weather, "crops": matched, "cropCatalog": all, "setup": setup, "sources": doc.Sources, "guidance": "General Bangladesh seasonal shortlist, not district-specific certification. Unknown values are not treated as suitable."})
}
