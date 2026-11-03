package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

const sessionDuration = 7 * 24 * time.Hour

type app struct {
	db             *sql.DB
	cookieSecure   bool
	frontendOrigin string
	staticDir      string
	weather        *weatherService
}

type supportedLocation struct {
	Latitude  float64
	Longitude float64
}

var supportedLocations = map[string]supportedLocation{
	"Dhaka":      {23.8103, 90.4125},
	"Chattogram": {22.3569, 91.7832},
	"Rajshahi":   {24.3745, 88.6042},
	"Khulna":     {22.8456, 89.5403},
	"Sylhet":     {24.8949, 91.8687},
	"Rangpur":    {25.7439, 89.2752},
}

type weatherCurrent struct {
	Time             string  `json:"time"`
	Temperature      float64 `json:"temperature"`
	RelativeHumidity float64 `json:"relativeHumidity"`
	WeatherCode      int     `json:"weatherCode"`
	Description      string  `json:"description"`
	WindSpeed        float64 `json:"windSpeed"`
}

type weatherDay struct {
	Date              string   `json:"date"`
	MinTemperature    float64  `json:"minTemperature"`
	MaxTemperature    float64  `json:"maxTemperature"`
	PrecipitationSum  float64  `json:"precipitationSum"`
	PrecipitationProb *float64 `json:"precipitationProbability,omitempty"`
}

type weatherResponse struct {
	Status      string          `json:"status"`
	Provider    string          `json:"provider,omitempty"`
	Attribution string          `json:"attribution,omitempty"`
	ValidTime   string          `json:"validTime,omitempty"`
	FetchedAt   string          `json:"fetchedAt,omitempty"`
	Stale       bool            `json:"stale,omitempty"`
	Message     string          `json:"message,omitempty"`
	Current     *weatherCurrent `json:"current,omitempty"`
	Forecast    []weatherDay    `json:"forecast,omitempty"`
}

type weatherCacheEntry struct {
	weather   weatherResponse
	fetchedAt time.Time
}

type weatherService struct {
	client  *http.Client
	baseURL string
	ttl     time.Duration
	mu      sync.Mutex
	cache   map[string]weatherCacheEntry
}

type profile struct {
	ID             int64    `json:"id"`
	IsAdmin        bool     `json:"isAdmin"`
	Name           string   `json:"name"`
	Email          string   `json:"email"`
	ProfilePhoto   *string  `json:"profilePhoto,omitempty"`
	District       *string  `json:"district,omitempty"`
	Latitude       *float64 `json:"latitude,omitempty"`
	Longitude      *float64 `json:"longitude,omitempty"`
	GardeningSpace *string  `json:"gardeningSpace,omitempty"`
	SunlightHours  *float64 `json:"sunlightHours,omitempty"`
	RainExposure   *bool    `json:"rainExposure,omitempty"`
	Bio            *string  `json:"bio,omitempty"`
}

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type profileRequest struct {
	Name           string   `json:"name"`
	District       string   `json:"district"`
	Latitude       *float64 `json:"latitude"`
	Longitude      *float64 `json:"longitude"`
	GardeningSpace string   `json:"gardeningSpace"`
	SunlightHours  *float64 `json:"sunlightHours"`
	RainExposure   *bool    `json:"rainExposure"`
	Bio            string   `json:"bio"`
}

