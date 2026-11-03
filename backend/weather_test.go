package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestWeatherProviderParsesForecastAndPreservesMissingRainProbability(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"current":{"time":"2025-10-01T12:00","temperature_2m":29.4,"relative_humidity_2m":70,"weather_code":3,"wind_speed_10m":12},"daily":{"time":["1","2","3","4","5","6","7"],"temperature_2m_min":[20,20,20,20,20,20,20],"temperature_2m_max":[30,30,30,30,30,30,30],"precipitation_probability_max":[null,20,30,40,50,60,70],"precipitation_sum":[0,1,2,3,4,5,6]},"timezone":"Asia/Dhaka"}`))
	}))
	defer server.Close()
	service := &weatherService{client: server.Client(), baseURL: server.URL, ttl: time.Minute, cache: make(map[string]weatherCacheEntry)}
	result, err := service.fetch(context.Background(), supportedLocation{23.8103, 90.4125})
	if err != nil {
		t.Fatalf("expected valid provider response: %v", err)
	}
	if len(result.Forecast) != 7 || result.Forecast[0].PrecipitationProb != nil {
		t.Fatal("missing precipitation probability must remain unknown, not become zero")
	}
}

func TestWeatherCacheUsesFreshValueAndStaleValueOnFailure(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if requests > 1 {
			http.Error(w, "offline", http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		forecast := `["2025-10-01","2025-10-02","2025-10-03","2025-10-04","2025-10-05","2025-10-06","2025-10-07"]`
		values := `[20,20,20,20,20,20,20]`
		w.Write([]byte(`{"current":{"time":"2025-10-01T12:00","temperature_2m":29,"relative_humidity_2m":70,"weather_code":1,"wind_speed_10m":8},"daily":{"time":` + forecast + `,"temperature_2m_min":` + values + `,"temperature_2m_max":` + values + `,"precipitation_sum":` + values + `},"timezone":"Asia/Dhaka"}`))
	}))
	defer server.Close()
	service := &weatherService{client: server.Client(), baseURL: server.URL, ttl: time.Nanosecond, cache: make(map[string]weatherCacheEntry)}
	location := supportedLocation{23.8103, 90.4125}
	first := service.get(context.Background(), location)
	if first.Status != "available" {
		t.Fatalf("expected first request to be available: %+v", first)
	}
	time.Sleep(2 * time.Millisecond)
	second := service.get(context.Background(), location)
	if second.Status != "stale" || !second.Stale || !strings.Contains(second.Message, "last fetched") {
		t.Fatalf("expected explicit stale cache response: %+v", second)
	}
}
