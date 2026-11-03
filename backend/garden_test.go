package main

import (
	"testing"
	"time"
)

func TestPlantValidationAndDhakaDate(t *testing.T) {
	now := time.Date(2026, 10, 21, 19, 0, 0, 0, time.UTC)
	p := plantInput{CropID: 1, Nickname: "বেগুন", PlantingDate: "2026-10-22", PropagationMethod: "seedling", GrowthStage: "growing"}
	if err := validatePlant(p, now); err != nil {
		t.Fatal(err)
	}
	p.PlantingDate = "2026-10-23"
	if validatePlant(p, now) == nil {
		t.Fatal("future planting accepted")
	}
	p.PlantingDate = "2026-10-22"
	p.GrowthStage = "automatically_healthy"
	if validatePlant(p, now) == nil {
		t.Fatal("invalid growth stage accepted")
	}
}
func TestSaleValidation(t *testing.T) {
	price := 250.0
	qty := 1
	p := postInput{Type: "sale", Title: "My plant", Content: "A home-grown plant", PriceBDT: &price, Quantity: &qty}
	if e := validatePost(p); e != nil {
		t.Fatal(e)
	}
	qty = 0
	if validatePost(p) == nil {
		t.Fatal("zero quantity accepted")
	}
	qty = 1
	price = -1
	if validatePost(p) == nil {
		t.Fatal("negative price accepted")
	}
	price = 0
	if e := validatePost(p); e != nil {
		t.Fatal("free listing rejected", e)
	}
}
func TestRoutineRulesAreStageAware(t *testing.T) {
	seed := routineRules("seedling")
	growing := routineRules("growing")
	flower := routineRules("flowering")
	if len(seed) != 3 || len(growing) != 2 || len(flower) != 3 {
		t.Fatal("unexpected stage prompts")
	}
	seen := map[string]bool{}
	for _, r := range seed {
		if seen[r.key] {
			t.Fatal("duplicate task identity")
		}
		seen[r.key] = true
	}
}
