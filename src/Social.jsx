import { useEffect, useRef, useState } from 'react'
import { Link, useSearchParams, useNavigate } from 'react-router-dom'
import api, { apiMessage } from './api'
import { Avatar } from './Avatar'
import { useLiveList, announceActivity } from './live'

const labels={waiting_payment:'Waiting for payment',buyer_paid:'Buyer marked Paid · awaiting seller',completed:'Payment received · Sold',cancelled:'Cancelled'}
function ErrorBox({error}){return error?<p className="form-error" role="alert">{error}</p>:null}
export function MessageButton({userId}){
 const navigate=useNavigate();const [error,setError]=useState(''),[busy,setBusy]=useState(false)
 async function start(){setBusy(true);setError('');try{const r=await api.post('/conversations',{userId});navigate(`/messages?thread=${r.data.id}`)}catch(e){setError(apiMessage(e))}finally{setBusy(false)}}
 return <><button className="btn btn-soft" disabled={busy} onClick={start}>✉ Message</button><ErrorBox error={error}/></>
}
function OrderCard({order:o,onUpdate}){
 const [method,setMethod]=useState('cash'),[reference,setReference]=useState(''),[busy,setBusy]=useState(false),[error,setError]=useState('')
 async function act(action){if(action==='received'&&!window.confirm('Confirm that you actually received the full payment? This marks the listing Sold.'))return;setBusy(true);setError('');try{await api.put(`/orders/${o.id}`,{action,method,reference});onUpdate();announceActivity()}catch(e){setError(apiMessage(e))}finally{setBusy(false)}}
 return <article className="card-surface order-card"><div className="section-heading"><span className="eyebrow">{o.isBuyer?'PURCHASE':'SALE'} · #{o.id}</span><span className={`status-pill order-${o.status}`}>{labels[o.status]}</span></div><h2>{o.title}</h2><p className="listing-price">৳ {Number(o.amountBdt).toLocaleString('en-BD')} <small>total · {o.quantity} item(s)</small></p><p>{o.isBuyer?'Seller':'Buyer'}: <Link to={`/people/${o.isBuyer?o.sellerId:o.buyerId}`}>{o.isBuyer?o.sellerName:o.buyerName}</Link> · {new Date(o.createdAt).toLocaleDateString()}</p>{o.archived&&<p className="muted">Listing archived. Your transaction record is retained.</p>}{o.method&&<p className="payment-detail">Reported method: <strong>{o.method}</strong>{o.reference&&<> · Reference: {o.reference}</>}</p>}<ErrorBox error={error}/>
 {o.isBuyer&&o.status==='waiting_payment'&&<div className="payment-form"><p>Pay the seller directly, then mark Paid. Do not enter any PIN, OTP or password here.</p><div className="form-grid"><label>How did you pay?<select value={method} onChange={e=>setMethod(e.target.value)}><option value="cash">Cash</option><option value="bkash">bKash</option><option value="other">Other</option></select></label><label>Transaction reference (optional)<input maxLength={160} value={reference} onChange={e=>setReference(e.target.value)}/></label></div><button className="btn btn-leaf" disabled={busy} onClick={()=>act('paid')}>I have paid</button></div>}
 {!o.isBuyer&&['waiting_payment','buyer_paid'].includes(o.status)&&<div className="payment-form"><strong>Payment received?</strong><p>{o.status==='waiting_payment'?'Waiting for the buyer to mark Paid.':'Check your cash or payment account before confirming.'}</p><button className="btn btn-leaf" disabled={busy||o.status!=='buyer_paid'} onClick={()=>act('received')}>Yes, payment received</button>{o.status==='buyer_paid'&&<button disabled={busy} className="btn btn-soft" onClick={()=>act('not_received')}>Not received — ask buyer to check</button>}</div>}
 <div className="card-actions"><MessageButton userId={o.isBuyer?o.sellerId:o.buyerId}/>{o.status==='waiting_payment'&&<button disabled={busy} className="text-button danger-text" onClick={()=>{if(window.confirm('Cancel this reservation?'))act('cancel')}}>Cancel reservation</button>}</div></article>
}
export function OrdersPage(){
 const [params,setParams]=useSearchParams()
 const selected=params.get('order')||undefined
 const [filter,setFilter]=useState('all')
 const {items,next,error,loading,busy,refresh:load,loadMore}=useLiveList('/orders','orders',{id:selected},5000)
 useEffect(()=>setFilter('all'),[selected])
 const visible=items.filter(o=>filter==='all'||(filter==='buy'?o.isBuyer:!o.isBuyer))
 return <main className="page-content"><div className="page-heading"><div><span className="eyebrow">Your garden exchanges</span><h1>Buy &amp; sell history</h1><p className="muted">Every reservation and payment confirmation, in one place.</p></div><button className="btn btn-soft" disabled={busy} onClick={()=>load()}>Refresh</button></div>{selected&&<p className="market-note">Showing order #{selected}. <button className="text-button" onClick={()=>setParams({})}>View all history</button></p>}<p className="live-hint">Payment status updates automatically.</p><p className="market-note">Payments happen directly between members. “Paid” is the buyer’s declaration; “Payment received” is the seller’s confirmation.</p><div className="garden-tabs">{[['all','All activity'],['buy','My purchases'],['sell','My sales']].map(([v,l])=><button key={v} className={filter===v?'selected':''} onClick={()=>setFilter(v)}>{l}</button>)}</div><ErrorBox error={error}/>{loading?<p>Loading history…</p>:!visible.length?<div className="card-surface empty-state"><h2>No {filter==='all'?'transactions':filter==='buy'?'purchases':'sales'} on this page</h2><Link to="/marketplace">Explore marketplace →</Link></div>:<div className="order-list">{visible.map(o=><OrderCard key={o.id} order={o} onUpdate={()=>load()}/>)}</div>}{next&&<button className="btn btn-soft" disabled={busy} onClick={loadMore}>Load older history</button>}</main>
}
export function AdminPage(){
 const [items,setItems]=useState([]),[next,setNext]=useState(null),[error,setError]=useState(''),[busy,setBusy]=useState(false)
 async function load(before){setBusy(true);setError('');try{const r=await api.get('/admin/listings',{params:{before}});setItems(p=>before?[...p,...r.data.listings]:r.data.listings);setNext(r.data.nextCursor)}catch(e){setError(apiMessage(e))}finally{setBusy(false)}}
 useEffect(()=>{load()},[])
 async function archive(id){setBusy(true);try{await api.put(`/admin/listings/${id}/archive`);await load()}catch(e){setError(apiMessage(e))}finally{setBusy(false)}}
 return <main className="page-content"><span className="eyebrow">Administrator</span><h1>Marketplace archive</h1><p className="muted">Archive sold listings from the feed and market. Buyer and seller histories remain available.</p><ErrorBox error={error}/><div className="order-list">{items.map(p=><article className="card-surface order-card" key={p.id}><h2>{p.title}</h2><p>{p.sellerName} · {p.archived?'Archived':'Sold'}</p><button className="btn btn-soft" disabled={busy||p.archived} onClick={()=>archive(p.id)}>{p.archived?'Archived':'Archive listing'}</button></article>)}</div>{!busy&&!error&&!items.length&&<p>No sold listings yet.</p>}{next&&<button className="btn btn-soft" disabled={busy} onClick={()=>load(next)}>More listings</button>}</main>
}
export function MessagesPage(){
 const [params,setParams]=useSearchParams();const thread=Number(params.get('thread'))||0
 const [messages,setMessages]=useState([]),[next,setNext]=useState(null),[content,setContent]=useState(''),[error,setError]=useState(''),[busy,setBusy]=useState(false),[ready,setReady]=useState(false)
 const selected=useRef(thread);selected.current=thread
 const initialized=useRef(false)
 const messagePane=useRef(null),lastSeen=useRef(0)
 useEffect(()=>{const el=messagePane.current;const latest=messages.at(-1);if(!el||!latest)return;if(latest.id>lastSeen.current&&(lastSeen.current===0||latest.mine||el.scrollHeight-el.scrollTop-el.clientHeight<160))el.scrollTop=el.scrollHeight;lastSeen.current=latest.id},[messages])
 const {items:threads,next:threadNext,error:inboxError,refresh:refreshInbox,loadMore:olderThreads}=useLiveList('/conversations','conversations',{},4000)
 function loadThreads(before){return before?olderThreads():refreshInbox()}
 useEffect(()=>{
  let stopped=false,inflight=false,latest=0
  initialized.current=false;lastSeen.current=0;setMessages([]);setNext(null);setReady(false);setError('');setContent('')
  if(!thread)return
  async function poll(){
   if(stopped||inflight||document.hidden)return
   inflight=true
   try{
    const r=await api.get(`/conversations/${thread}/messages`)
    if(stopped)return
    let incoming=r.data.messages
    // Catch up after a hidden/offline tab even when more than one page arrived.
    let cursor=r.data.nextCursor
    while(latest&&incoming.length&&incoming.at(-1).id>latest&&cursor){
     const older=await api.get(`/conversations/${thread}/messages`,{params:{before:cursor}})
     if(stopped)return
     incoming=[...incoming,...older.data.messages]
     if(older.data.nextCursor===cursor)break
     cursor=older.data.nextCursor
    }
    setMessages(prev=>{const map=new Map(prev.map(m=>[m.id,m]));incoming.forEach(m=>map.set(m.id,m));return [...map.values()].sort((a,b)=>a.id-b.id)})
    if(!initialized.current){setNext(r.data.nextCursor);initialized.current=true}
    setReady(true);setError('')
    latest=Math.max(latest,...incoming.map(m=>m.id))
    if(incoming.length&&!document.hidden){await api.put(`/conversations/${thread}/read`,{through:latest})}
   }catch(e){if(!stopped)setError(apiMessage(e))}finally{inflight=false}
  }
  poll();const timer=setInterval(poll,4000)
  window.addEventListener('focus',poll);window.addEventListener('online',poll);document.addEventListener('visibilitychange',poll)
  return()=>{stopped=true;clearInterval(timer);window.removeEventListener('focus',poll);window.removeEventListener('online',poll);document.removeEventListener('visibilitychange',poll)}
 },[thread])
 async function older(){setBusy(true);try{const r=await api.get(`/conversations/${thread}/messages`,{params:{before:next}});if(selected.current!==thread)return;setMessages(p=>{const map=new Map([...r.data.messages,...p].map(m=>[m.id,m]));return [...map.values()].sort((a,b)=>a.id-b.id)});setNext(r.data.nextCursor)}catch(e){setError(apiMessage(e))}finally{setBusy(false)}}
 async function send(e){e.preventDefault();setBusy(true);setError('');try{await api.post(`/conversations/${thread}/messages`,{content});if(selected.current!==thread)return;setContent('');const r=await api.get(`/conversations/${thread}/messages`);if(selected.current!==thread)return;setMessages(p=>{const map=new Map([...p,...r.data.messages].map(m=>[m.id,m]));return [...map.values()].sort((a,b)=>a.id-b.id)});loadThreads()}catch(e){setError(apiMessage(e))}finally{setBusy(false)}}
 const active=threads.find(t=>t.id===thread)
 return <main className="page-content"><div className="page-heading"><div><span className="eyebrow">Good things start with a hello</span><h1>Messages</h1><p className="muted">Private conversations between community members.</p></div><button className="btn btn-soft" onClick={()=>loadThreads()}>Refresh inbox</button></div><ErrorBox error={error||inboxError}/><div className="chat-layout card-surface"><aside className="chat-inbox"><h2>Your conversations</h2>{!threads.length&&<p className="muted">Open a grower’s profile or listing and choose Message.</p>}{threads.map(t=><button key={t.id} className={`chat-contact ${thread===t.id?'active':''}`} onClick={()=>setParams({thread:t.id})}><Avatar name={t.name} photo={t.profilePhoto} /><span><strong>{t.name} {t.unread>0&&<b className="unread-badge">{t.unread}</b>}</strong><small>{t.lastMessage||'Say hello'}</small></span></button>)}{threadNext&&<button className="btn btn-soft" onClick={()=>loadThreads(threadNext)}>Older conversations</button>}</aside><section className="chat-thread">{thread?<><div className="chat-title"><Link className="chat-profile-link" to={active?`/people/${active.userId}`:'/community'}><Avatar name={active?.name} photo={active?.profilePhoto} className="avatar-small" /><strong>{active?.name||'Conversation'}</strong></Link><small>Updates every 4 seconds while open</small></div><div ref={messagePane} className="chat-messages" aria-label="Conversation messages">{next&&<button disabled={busy} className="btn btn-soft" onClick={older}>Older messages</button>}{ready&&!messages.length&&<p className="muted">A friendly hello can grow into something good.</p>}{messages.map(m=><div key={m.id} className={`message-bubble ${m.mine?'mine':''}`}><p>{m.content}</p><small>{new Date(m.createdAt).toLocaleTimeString([], {hour:'2-digit',minute:'2-digit'})}{m.mine?m.read?' · Read':' · Sent':''}</small></div>)}</div><form className="chat-compose" onSubmit={send}><textarea aria-label="Message" placeholder="Write a message…" maxLength={2000} required value={content} onChange={e=>setContent(e.target.value)}/><button className="btn btn-leaf" disabled={busy||!ready||!content.trim()}>Send</button></form></>:<div className="empty-state"><h2>A space to connect</h2><p>Choose a conversation to get started.</p></div>}</section></div></main>
}