type cropRecommendation struct {
	ID                     int64    `json:"id"`
	Name                   string   `json:"name"`
	BengaliName            string   `json:"bengaliName"`
	VarietyCaveat          string   `json:"varietyCaveat,omitempty"`
	Description            string   `json:"description"`
	ImagePath              *string  `json:"imagePath,omitempty"`
	PlantingMonths         []int    `json:"plantingMonths"`
	MinSunlightHours       float64  `json:"minSunlightHours"`
	MaxSunlightHours       float64  `json:"maxSunlightHours"`
	SuitableSpaces         []string `json:"suitableSpaces"`
	ContainerGuidance      string   `json:"containerGuidance"`
	CareGuidance           string   `json:"careGuidance"`
	SourceReference        string   `json:"sourceReference"`
	SourceURL              string   `json:"sourceUrl"`
	SourceNotes            string   `json:"sourceNotes"`
	ReviewedDate           string   `json:"reviewedDate"`
	RegionNote             string   `json:"regionNote"`
	ContainerSuitability   string   `json:"containerSuitability"`
	ContainerDepthInches   *float64 `json:"containerDepthInches,omitempty"`
	ContainerVolumeGallons *float64 `json:"containerVolumeGallons,omitempty"`
	TemperatureGuidance    string   `json:"temperatureGuidance,omitempty"`
	VerificationStatus     string   `json:"verificationStatus"`
	Reason                 string   `json:"reason"`
	Unknowns               []string `json:"unknowns,omitempty"`
	Conditions             []string `json:"conditions,omitempty"`
	Sources                []string `json:"sources,omitempty"`
}

func main() {
	loadEnvFile(".env")
	loadEnvFile("../.env")
	repair := flag.Bool("repair-notifications", false, "Create missing notifications table; preserve existing data and exit")
	check := flag.Bool("check-db", false, "Check database connection and commerce schema, then exit")
	cropUpgrade := flag.Bool("upgrade-crops", false, "Install/update reviewed crop catalog and garden setup tables, preserving plant IDs")
	flag.Parse()
	dsn := os.Getenv("DATABASE_DSN")
	var db *sql.DB
	if dsn != "" {
		var err error
		db, err = openDatabase(dsn)
		if err != nil {
			panic(err)
		}
		defer db.Close()
	}

	if db == nil {
		log.Fatal("DATABASE_DSN is required; copy .env.example to .env and configure it")
	}
	if *cropUpgrade {
		if err := upgradeCrops(db); err != nil {
			log.Fatal(err)
		}
		log.Print("Crop catalog upgraded; existing plants preserved.")
		return
	}
	if *repair {
		if err := repairNotifications(db); err != nil {
			log.Fatal(err)
		}
		log.Print("Notification schema ready; existing data preserved.")
		return
	}
	if err := checkSchema(db); err != nil {
		log.Fatal(err)
	}
	if *check {
		log.Print("Database connection and commerce schema OK")
		return
	}
	cookieSecure, _ := strconv.ParseBool(os.Getenv("SESSION_COOKIE_SECURE"))
	if os.Getenv("APP_ENV") == "production" {
		origin, err := url.Parse(os.Getenv("FRONTEND_ORIGIN"))
		if err != nil || origin.Scheme != "https" || origin.Host == "" || !cookieSecure {
			log.Fatal("Production requires an HTTPS FRONTEND_ORIGIN and SESSION_COOKIE_SECURE=true")
		}
	}
	application := &app{
		db:             db,
		cookieSecure:   cookieSecure,
		frontendOrigin: envOr("FRONTEND_ORIGIN", "http://localhost:5173"),
		staticDir:      os.Getenv("STATIC_DIR"),
		weather:        newWeatherService(),
	}
	router := gin.New()
	router.SetTrustedProxies(nil)
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(application.originGuard)
	router.Use(func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		if c.Request.Method != "GET" {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 6*1024*1024)
		}
		c.Next()
	})
	application.gardenRoutes(router)
	application.socialRoutes(router)

	router.GET("/api/v1/health", func(c *gin.Context) {
		response := gin.H{"status": "ok", "service": "chadkrishi-api"}
		if application.db == nil {
			response["database"] = "not_configured"
		} else if err := checkSchema(application.db); err != nil {
			response["status"] = "degraded"
			response["database"] = "unavailable"
		} else {
			response["database"] = "ok"
		}
		if response["database"] != "ok" {
			c.JSON(http.StatusServiceUnavailable, response)
			return
		}
		c.JSON(http.StatusOK, response)
	})
	router.NoRoute(application.serveFrontend)
	auth := router.Group("/api/v1/auth")
	auth.POST("/register", application.register)
	auth.POST("/login", application.login)
	auth.POST("/logout", application.requireUser, application.logout)
	auth.GET("/me", application.requireUser, application.currentUser)
	router.PUT("/api/v1/profile", application.requireUser, application.updateProfile)
	router.GET("/api/v1/recommendations", application.requireUser, application.researchRecommendations)
	router.GET("/api/v1/weather", application.requireUser, application.weatherForecast)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{Addr: ":" + port, Handler: router, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}

}

