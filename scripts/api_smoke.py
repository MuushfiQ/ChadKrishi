"""Integration test against an isolated running installation. Creates test records.
Run: python scripts/api_smoke.py --test-only
Optional TEST_MYSQL_BIN / TEST_DB_PORT permit admin and expiry fixtures locally.
Never target production. Uses Python standard library only.
"""
import argparse, base64, concurrent.futures, http.cookiejar, json, os, secrets, subprocess, urllib.request, urllib.error
from datetime import datetime, timedelta, timezone
from pathlib import Path

class Client:
    def __init__(self, base):
        self.base=base; self.jar=http.cookiejar.CookieJar()
        self.http=urllib.request.build_opener(urllib.request.ProxyHandler({}),urllib.request.HTTPCookieProcessor(self.jar))
    def call(self, method, path, data=None, status=200, headers=None):
        h={'Origin':os.getenv('TEST_ORIGIN','http://127.0.0.1:5173')}
        if data is not None and not isinstance(data,bytes): data=json.dumps(data).encode();h['Content-Type']='application/json'
        h.update(headers or {})
        req=urllib.request.Request(self.base+path,data=data,headers=h,method=method)
        try:r=self.http.open(req,timeout=20)
        except urllib.error.HTTPError as e:r=e
        body=r.read(); assert r.status==status,(method,path,r.status,body[:500])
        return json.loads(body) if body and r.headers.get('Content-Type','').startswith('application/json') else body

