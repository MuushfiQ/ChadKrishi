package main

import (
	"encoding/json"
	"testing"
)

func fp(v float64) *float64 { return &v }
func bp(v bool) *bool       { return &v }
func testResearchCrop() researchCrop {
	return researchCrop{CoreVerified: true, ContainerSupported: bp(true), MinSunHours: fp(6), MinDepthCm: fp(30), MinVolumeL: fp(8), SeedMonths: []int{10}, TransplantMonths: []int{11}}
}
func TestResearchAssessment(t *testing.T) {
	s := gardenSetup{Depth: fp(35), Volume: fp(10), Drainage: bp(true), Support: bp(true), Method: "seed"}
	c := testResearchCrop()
	if got := assessCrop(c, 10, "balcony", 12, s); got.Status != "candidate" {
		t.Fatal("preferred maximum wrongly excludes sun", got)
	}
	if got := assessCrop(c, 10, "balcony", 4, s); got.Status != "not_fit" {
		t.Fatal(got)
	}
	if got := assessCrop(c, 12, "rooftop", 8, s); got.Status != "out_of_season" {
		t.Fatal(got)
	}
	s.Method = "seedling"
	if got := assessCrop(c, 10, "rooftop", 8, s); got.Status != "out_of_season" {
		t.Fatal("transplant reused sowing month", got)
	}
	if got := assessCrop(c, 11, "rooftop", 8, s); got.Status != "candidate" {
		t.Fatal(got)
	}
	c.TransplantMonths = nil
	if got := assessCrop(c, 11, "rooftop", 8, s); got.Status != "needs_evidence" {
		t.Fatal("unknown calendar accepted", got)
	}
	s.Method = "seed"
	s.Depth = nil
	if got := assessCrop(c, 10, "rooftop", 8, s); got.Status != "needs_setup" {
		t.Fatal("unknown input accepted", got)
	}
	s.Depth = fp(5)
	if got := assessCrop(c, 10, "rooftop", 8, s); got.Status != "not_fit" {
		t.Fatal("small pot accepted", got)
	}
	s.Depth = fp(35)
	c.MinDepthCm = nil
	if got := assessCrop(c, 10, "rooftop", 8, s); got.Status != "conditional" {
		t.Fatal("unknown requirement accepted", got)
	}
	s.Drainage = bp(false)
	if got := assessCrop(c, 10, "rooftop", 8, s); got.Status != "not_fit" {
		t.Fatal(got)
	}
	c.CoreVerified = false
	if got := assessCrop(c, 10, "rooftop", 8, s); got.Status != "needs_evidence" {
		t.Fatal("report-only entry recommended", got)
	}
}
func TestResearchCatalogIntegrity(t *testing.T) {
	var doc researchDocument
	if err := json.Unmarshal(cropCatalogJSON, &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Crops) != 10 {
		t.Fatal("expected ten report crops")
	}
	seen := map[string]bool{}
	for _, c := range doc.Crops {
		if seen[c.Slug] || c.Slug == "" {
			t.Fatal("invalid stable slug")
		}
		seen[c.Slug] = true
		for _, month := range append(c.SeedMonths, c.TransplantMonths...) {
			if month < 1 || month > 12 {
				t.Fatal("invalid month")
			}
		}
		for _, id := range c.SourceIDs {
			if _, ok := doc.Sources[id]; !ok {
				t.Fatal("missing source", id)
			}
		}
		for _, ids := range c.FieldSources {
			for _, id := range ids {
				if _, ok := doc.Sources[id]; !ok {
					t.Fatal("missing field source", id)
				}
			}
		}
		if c.CoreVerified && (c.MinSunHours == nil || c.ContainerSupported == nil) {
			t.Fatal("incomplete core evidence", c.Slug)
		}
	}
}
func TestGardenSetupValidation(t *testing.T) {
	if !validGardenSetup(gardenSetup{Method: "seed"}) {
		t.Fatal("unknown values should be allowed")
	}
	if validGardenSetup(gardenSetup{Method: "seed", Depth: fp(0)}) || validGardenSetup(gardenSetup{Method: "seed", Volume: fp(-2)}) || validGardenSetup(gardenSetup{Method: "automatic"}) {
		t.Fatal("invalid setup accepted")
	}
}

// Missing user input outranks a conditional unknown crop requirement, regardless of field order.
func TestMissingSetupCannotBecomeShortlist(t *testing.T) {
	c := testResearchCrop()
	c.MinDepthCm = nil
	s := gardenSetup{Method: "seed", Drainage: bp(true)}
	if got := assessCrop(c, 10, "rooftop", 8, s); got.Status != "needs_setup" {
		t.Fatal(got)
	}
	s.Volume = fp(10)
	s.Drainage = nil
	if got := assessCrop(c, 10, "rooftop", 8, s); got.Status != "needs_setup" {
		t.Fatal(got)
	}
	s.Depth = fp(35)
	c = testResearchCrop()
	c.MinVolumeL = nil
	if got := assessCrop(c, 10, "rooftop", 8, s); got.Status != "needs_setup" {
		t.Fatal(got)
	}
}

func TestGardenSetupStorageBounds(t *testing.T) {
	for _, setup := range []gardenSetup{{Method: "seed", Depth: fp(0.001)}, {Method: "seed", Volume: fp(0.0001)}} {
		if validGardenSetup(setup) {
			t.Fatal("value would round to zero in the database", setup)
		}
	}
	if !validGardenSetup(gardenSetup{Method: "seed", Depth: fp(0.01), Volume: fp(0.001)}) {
		t.Fatal("valid storage minimum rejected")
	}
}
