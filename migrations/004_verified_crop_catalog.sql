USE defaultdb;

ALTER TABLE crops
  MODIFY COLUMN verification_status ENUM('demo', 'reviewed', 'source-backed', 'source-pending') NOT NULL DEFAULT 'source-pending',
  ADD COLUMN bengali_name VARCHAR(100) NULL AFTER name,
  ADD COLUMN variety_caveat VARCHAR(300) NULL AFTER bengali_name,
  ADD COLUMN region_note VARCHAR(300) NULL AFTER planting_months,
  ADD COLUMN container_suitability ENUM('suitable', 'conditional', 'unknown') NOT NULL DEFAULT 'unknown' AFTER suitable_spaces,
  ADD COLUMN container_depth_inches DECIMAL(5, 1) NULL AFTER container_suitability,
  ADD COLUMN container_volume_gallons DECIMAL(5, 1) NULL AFTER container_depth_inches,
  ADD COLUMN temperature_guidance VARCHAR(300) NULL AFTER container_volume_gallons,
  ADD COLUMN source_url VARCHAR(500) NULL AFTER source_reference,
  ADD COLUMN source_notes TEXT NULL AFTER source_url,
  ADD COLUMN reviewed_date DATE NULL AFTER source_notes;

UPDATE crops
SET
  bengali_name = CASE name
    WHEN 'Lal shak' THEN 'লাল শাক'
    WHEN 'Chili' THEN 'মরিচ'
    WHEN 'Tomato' THEN 'টমেটো'
  END,
  variety_caveat = 'Variety-specific performance and exact container size are not verified.',
  region_note = 'Bangladesh city-level recommendation; confirm local conditions.',
  container_suitability = 'conditional',
  source_url = 'https://extension.umd.edu/resource/growing-vegetables-containers',
  source_notes = 'UMD Extension supports the general container, drainage, and sunlight guidance. BAMIS crop-calendar detail must be checked for the selected crop and district before treating planting months as verified.',
  reviewed_date = CURRENT_DATE(),
  verification_status = 'source-pending';

INSERT INTO crops
  (name, bengali_name, variety_caveat, description, planting_months, region_note,
   min_sunlight_hours, max_sunlight_hours, suitable_spaces, container_suitability,
   container_guidance, care_guidance, source_reference, source_url, source_notes,
   reviewed_date, verification_status)
VALUES
  ('Cucumber', 'শসা', 'Cultivar-specific yield and trellis needs vary.',
   'A warm-season vine with a verified Dhaka-region BAMIS calendar window.',
   '[11,12,1,2,3]', 'Dhaka-region calendar; other regions may differ.',
   6, 8, '["rooftop","balcony"]', 'conditional',
   'Use a container with drainage and verify the volume/depth for the chosen cultivar; provide trellis support.',
   'Keep foliage ventilated, monitor moisture, and support the vine.',
   'UMD Extension: warm-season container vegetables generally need 6–8 hours of direct sun.',
   'https://www.bamis.gov.bd/res/public/calendars/2023/04/16/26985.pdf',
   'BAMIS official Crop Weather Calendar, Cucumber, Dhaka PDF (page 1); sunlight and container principles from UMD Extension container guidance. Exact cultivar container volume is unknown.',
   CURRENT_DATE(), 'source-backed')
ON DUPLICATE KEY UPDATE
  bengali_name = VALUES(bengali_name),
  description = VALUES(description),
  planting_months = VALUES(planting_months),
  source_url = VALUES(source_url),
  source_notes = VALUES(source_notes),
  reviewed_date = VALUES(reviewed_date),
  verification_status = VALUES(verification_status);
