"""Only for a disposable test database; creates accounts and plants.
Run after migrations 001-009: python scripts/crop_smoke.py --test-only --base http://127.0.0.1:8081/api/v1
"""
import argparse,secrets
from datetime import datetime,timezone,timedelta
from api_smoke import Client

def main():
 p=argparse.ArgumentParser();p.add_argument('--test-only',required=True,action='store_true');p.add_argument('--base',default='http://127.0.0.1:8081/api/v1');args=p.parse_args()
 a,b=Client(args.base),Client(args.base);password=secrets.token_urlsafe(16)
 for c in [a,b]:
  c.call('POST','/auth/register',{'name':'Crop Planner','email':secrets.token_hex(8)+'@example.test','password':password},201)
  c.call('PUT','/profile',{'name':'Crop Planner','district':'Dhaka','latitude':23.8103,'longitude':90.4125,'gardeningSpace':'rooftop','sunlightHours':8,'rainExposure':True,'bio':''})
 assert a.call('GET','/garden-setup')['depthCm'] is None
 setup={'depthCm':40,'volumeL':25,'drainage':True,'support':True,'method':'seed'}
 a.call('PUT','/garden-setup',setup)
 assert a.call('GET','/garden-setup')==setup
 assert b.call('GET','/garden-setup')['depthCm'] is None
 a.call('PUT','/garden-setup',{**setup,'depthCm':-3},400)
 a.call('PUT','/garden-setup',{**setup,'depthCm':0.001},400)
 rounded=a.call('PUT','/garden-setup',{**setup,'depthCm':40.126,'volumeL':25.1234})
 assert rounded['depthCm']==40.13 and rounded['volumeL']==25.123
 assert a.call('GET','/garden-setup')==rounded
 a.call('PUT','/garden-setup',setup)
 rec=a.call('GET','/recommendations')
 assert len(rec['cropCatalog'])==10 and len({c['slug'] for c in rec['cropCatalog']})==10
 assert rec['weather']['status'] in ['available','unavailable','stale']
 crops={c['slug']:c for c in rec['cropCatalog']}
 assert crops['radish']['minDepthCm'] is None
 assert crops['tomato']['status']=='needs_evidence'
 assert all(c['status'] in ['candidate','conditional'] for c in rec['crops'])
 a.call('PUT','/garden-setup',{**setup,'method':'seedling'})
 rec=a.call('GET','/recommendations')
 assert all(c['status']=='needs_evidence' for c in rec['cropCatalog'])
 today=datetime.now(timezone(timedelta(hours=6))).strftime('%Y-%m-%d')
 plant={'cropId':crops['cucumber']['id'],'nickname':'Research cucumber','plantingDate':today,'propagationMethod':'seed','growthStage':'growing','note':'','photoPath':''}
 pid=a.call('POST','/plants',plant,201)['id']
 saved=next(p for p in a.call('GET','/plants')['plants'] if p['id']==pid)
 assert saved['careGuidance']==crops['cucumber']['careNote'] and saved['verificationStatus']=='core-reviewed; check setup'
 tasks=a.call('GET','/tasks')['tasks'];notes=[t for t in tasks if t['plantId']==pid and t['title']=='Review this crop’s care observations']
 assert len(notes)==1 and notes[0]['reason']==crops['cucumber']['careNote']
 a.call('PUT',f'/tasks/{notes[0]["id"]}',{'completed':True},204)
 assert next(t for t in a.call('GET','/tasks')['tasks'] if t['id']==notes[0]['id'])['completed']
 print('PASS 10-crop API, nulls, method separation, setup validation/ownership/persistence, crop note task idempotency')
if __name__=='__main__':main()
