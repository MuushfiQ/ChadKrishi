package main

import (
	"database/sql"
	"encoding/json"
	"strings"
	"testing"
)

func TestValidateRegistration(t *testing.T) {
	valid := registerRequest{Name: "Mushfiq", Email: "user@example.com", Password: "strongpass"}
	if err := validateRegistration(valid); err != nil {
		t.Fatalf("valid registration rejected: %v", err)
	}
	for name, request := range map[string]registerRequest{
		"short password": {Name: "User", Email: "user@example.com", Password: "short"},
		"invalid email":  {Name: "User", Email: "not-an-email", Password: "strongpass"},
		"short name":     {Name: "U", Email: "user@example.com", Password: "strongpass"},
	} {
		t.Run(name, func(t *testing.T) {
			if validateRegistration(request) == nil {
				t.Error("expected validation error")
			}
		})
	}
}

func TestValidateProfile(t *testing.T) {
	latitude, longitude, sunlight := 23.8103, 90.4125, 6.0
	rain := true
	request := profileRequest{
		Name: "Mushfiq", District: "Dhaka", Latitude: &latitude, Longitude: &longitude,
		GardeningSpace: "rooftop", SunlightHours: &sunlight, RainExposure: &rain,
	}
	if err := validateProfile(request); err != nil {
		t.Fatalf("valid profile rejected: %v", err)
	}
	request.GardeningSpace = "garden"
	if validateProfile(request) == nil {
		t.Error("expected invalid space to be rejected")
	}
}

func TestHashTokenIsDeterministicAndNotPlaintext(t *testing.T) {
	token := "random-session-token"
	if hashToken(token) == token || hashToken(token) != hashToken(token) {
		t.Error("token hash should be deterministic and not plaintext")
	}
}

func TestIsSuitableUsesMonthSpaceAndSunlight(t *testing.T) {
	crop := cropRecommendation{
		PlantingMonths:   []int{10, 11},
		MinSunlightHours: 4,
		MaxSunlightHours: 8,
		SuitableSpaces:   []string{"rooftop"},
	}
	if !isSuitable(crop, 10, "rooftop", 6) {
		t.Error("expected matching crop profile to be suitable")
	}
	if isSuitable(crop, 7, "rooftop", 6) || isSuitable(crop, 10, "balcony", 6) || isSuitable(crop, 10, "rooftop", 2) {
		t.Error("expected month, space, and sunlight mismatches to be rejected")
	}
}

func TestIsProfileCompleteDistinguishesNullProfileFields(t *testing.T) {
	valid := sql.NullString{String: "Dhaka", Valid: true}
	space := sql.NullString{String: "rooftop", Valid: true}
	sunlight := sql.NullFloat64{Float64: 6, Valid: true}
	if !isProfileComplete(valid, space, sunlight) {
		t.Fatal("expected complete profile")
	}

	if isProfileComplete(sql.NullString{}, space, sunlight) {
		t.Error("expected missing district to be incomplete")
	}
	if isProfileComplete(valid, sql.NullString{}, sunlight) {
		t.Error("expected missing gardening space to be incomplete")
	}
	if isProfileComplete(valid, space, sql.NullFloat64{}) {
		t.Error("expected missing sunlight to be incomplete")
	}
}

func TestWeatherDescriptionUsesWMOCodeNames(t *testing.T) {
	for code, expected := range map[int]string{
		0: "Clear sky", 2: "Partly cloudy", 61: "Rain", 95: "Thunderstorm",
	} {
		if got := weatherDescription(code); got != expected {
			t.Errorf("weatherDescription(%d) = %q, want %q", code, got, expected)
		}
	}
}

func TestCropJSONKeepsBengaliTextUTF8(t *testing.T) {
	payload, err := json.Marshal(cropRecommendation{BengaliName: "বেগুন"})
	if err != nil {
		t.Fatalf("marshal crop: %v", err)
	}
	if !strings.Contains(string(payload), `"bengaliName":"বেগুন"`) {
		t.Fatalf("Bengali JSON was not preserved: %s", payload)
	}
}
