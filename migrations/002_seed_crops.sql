USE defaultdb;

INSERT INTO crops
  (name, description, planting_months, min_sunlight_hours, max_sunlight_hours,
   suitable_spaces, container_guidance, care_guidance, source_reference, verification_status)
VALUES
  ('Lal shak', 'A fast leafy green suitable for containers.', '[10,11,12,1,2,3]', 3, 7,
   '["rooftop","balcony"]', 'Use a shallow container with drainage holes.',
   'Check soil moisture and harvest outer leaves first.', 'Demonstration catalog; verify locally.', 'demo'),
  ('Chili', 'A compact fruiting plant for a sunny container.', '[2,3,4,5,6,7,8]', 5, 8,
   '["rooftop","balcony"]', 'Use a deep container and provide a support stake.',
   'Inspect leaves and support the growing stem.', 'Demonstration catalog; verify locally.', 'demo'),
  ('Tomato', 'A sun-loving crop for a larger rooftop container.', '[10,11,12,1,2]', 6, 8,
   '["rooftop"]', 'Use a large container and a sturdy support.',
   'Inspect leaves and check support as the plant grows.', 'Demonstration catalog; verify locally.', 'demo')
ON DUPLICATE KEY UPDATE
  description = VALUES(description),
  planting_months = VALUES(planting_months),
  verification_status = VALUES(verification_status);