func newWeatherService() *weatherService {
	ttlMinutes, err := strconv.Atoi(envOr("WEATHER_CACHE_MINUTES", "15"))
	if err != nil || ttlMinutes < 1 {
		ttlMinutes = 15
	}
	return &weatherService{
		client:  &http.Client{Timeout: 8 * time.Second},
		baseURL: strings.TrimRight(envOr("OPEN_METEO_BASE_URL", "https://api.open-meteo.com/v1/forecast"), "/"),
		ttl:     time.Duration(ttlMinutes) * time.Minute,
		cache:   make(map[string]weatherCacheEntry),
	}
}

func openDatabase(dsn string) (*sql.DB, error) {
	tlsEnabled := os.Getenv("DATABASE_TLS_CA_PEM") != "" || os.Getenv("DATABASE_TLS_CA_FILE") != ""
	if tlsEnabled {
		if err := configureDatabaseTLS(); err != nil {
			return nil, err
		}
	}
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return nil, err
	}
	cfg.ParseTime = true
	cfg.Loc = time.UTC
	if cfg.Params == nil {
		cfg.Params = map[string]string{}
	}
	cfg.Params["charset"] = "utf8mb4"
	cfg.Params["time_zone"] = "'+00:00'"
	if tlsEnabled {
		if cfg.TLSConfig == "" {
			cfg.TLSConfig = "aiven"
		}
	}
	return sql.Open("mysql", cfg.FormatDSN())
}

