"""Regenerate idempotent migration 009 after reviewing backend/crop_catalog.json.
Does not connect to a database. Existing crop IDs and user plants are preserved.
"""
import json,re
from pathlib import Path
root=Path(__file__).resolve().parents[1]
doc=json.loads((root/'backend/crop_catalog.json').read_text())
go=(root/'backend/crop_research.go').read_text()
def q(v): return "CONVERT(0x"+v.encode().hex()+" USING utf8mb4) COLLATE utf8mb4_unicode_ci"
sql=['-- Reviewed crop catalog and user-reported pot setup. Safe to reapply.','-- Back up first. Does not delete or renumber existing crops/plants.','USE chadkrishi;']
for const in ['cropEvidenceDDL','gardenSetupDDL']:
 sql.append(re.search(r'const '+const+r'\s*=\s*`([^`]+)`',go).group(1)+';')
sql.append('START TRANSACTION;')
for c in doc['crops']:
 vals=[q(c['legacyName']),q(c['bengaliName']),q(c['varietyNote']),"'[]'",'0','24',"'[\"rooftop\",\"balcony\"]'",q(c['containerNote']),q(c['careNote']),q('Research catalog: see crop_evidence and field source links'),"'source-pending'"]
 sql.append('INSERT INTO crops(name,bengali_name,description,planting_months,min_sunlight_hours,max_sunlight_hours,suitable_spaces,container_guidance,care_guidance,source_reference,verification_status) VALUES('+','.join(vals)+') ON DUPLICATE KEY UPDATE id=id;')
 sql.append('INSERT INTO crop_evidence(slug,crop_id,data) SELECT '+q(c['slug'])+',id,'+q(json.dumps(c,ensure_ascii=False))+' FROM crops WHERE name='+q(c['legacyName'])+' ON DUPLICATE KEY UPDATE data=VALUES(data);')
sql.append('COMMIT;')
(root/'migrations/009_crop_research.sql').write_text('\n\n'.join(sql)+'\n')
print('Generated migrations/009_crop_research.sql for',len(doc['crops']),'crops')