def main():
    parser=argparse.ArgumentParser();parser.add_argument('--test-only',action='store_true',required=True);parser.add_argument('--base',default='http://127.0.0.1:8080/api/v1');args=parser.parse_args()
    run=secrets.token_hex(5);pwd=secrets.token_urlsafe(16);clients=[];accounts=[]
    for name in ['Seller','Buyer','Other','Admin']:
        c=Client(args.base);email=f'{name.lower()}.{run}@example.test'
        u=c.call('POST','/auth/register',{'name':name+' Test','email':email,'password':pwd},201)['user']
        clients.append(c);accounts.append({'id':u['id'],'email':email,'password':pwd})
    seller,buyer,other,admin=clients
    profile={'name':'Ayesha Rahman','district':'Dhaka','latitude':23.8103,'longitude':90.4125,'gardeningSpace':'rooftop','sunlightHours':6,'rainExposure':True,'bio':'Growing a greener Dhaka, one pot at a time.'}
    seller.call('PUT','/profile',profile)
    assert seller.call('GET','/auth/me')['user']['bio']==profile['bio']
    crop=seller.call('GET','/crops')['crops'][0]['id']
    today=datetime.now(timezone(timedelta(hours=6))).strftime('%Y-%m-%d')
    plant={'cropId':crop,'nickname':'Rooftop companion','plantingDate':today,'propagationMethod':'seedling','growthStage':'growing','note':'PRIVATE NOTE NEVER SHARED','photoPath':''}
    pid=seller.call('POST','/plants',plant,201)['id'];buyer.call('PUT',f'/plants/{pid}',plant,404)
    plant['growthStage']='flowering';seller.call('PUT',f'/plants/{pid}',plant)
    tasks=seller.call('GET','/tasks')['tasks'];assert len(tasks)>=3
    assert [t['id'] for t in tasks]==[t['id'] for t in seller.call('GET','/tasks')['tasks']]
    tid=tasks[0]['id'];buyer.call('PUT',f'/tasks/{tid}',{'completed':True},404)
    seller.call('PUT',f'/tasks/{tid}',{'completed':True},204)
    assert next(t for t in seller.call('GET','/tasks')['tasks'] if t['id']==tid)['completed']
    seller.call('PUT',f'/plants/{pid}',{**plant,'archived':True});assert not seller.call('GET','/tasks')['tasks']
    seller.call('PUT',f'/plants/{pid}',{**plant,'archived':False})
    custom={'plantId':pid,'title':'নিজের কাজ: support check','reason':'My own reminder','date':today}
    buyer.call('POST','/tasks',custom,404)
    seller.call('POST','/tasks',{**custom,'title':'   '},400)
    seller.call('POST','/tasks',{**custom,'date':'2099-01-01'},400)
    custom_id=seller.call('POST','/tasks',custom,201)['id']
    row=next(t for t in seller.call('GET','/tasks')['tasks'] if t['id']==custom_id)
    assert row['custom'] and row['plantId']==pid and row['title']==custom['title']
    assert not buyer.call('GET','/tasks')['tasks']
    buyer.call('DELETE',f'/tasks/{custom_id}',status=404)
    seller.call('DELETE',f'/tasks/{tid}',status=404)
    seller.call('PUT',f'/tasks/{custom_id}',{'completed':True},204)
    assert next(t for t in seller.call('GET','/tasks')['tasks'] if t['id']==custom_id)['completed']
    tomorrow=(datetime.now(timezone(timedelta(hours=6)))+timedelta(days=1)).strftime('%Y-%m-%d')
    future_id=seller.call('POST','/tasks',{**custom,'date':tomorrow},201)['id']
    assert [t['id'] for t in seller.call('GET',f'/tasks?date={tomorrow}')['tasks']]==[future_id]
    seller.call('DELETE',f'/tasks/{custom_id}',status=204)
    seller.call('PUT',f'/plants/{pid}',{**plant,'archived':True})
    seller.call('POST','/tasks',custom,404)
    assert not seller.call('GET',f'/tasks?date={tomorrow}')['tasks']
    seller.call('PUT',f'/plants/{pid}',{**plant,'archived':False})
    print('PASS custom task ownership, validation, persistence, completion, delete, future scheduling and archive')
    print('PASS auth/profile, plant ownership, care idempotency/completion/archive')
    boundary='testboundary';png=base64.b64decode('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jX1sAAAAASUVORK5CYII=')
    # Generate a valid RGB PNG with standard library, no external imaging dependency.
    import zlib,struct
    def chunk(typ,data):return struct.pack('!I',len(data))+typ+data+struct.pack('!I',zlib.crc32(typ+data)&0xffffffff)
    png=b'\x89PNG\r\n\x1a\n'+chunk(b'IHDR',struct.pack('!2I5B',4,4,8,2,0,0,0))+chunk(b'IDAT',zlib.compress((b'\x00'+bytes([110,160,95])*4)*4))+chunk(b'IEND',b'')
    def upload(c,raw,expected):
        body=(f'--{boundary}\r\nContent-Disposition: form-data; name="image"; filename="test.png"\r\nContent-Type: image/png\r\n\r\n').encode()+raw+f'\r\n--{boundary}--\r\n'.encode()
        return c.call('POST','/uploads',body,expected,{'Content-Type':'multipart/form-data; boundary='+boundary})
    upload(seller,b'<svg>bad</svg>',400);photo=upload(seller,png,201)['path']
    def media(c,path,status):
        base=c.base;c.base=base.split('/api/v1')[0]
        try:return c.call('GET',path,status=status)
        finally:c.base=base
    media(buyer,photo,404)
    # Photo ownership, save/remove persistence, and member media access before any post exists.
    buyer.call('PUT','/profile/photo',{'photoPath':photo},400)
    seller.call('PUT','/profile/photo',{},400)
    seller.call('PUT','/profile/photo',{'photoPath':'https://example.test/fake.png'},400)
    seller.call('PUT','/profile/photo',{'photoPath':photo})
    assert seller.call('GET','/auth/me')['user']['profilePhoto']==photo
    public=buyer.call('GET',f'/people/{accounts[0]["id"]}')['person']
    assert public['profilePhoto']==photo and not public['isSelf'] and public['postCount']==0
    assert seller.call('GET',f'/people/{accounts[0]["id"]}')['person']['isSelf']
    assert 'email' not in public and 'latitude' not in public
    media(buyer,photo,200)
    seller.call('PUT','/profile/photo',{'photoPath':''})
    assert not seller.call('GET','/auth/me')['user'].get('profilePhoto')
    media(buyer,photo,404)
    seller.call('PUT','/profile/photo',{'photoPath':photo})
    # An incomplete new member can upload without completing their garden profile.
    other_photo=upload(other,png,201)['path']
    other.call('PUT','/profile/photo',{'photoPath':other_photo})
    assert other.call('GET','/auth/me')['user']['profilePhoto']==other_photo
    print('PASS profile picture ownership, persistence, removal, public access and incomplete garden setup')
    post={'type':'update','title':'A little green in the city','content':'Fresh leaves, morning light, and a new beginning. আমাদের ছাদবাগান 🌱','imagePath':photo,'plantId':pid}
    postid=seller.call('POST','/posts',post,201)['id'];media(buyer,photo,200)
    public=buyer.call('GET',f'/people/{accounts[0]["id"]}')['person'];assert 'email' not in public and 'latitude' not in public
    buyer.call('PUT',f'/posts/{postid}',post,404)
    for reaction in ['like','love','love','angry','sad']:
        buyer.call('PUT',f'/posts/{postid}/reaction',{'reaction':reaction},204)
    feed=buyer.call('GET','/posts')['posts'];p=next(p for p in feed if p['id']==postid)
    assert p['authorPhoto']==photo
    assert p['likeCount']==1 and p['reaction']=='sad' and p['reactions']['sad']==1 and p['reactions']['love']==0
    assert 'PRIVATE NOTE' not in json.dumps(feed)
    cid=buyer.call('POST',f'/posts/{postid}/comments',{'content':'Beautiful garden! খুব সুন্দর।'},201)['id'];other.call('DELETE',f'/comments/{cid}',status=404)
    buyer.call('DELETE',f'/comments/{cid}',status=204)
    buyer.call('POST',f'/posts/{postid}/comments',{'content':'Love this little green corner!'},201)
    author_posts=buyer.call('GET',f'/posts?author={accounts[0]["id"]}')['posts']
    assert author_posts and all(p['authorId']==accounts[0]['id'] for p in author_posts)
    assert buyer.call('GET',f'/people/{accounts[0]["id"]}')['person']['postCount']==1
    assert buyer.call('GET',f'/posts/{postid}/comments')['comments'][0]['authorPhoto']==''
    print('PASS image validation/privacy, Unicode caption, reaction replacement, comments, safe public profiles')
    listing={'type':'sale','title':'Healthy rooftop seedling','content':'A lovingly grown plant looking for a new home. Collection in Dhaka.','priceBdt':250,'quantity':2,'contactMethod':'Please message me here.','plantId':pid,'imagePath':photo}
    sale=seller.call('POST','/posts',listing,201)['id'];seller.call('POST',f'/posts/{sale}/buy',status=400)
    def attempt(c):
        req=urllib.request.Request(c.base+f'/posts/{sale}/buy',data=b'{}',headers={'Origin':os.getenv('TEST_ORIGIN','http://127.0.0.1:5173'),'Content-Type':'application/json'},method='POST')
        try:r=c.http.open(req,timeout=20)
        except urllib.error.HTTPError as e:r=e
        return c,r.status,json.loads(r.read())
    with concurrent.futures.ThreadPoolExecutor(2) as pool:results=list(pool.map(attempt,[buyer,other]))
    assert sorted(r[1] for r in results)==[201,409],results
    winner,_,result=next(r for r in results if r[1]==201);loser=other if winner is buyer else buyer;oid=result['id']
    n=seller.call('GET','/notifications')
    assert n['unreadCount']==1 and n['notifications'][0]['kind']=='purchase'
    purchase_note=n['notifications'][0]['id']
    assert not winner.call('GET','/notifications')['notifications']
    loser.call('PUT',f'/notifications/{purchase_note}/read',status=404)
    assert seller.call('GET',f'/orders?id={oid}')['orders'][0]['id']==oid
    assert not loser.call('GET',f'/orders?id={oid}')['orders']

    assert any(p['id']==sale and p['saleStatus']=='reserved' for p in buyer.call('GET','/posts?type=sale')['posts'])
    seller.call('PUT',f'/posts/{sale}',listing,409)
    seller.call('PUT',f'/posts/{sale}/status',{'status':'sold'},409)
    seller.call('PUT',f'/orders/{oid}',{'action':'received'},409)
    seller.call('PUT',f'/orders/{oid}',{'action':'paid','method':'cash'},409)
    loser.call('PUT',f'/orders/{oid}',{'action':'received'},404)
    winner.call('PUT',f'/orders/{oid}',{'action':'paid','method':'bkash','reference':'TEST-REF'})
    assert seller.call('GET','/notifications')['notifications'][0]['kind']=='order_paid'

    assert any(p['id']==sale and p['saleStatus']=='reserved' for p in buyer.call('GET','/posts?type=sale')['posts'])
    winner.call('PUT',f'/orders/{oid}',{'action':'received'},409)
    winner.call('PUT',f'/orders/{oid}',{'action':'cancel'},409)
    seller.call('PUT',f'/orders/{oid}',{'action':'not_received'})
    assert winner.call('GET','/notifications')['notifications'][0]['kind']=='order_not_received'

    winner.call('PUT',f'/orders/{oid}',{'action':'paid','method':'cash'})
    seller.call('PUT',f'/orders/{oid}',{'action':'received'})
    assert winner.call('GET','/notifications')['notifications'][0]['kind']=='order_received'

    seller.call('PUT',f'/orders/{oid}',{'action':'received'},409)
    assert any(p['id']==sale and p['saleStatus']=='sold' for p in buyer.call('GET','/posts?type=sale')['posts'])
    for c in [seller,winner]:assert next(o for o in c.call('GET','/orders')['orders'] if o['id']==oid)['status']=='completed'
    seller.call('DELETE',f'/posts/{sale}',status=409)
    sale2=seller.call('POST','/posts',{**listing,'title':'Fresh mint for your balcony'},201)['id']
    o2=buyer.call('POST',f'/posts/{sale2}/buy',status=201)['id'];buyer.call('PUT',f'/orders/{o2}',{'action':'cancel'})
    assert any(p['id']==sale2 for p in buyer.call('GET','/posts?status=available')['posts'])
    assert seller.call('GET','/notifications')['notifications'][0]['kind']=='order_cancel'
    seller.call('PUT',f'/notifications/{purchase_note}/read',status=204)
    snapshot=seller.call('GET','/notifications')
    seller.call('PUT','/notifications/read-all',{'through':snapshot['latestId']},204)
    assert seller.call('GET','/notifications')['unreadCount']==0
    print('PASS transactional order notifications, counterparty targeting, ownership and read states')

    print('PASS concurrent reservation, role/state guards, payment retry/receipt, Sold, immutable history, cancellation')
    conv=buyer.call('POST','/conversations',{'userId':accounts[0]['id']})['id']
    assert seller.call('POST','/conversations',{'userId':accounts[1]['id']})['id']==conv
    assert next(t for t in buyer.call('GET','/conversations')['conversations'] if t['id']==conv)['profilePhoto']==photo
    other.call('GET',f'/conversations/{conv}/messages',status=404)
    other.call('POST',f'/conversations/{conv}/messages',{'content':'intrusion'},404)
    mid=buyer.call('POST',f'/conversations/{conv}/messages',{'content':'Is the seedling ready for collection? গাছটি নিতে চাই।'},201)['id']
    msg=seller.call('GET',f'/conversations/{conv}/messages')['messages'][0];assert msg['id']==mid and not msg['mine']
    first_note=seller.call('GET','/notifications')['notifications'][0]
    assert first_note['kind']=='message' and first_note['link']==f'/messages?thread={conv}' and not first_note['read']
    later=buyer.call('POST',f'/conversations/{conv}/messages',{'content':'Later message'},201)['id']
    seller.call('PUT','/notifications/read-all',{'through':first_note['id']},204)
    assert seller.call('GET','/notifications')['unreadCount']==1

    seller.call('PUT',f'/conversations/{conv}/read',{'through':mid},204)
    assert next(m for m in buyer.call('GET',f'/conversations/{conv}/messages')['messages'] if m['id']==mid)['read']
    assert seller.call('GET','/notifications')['unreadCount']==1
    seller.call('PUT',f'/conversations/{conv}/read',{'through':later},204)
    assert seller.call('GET','/notifications')['unreadCount']==0
    seller.call('POST',f'/conversations/{conv}/messages',{'content':'Yes! Let’s arrange a convenient time.'},201)
    # Pagination on actual records, including older chat messages.
    for i in range(51):buyer.call('POST',f'/conversations/{conv}/messages',{'content':f'History check {i+1}'},201)
    first=buyer.call('GET',f'/conversations/{conv}/messages');assert len(first['messages'])==50 and first['nextCursor']
    older=buyer.call('GET',f'/conversations/{conv}/messages?before={first["nextCursor"]}')
    assert len(older['messages'])==4 and not {m['id'] for m in first['messages']}&{m['id'] for m in older['messages']}
    seller.call('GET','/admin/listings',status=403);seller.call('PUT',f'/admin/listings/{sale}/archive',status=403)
    nf=seller.call('GET','/notifications')
    assert len(nf['notifications'])==20 and nf['nextCursor'] and nf['unreadCount']==51
    no=seller.call('GET',f'/notifications?before={nf["nextCursor"]}')
    assert not {n['id'] for n in nf['notifications']}&{n['id'] for n in no['notifications']}
    assert not admin.call('GET','/notifications')['notifications']
    print('PASS durable message notifications, bounded mark-all/read receipts and notification pagination')
    print('PASS private chat membership, unique conversation, read receipts, message pagination, admin authorization')
    dbbin=os.getenv('TEST_MYSQL_BIN')
    if dbbin:
        def sql(query):subprocess.run([dbbin,'--protocol=tcp','-h127.0.0.1','-P'+os.getenv('TEST_DB_PORT','3307'),'-uroot','chadkrishi','-e',query],check=True,capture_output=True)
        sql(f'UPDATE users SET is_admin=TRUE WHERE id={accounts[3]["id"]}')
        assert admin.call('GET','/auth/me')['user']['isAdmin']
        admin.call('PUT',f'/admin/listings/{sale2}/archive',status=409)
        admin.call('PUT',f'/admin/listings/{sale}/archive',status=204)
        assert all(p['id']!=sale for p in buyer.call('GET','/posts')['posts'])
        assert next(o for o in winner.call('GET','/orders')['orders'] if o['id']==oid)['archived']
        # Force notification storage failure: the triggering write must roll back.
        rollback_sale=seller.call('POST','/posts',listing,201)['id']
        rollback_order=buyer.call('POST',f'/posts/{sale2}/buy',status=201)['id']
        before_messages=buyer.call('GET',f'/conversations/{conv}/messages')['messages']
        sql('RENAME TABLE notifications TO test_notifications_backup')
        try:
            buyer.call('POST',f'/conversations/{conv}/messages',{'content':'Must roll back'},503)
            buyer.call('PUT',f'/orders/{rollback_order}',{'action':'cancel'},503)
            buyer.call('POST',f'/posts/{rollback_sale}/buy',status=503)
        finally:
            sql('RENAME TABLE test_notifications_backup TO notifications')
        assert buyer.call('GET',f'/conversations/{conv}/messages')['messages']==before_messages
        buyer.call('PUT',f'/orders/{rollback_order}',{'action':'cancel'})
        assert any(p['id']==sale2 and p['saleStatus']=='available' for p in buyer.call('GET','/posts?type=sale')['posts'])
        assert any(p['id']==rollback_sale and p['saleStatus']=='available' for p in buyer.call('GET','/posts?type=sale')['posts'])
        print('PASS notification failure rolls back message, purchase and cancellation atomically')
        # Expired sessions must not access data; preserve the other test account sessions.
        sql(f'UPDATE sessions SET expires_at=DATE_SUB(UTC_TIMESTAMP(),INTERVAL 1 DAY) WHERE user_id={accounts[2]["id"]}')
        other.call('GET','/plants',status=401)
        print('PASS admin archive preserves history, expired session denied')
    seller.call('POST','/auth/logout',status=204)
    seller.call('GET','/plants',status=401)
    seller.call('POST','/auth/login',{'email':accounts[0]['email'],'password':pwd})
    assert seller.call('GET','/plants')['plants'][0]['nickname']==plant['nickname']
    Path(os.getenv('TEST_FIXTURE_PATH','/tmp/chadkrishi-fixture.json')).write_text(json.dumps({'accounts':accounts,'postId':postid,'saleId':sale2,'thread':conv,'orderId':oid}))
    print('PASS logout/relogin persistence. Integration suite complete.')
if __name__=='__main__':main()