func configureDatabaseTLS() error {
	caPEM := os.Getenv("DATABASE_TLS_CA_PEM")
	if caPEM == "" {
		caPath := os.Getenv("DATABASE_TLS_CA_FILE")
		data, err := os.ReadFile(caPath)
		if err != nil {
			return fmt.Errorf("database TLS CA file could not be read: %w", err)
		}
		caPEM = string(data)
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM([]byte(caPEM)) {
		return fmt.Errorf("database TLS CA certificate is not valid PEM")
	}
	serverName := os.Getenv("DATABASE_TLS_SERVER_NAME")
	if serverName == "" {
		return fmt.Errorf("DATABASE_TLS_SERVER_NAME is required when database TLS is enabled")
	}
	if err := mysql.RegisterTLSConfig("aiven", &tls.Config{
		MinVersion: tls.VersionTLS12,
		RootCAs:    roots,
		ServerName: serverName,
	}); err != nil {
		return fmt.Errorf("database TLS configuration failed: %w", err)
	}
	return nil
}

func (a *app) serveFrontend(c *gin.Context) {
	if a.staticDir == "" || strings.HasPrefix(c.Request.URL.Path, "/api/") {
		c.Status(http.StatusNotFound)
		return
	}
	requestPath := strings.TrimPrefix(filepath.Clean(c.Request.URL.Path), string(filepath.Separator))
	path := filepath.Join(a.staticDir, requestPath)
	if info, err := os.Stat(path); err != nil || info.IsDir() {
		path = filepath.Join(a.staticDir, "index.html")
	}
	http.ServeFile(c.Writer, c.Request, path)
}

func locationForDistrict(district string) (supportedLocation, bool) {
	location, ok := supportedLocations[district]
	return location, ok
}

func (s *weatherService) get(ctx context.Context, location supportedLocation) weatherResponse {
	key := fmt.Sprintf("%.4f,%.4f", location.Latitude, location.Longitude)
	now := time.Now().UTC()
	s.mu.Lock()
	cached, hasCached := s.cache[key]
	if hasCached && now.Sub(cached.fetchedAt) < s.ttl {
		result := cached.weather
		s.mu.Unlock()
		return result
	}
	s.mu.Unlock()

	result, err := s.fetch(ctx, location)
	if err == nil {
		s.mu.Lock()
		s.cache[key] = weatherCacheEntry{weather: result, fetchedAt: now}
		s.mu.Unlock()
		return result
	}
	if hasCached {
		result := cached.weather
		result.Status = "stale"
		result.Stale = true
		result.Message = "Live weather is temporarily unavailable; showing the last fetched forecast."
		s.mu.Lock()
		s.cache[key] = cached
		s.mu.Unlock()
		return result
	}
	return weatherResponse{Status: "unavailable", Message: "Weather unavailable: " + err.Error()}
}

func (s *weatherService) fetch(ctx context.Context, location supportedLocation) (weatherResponse, error) {
	query := url.Values{}
	query.Set("latitude", strconv.FormatFloat(location.Latitude, 'f', 4, 64))
	query.Set("longitude", strconv.FormatFloat(location.Longitude, 'f', 4, 64))
	query.Set("current", "temperature_2m,relative_humidity_2m,weather_code,wind_speed_10m")
	query.Set("daily", "temperature_2m_min,temperature_2m_max,precipitation_probability_max,precipitation_sum")
	query.Set("forecast_days", "7")
	query.Set("timezone", "Asia/Dhaka")
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL+"?"+query.Encode(), nil)
	if err != nil {
		return weatherResponse{}, err
	}
	response, err := s.client.Do(request)
	if err != nil {
		return weatherResponse{}, fmt.Errorf("provider request failed")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return weatherResponse{}, fmt.Errorf("provider returned HTTP %d", response.StatusCode)
	}
	var payload struct {
		Current struct {
			Time             string   `json:"time"`
			Temperature      *float64 `json:"temperature_2m"`
			RelativeHumidity *float64 `json:"relative_humidity_2m"`
			WeatherCode      *int     `json:"weather_code"`
			WindSpeed        *float64 `json:"wind_speed_10m"`
		} `json:"current"`
		Daily struct {
			Time                     []string   `json:"time"`
			MinTemperature           []float64  `json:"temperature_2m_min"`
			MaxTemperature           []float64  `json:"temperature_2m_max"`
			PrecipitationProbability []*float64 `json:"precipitation_probability_max"`
			PrecipitationSum         []float64  `json:"precipitation_sum"`
		} `json:"daily"`
		Timezone string `json:"timezone"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return weatherResponse{}, fmt.Errorf("provider returned invalid JSON")
	}
	if payload.Current.Temperature == nil || payload.Current.RelativeHumidity == nil ||
		payload.Current.WeatherCode == nil || payload.Current.WindSpeed == nil ||
		len(payload.Daily.Time) != 7 || len(payload.Daily.MinTemperature) != 7 ||
		len(payload.Daily.MaxTemperature) != 7 || len(payload.Daily.PrecipitationSum) != 7 {
		return weatherResponse{}, fmt.Errorf("provider response is missing required fields")
	}
	result := weatherResponse{
		Status:      "available",
		Provider:    "Open-Meteo",
		Attribution: "Weather data by Open-Meteo.com",
		ValidTime:   payload.Current.Time,
		FetchedAt:   time.Now().UTC().Format(time.RFC3339),
		Current: &weatherCurrent{
			Time: payload.Current.Time, Temperature: *payload.Current.Temperature,
			RelativeHumidity: *payload.Current.RelativeHumidity, WeatherCode: *payload.Current.WeatherCode,
			Description: weatherDescription(*payload.Current.WeatherCode),
			WindSpeed:   *payload.Current.WindSpeed,
		},
		Forecast: make([]weatherDay, 7),
	}

	for i := range result.Forecast {
		var probability *float64
		if len(payload.Daily.PrecipitationProbability) == 7 {
			probability = payload.Daily.PrecipitationProbability[i]
		}
		result.Forecast[i] = weatherDay{
			Date: payload.Daily.Time[i], MinTemperature: payload.Daily.MinTemperature[i],
			MaxTemperature: payload.Daily.MaxTemperature[i], PrecipitationSum: payload.Daily.PrecipitationSum[i],
			PrecipitationProb: probability,
		}
	}
	return result, nil
}

func weatherDescription(code int) string {
	switch code {
	case 0:
		return "Clear sky"
	case 1:
		return "Mainly clear"
	case 2:
		return "Partly cloudy"
	case 3:
		return "Overcast"
	case 45, 48:
		return "Fog"
	case 51, 53, 55:
		return "Drizzle"
	case 56, 57:
		return "Freezing drizzle"
	case 61, 63, 65:
		return "Rain"
	case 66, 67:
		return "Freezing rain"
	case 71, 73, 75, 77:
		return "Snow"
	case 80, 81, 82:
		return "Rain showers"
	case 85, 86:
		return "Snow showers"
	case 95:
		return "Thunderstorm"
	case 96, 99:
		return "Thunderstorm with hail"
	default:
		return "Unknown weather conditions"
	}
}

func (a *app) originGuard(c *gin.Context) {
	origin := c.GetHeader("Origin")
	if origin != "" && origin != a.frontendOrigin {
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		writeError(c, http.StatusForbidden, "request origin is not allowed")
		c.Abort()
		return
	}
	if origin == a.frontendOrigin {
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		c.Header("Access-Control-Allow-Methods", "GET, HEAD, POST, PUT, DELETE, OPTIONS")
		c.Header("Vary", "Origin")
	}
	if c.Request.Method == http.MethodOptions {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}
	if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead {
		c.Next()
		return
	}
	c.Next()
}

func (a *app) requireDatabase(c *gin.Context) bool {
	if a.db == nil {
		writeError(c, http.StatusServiceUnavailable, "database is not configured")
		return false
	}
	return true
}

func (a *app) register(c *gin.Context) {
	if !a.requireDatabase(c) {
		return
	}
	var request registerRequest
	if !decodeJSON(c, &request) {
		return
	}
	request.Email = strings.ToLower(strings.TrimSpace(request.Email))
	request.Name = strings.TrimSpace(request.Name)
	if err := validateRegistration(request); err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "could not secure password")
		return
	}
	result, err := a.db.ExecContext(c, `INSERT INTO users (name, email, password_hash) VALUES (?, ?, ?)`,
		request.Name, request.Email, string(hash))
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			writeError(c, http.StatusConflict, "an account with this email already exists")
			return
		}
		writeError(c, http.StatusInternalServerError, "could not create account")
		return
	}
	userID, err := result.LastInsertId()
	if err != nil {
		writeError(c, http.StatusInternalServerError, "could not create account")
		return
	}
	if !a.startSession(c, userID) {
		return
	}
	c.JSON(http.StatusCreated, gin.H{"user": profile{ID: userID, Name: request.Name, Email: request.Email}})
}

func (a *app) login(c *gin.Context) {
	if !a.requireDatabase(c) {
		return
	}
	var request loginRequest
	if !decodeJSON(c, &request) {
		return
	}
	request.Email = strings.ToLower(strings.TrimSpace(request.Email))
	if request.Email == "" || request.Password == "" {
		writeError(c, http.StatusBadRequest, "email and password are required")
		return
	}
	var userID int64
	var name, email, passwordHash string
	err := a.db.QueryRowContext(c, `SELECT id, name, email, password_hash FROM users WHERE email = ?`, request.Email).
		Scan(&userID, &name, &email, &passwordHash)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(request.Password)) != nil {
		writeError(c, http.StatusUnauthorized, "email or password is incorrect")
		return
	}
	if !a.startSession(c, userID) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": profile{ID: userID, Name: name, Email: email}})
}

func (a *app) logout(c *gin.Context) {
	sessionToken, exists := c.Get("sessionToken")
	if exists && a.db != nil {
		if _, err := a.db.ExecContext(c, `UPDATE sessions SET revoked_at = UTC_TIMESTAMP() WHERE token_hash = ?`, hashToken(sessionToken.(string))); err != nil {
			a.clearSessionCookie(c)
			writeError(c, http.StatusInternalServerError, "could not sign out")
			return
		}
	}
	a.clearSessionCookie(c)
	c.Status(http.StatusNoContent)
}

func (a *app) currentUser(c *gin.Context) {
	if !a.requireDatabase(c) {
		return
	}
	userID := c.MustGet("userID").(int64)
	user, err := a.loadProfile(c, userID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "could not load profile")
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (a *app) updateProfile(c *gin.Context) {
	if !a.requireDatabase(c) {
		return
	}
	var request profileRequest
	if !decodeJSON(c, &request) {
		return
	}
	request.Name = strings.TrimSpace(request.Name)
	request.District = strings.TrimSpace(request.District)
	if err := validateProfile(request); err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	userID := c.MustGet("userID").(int64)
	_, err := a.db.ExecContext(c, `UPDATE users SET name = ?, district = ?, latitude = ?, longitude = ?,
		gardening_space = ?, sunlight_hours = ?, rain_exposure = ?, bio = ? WHERE id = ?`,
		request.Name, request.District, request.Latitude, request.Longitude, request.GardeningSpace,
		request.SunlightHours, request.RainExposure, strings.TrimSpace(request.Bio), userID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "could not update profile")
		return
	}
	user, err := a.loadProfile(c, userID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "could not load updated profile")
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (a *app) weatherForecast(c *gin.Context) {
	userID := c.MustGet("userID").(int64)
	var district sql.NullString
	if err := a.db.QueryRowContext(c, `SELECT district FROM users WHERE id = ?`, userID).Scan(&district); err != nil {
		writeError(c, http.StatusInternalServerError, "could not load gardening profile")
		return
	}
	location, ok := locationForDistrict(district.String)
	if !district.Valid || !ok {
		writeError(c, http.StatusUnprocessableEntity, "complete your gardening profile first")
		return
	}
	c.JSON(http.StatusOK, a.weather.get(c, location))
}

func isProfileComplete(district sql.NullString, gardeningSpace sql.NullString, sunlightHours sql.NullFloat64) bool {
	return district.Valid && district.String != "" &&
		gardeningSpace.Valid && gardeningSpace.String != "" &&
		sunlightHours.Valid
}

func isSuitable(crop cropRecommendation, month int, space string, sunlight float64) bool {
	if crop.VerificationStatus != "" && crop.VerificationStatus != "source-backed" {
		return false
	}
	monthMatches := false
	for _, plantingMonth := range crop.PlantingMonths {
		if plantingMonth == month {
			monthMatches = true
			break
		}
	}
	spaceMatches := false
	for _, suitableSpace := range crop.SuitableSpaces {
		if suitableSpace == space {
			spaceMatches = true
			break
		}
	}
	return monthMatches && spaceMatches && sunlight >= crop.MinSunlightHours && sunlight <= crop.MaxSunlightHours
}

func (a *app) requireUser(c *gin.Context) {
	if !a.requireDatabase(c) {
		c.Abort()
		return
	}
	cookie, err := c.Request.Cookie("chadkrishi_session")
	if err != nil || cookie.Value == "" {
		writeError(c, http.StatusUnauthorized, "login required")
		c.Abort()
		return
	}
	var userID int64
	err = a.db.QueryRowContext(c, `SELECT user_id FROM sessions
		WHERE token_hash = ? AND revoked_at IS NULL AND expires_at > UTC_TIMESTAMP()`,
		hashToken(cookie.Value)).Scan(&userID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			writeError(c, 503, "session service unavailable")
			c.Abort()
			return
		}
		a.clearSessionCookie(c)
		writeError(c, http.StatusUnauthorized, "login required")
		c.Abort()
		return
	}
	c.Set("userID", userID)
	c.Set("sessionToken", cookie.Value)
	c.Next()
}

func (a *app) startSession(c *gin.Context, userID int64) bool {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		writeError(c, http.StatusInternalServerError, "could not create session")
		return false
	}
	token := hex.EncodeToString(tokenBytes)
	_, err := a.db.ExecContext(c, `INSERT INTO sessions (user_id, token_hash, expires_at)
		VALUES (?, ?, UTC_TIMESTAMP() + INTERVAL 7 DAY)`, userID, hashToken(token))
	if err != nil {
		writeError(c, http.StatusInternalServerError, "could not create session")
		return false
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "chadkrishi_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   a.cookieSecure,
		SameSite: sessionCookieSameSite(a.cookieSecure),
		Expires:  time.Now().Add(sessionDuration),
	})
	return true
}

func (a *app) clearSessionCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{Name: "chadkrishi_session", Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: a.cookieSecure, SameSite: sessionCookieSameSite(a.cookieSecure)})
}

func sessionCookieSameSite(secure bool) http.SameSite {
	if secure {
		return http.SameSiteNoneMode
	}
	return http.SameSiteLaxMode
}

func (a *app) loadProfile(c *gin.Context, userID int64) (profile, error) {
	var user profile
	user.ID = userID
	err := a.db.QueryRowContext(c, `SELECT name, email, profile_photo, district, latitude, longitude,
		gardening_space, sunlight_hours, rain_exposure, bio, is_admin FROM users WHERE id = ?`, userID).Scan(
		&user.Name, &user.Email, &user.ProfilePhoto, &user.District, &user.Latitude, &user.Longitude,
		&user.GardeningSpace, &user.SunlightHours, &user.RainExposure, &user.Bio, &user.IsAdmin)
	return user, err
}

func decodeJSON(c *gin.Context, target any) bool {
	if err := c.ShouldBindJSON(target); err != nil {
		writeError(c, http.StatusBadRequest, "request body must be valid JSON")
		return false
	}
	return true
}

func validateRegistration(request registerRequest) error {
	if len(request.Name) < 2 || len(request.Name) > 100 {
		return errors.New("name must be between 2 and 100 characters")
	}
	if !strings.Contains(request.Email, "@") || len(request.Email) > 254 {
		return errors.New("a valid email is required")
	}
	if len(request.Password) < 8 || len(request.Password) > 72 {
		return errors.New("password must be between 8 and 72 characters")
	}
	return nil
}

func validateProfile(request profileRequest) error {
	if len(request.Name) < 2 || len(request.Name) > 100 {
		return errors.New("name must be between 2 and 100 characters")
	}
	if request.District == "" || len(request.District) > 80 {
		return errors.New("district is required")
	}
	if request.Latitude == nil || *request.Latitude < 20.5 || *request.Latitude > 26.7 ||
		request.Longitude == nil || *request.Longitude < 88.0 || *request.Longitude > 92.7 {
		return errors.New("select a valid Bangladesh location")
	}
	location, ok := locationForDistrict(request.District)
	if !ok || abs(*request.Latitude-location.Latitude) > 0.001 || abs(*request.Longitude-location.Longitude) > 0.001 {
		return errors.New("select a supported Bangladesh city")
	}
	if request.GardeningSpace != "rooftop" && request.GardeningSpace != "balcony" {
		return errors.New("gardening space must be rooftop or balcony")
	}

	if request.SunlightHours == nil || *request.SunlightHours < 0 || *request.SunlightHours > 24 {
		return errors.New("sunlight must be between 0 and 24 hours")
	}
	if request.RainExposure == nil {
		return errors.New("rain exposure selection is required")
	}
	if len(request.Bio) > 500 {
		return errors.New("bio cannot exceed 500 characters")
	}
	return nil
}

func abs(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func writeError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": gin.H{"message": message}})
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func loadEnvFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if found && os.Getenv(strings.TrimSpace(key)) == "" {
			_ = os.Setenv(strings.TrimSpace(key), strings.Trim(strings.TrimSpace(value), `"`))
		}
	}
}
