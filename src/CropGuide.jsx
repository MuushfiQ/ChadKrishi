import React, { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import api, { apiMessage } from './api'

const months = ['Jan','Feb','Mar','Apr','May','Jun','Jul','Aug','Sep','Oct','Nov','Dec']
const names = { candidate:'Planning candidate', conditional:'Check conditions', needs_setup:'Add pot details', needs_evidence:'Evidence incomplete', out_of_season:'Outside planting window', not_fit:'Setup does not fit' }
const icons = { brinjal:'🍆', tomato:'🍅', chili:'🌶️', okra:'🌱', red_amaranth:'🥬', malabar_spinach:'🌿', spinach:'🥬', coriander:'🌿', radish:'🌱', cucumber:'🥒' }
const blank = { depthCm:null, volumeL:null, drainage:null, support:null, method:'seed' }
function SourceLink({ source, id }) { return source?.url ? <a href={source.url} target="_blank" rel="noreferrer">{source.title||id}</a> : <span>{source?.title||id}</span> }
const monthText = list => list?.length ? list.map(m=>months[m-1]).join(' · ') : 'Unknown / not established'

export function ContainerSetup({ initial, onSaved }) {
 const [form,setForm]=useState(initial||blank), [busy,setBusy]=useState(false), [error,setError]=useState(''), [saved,setSaved]=useState(false)
 useEffect(()=>{let active=true;if(!initial)api.get('/garden-setup').then(r=>{if(active)setForm(r.data)}).catch(e=>{if(active)setError(apiMessage(e))});return()=>{active=false}},[])
 function update(key,value){setForm(f=>({...f,[key]:value}));setSaved(false)}
 async function submit(e){e.preventDefault();setBusy(true);setError('');try{const r=await api.put('/garden-setup',{...form,depthCm:form.depthCm===''?null:form.depthCm,volumeL:form.volumeL===''?null:form.volumeL});setForm(r.data);setSaved(true);onSaved?.(r.data)}catch(e){setError(apiMessage(e))}finally{setBusy(false)}}
 return <form className="card-surface pot-setup" onSubmit={submit}>
  <div><span className="eyebrow">Your next planting</span><h3>Check one planned container</h3><p className="muted">Enter the pot you plan to use. These are your observations—not sensor readings. Leave unknown measurements blank.</p></div>
  <div className="pot-fields">
   <label>Pot depth (cm)<input className="form-control" type="number" step="0.01" min="0.01" max="1000" value={form.depthCm??''} onChange={e=>update('depthCm',e.target.value===''?null:Number(e.target.value))}/></label>
   <label>Pot volume (litres)<input className="form-control" type="number" step="0.001" min="0.001" max="1000" value={form.volumeL??''} onChange={e=>update('volumeL',e.target.value===''?null:Number(e.target.value))}/></label>
   <label>Drainage holes<select className="form-control" value={form.drainage==null?'':String(form.drainage)} onChange={e=>update('drainage',e.target.value===''?null:e.target.value==='true')}><option value="">Not checked</option><option value="true">Yes</option><option value="false">No</option></select></label>
   <label>Stakes / trellis available<select className="form-control" value={form.support==null?'':String(form.support)} onChange={e=>update('support',e.target.value===''?null:e.target.value==='true')}><option value="">Not checked</option><option value="true">Yes</option><option value="false">No</option></select></label>
   <label>Starting with<select className="form-control" value={form.method} onChange={e=>update('method',e.target.value)}><option value="seed">Seeds</option><option value="seedling">A seedling / transplant</option></select></label>
  </div><p className="small-text muted">Depth and volume are separate measurements. This setup does not change existing plants.</p>
  {error&&<p role="alert" className="form-error">{error}</p>}{saved&&<p role="status">Container setup saved.</p>}
  <button className="btn btn-leaf" disabled={busy}>{busy?'Saving…':'Save container setup'}</button>
 </form>
}
export function ResearchCropCard({ crop, sources={} }) {
 const [imageFailed,setImageFailed]=useState(false)
 useEffect(()=>setImageFailed(false),[crop.imagePath])
 const localImage=typeof crop.imagePath==='string' && /^\/crops\/[a-zA-Z0-9_-]+\.(jpg|jpeg|png|webp)$/i.test(crop.imagePath)
 return <article className={`research-crop crop-${crop.slug}`}>
  <div className="research-crop-top"><span className="crop-symbol" aria-hidden="true">{icons[crop.slug]||'🌱'}</span><span className={`research-status status-${crop.status}`}>{names[crop.status]||'Research entry'}</span></div>
  {localImage&&!imageFailed&&<img className="research-crop-photo" src={crop.imagePath} alt={crop.name} loading="lazy" onError={()=>setImageFailed(true)}/>}
  <div className="research-crop-body"><span className="eyebrow">{crop.name}</span><h3 lang="bn">{crop.bengaliName}</h3><p className="botanical-name">{crop.scientificName}</p>
   <div className="research-metrics"><div><span>Direct sun</span><strong>{crop.minSunHours==null?'Unknown':`${crop.minSunHours}+ h/day`}</strong></div><div><span>Pot depth</span><strong>{crop.minDepthCm==null?'Unknown':`${crop.minDepthCm} cm`}</strong></div><div><span>Pot volume</span><strong>{crop.minVolumeL==null?'Unknown':`${crop.minVolumeL} L`}</strong></div></div>
   <div className="crop-explanation">{crop.reasons?.map((r,i)=><p key={i}>{r}</p>)}</div>
   <p className="small-text"><strong>Seed sowing:</strong> {monthText(crop.seedMonths)}</p><p className="small-text"><strong>Transplanting:</strong> {monthText(crop.transplantMonths)}</p>
   <details className="crop-evidence"><summary>Growing notes & evidence</summary>
    <h4>Variety & season</h4><p>{crop.varietyNote}</p><p>{crop.seasonNote}</p>
    <h4>Before planting</h4>{crop.conditions?.map((v,i)=><p key={i}>{v}</p>)}<p>{crop.containerNote}</p>
    <h4>Care observations</h4><p>{crop.careNote}</p><p className="muted small-text">General observation reminders, not measured plant health or a treatment prescription. Scientific names, harvest estimates and support notes marked report-only still need independent review.</p>
    {crop.reportHarvest&&<><h4>Harvest estimate from the supplied report</h4><p>{crop.reportHarvest.min}–{crop.reportHarvest.max} days from {crop.reportHarvest.from}. {crop.reportHarvest.note}</p><p className="muted small-text">{crop.reportHarvest.status}</p></>}
    <h4>Verification</h4><p>{crop.reviewNote}</p><p className="small-text">Report checked: {crop.reportCheckedAt} · App review: {crop.checkedAt}. Temperature thresholds not validated for this cultivar.</p>
    <h4>Field sources</h4><ul>{Object.entries(crop.fieldSources||{}).filter(([,ids])=>ids.length).map(([field,ids])=><li key={field}><strong>{field}: </strong>{ids.map((id,i)=><React.Fragment key={id}>{i>0?' · ':''}<SourceLink source={sources[id]} id={id}/></React.Fragment>)}</li>)}</ul>
    <h4>Source review notes</h4>{crop.sourceIds?.map(id=><p key={id} className="small-text"><SourceLink source={sources[id]} id={id}/> — {sources[id]?.section} ({sources[id]?.status})</p>)}
   </details>
   <Link className="btn btn-soft" to={`/plants?crop=${crop.id}`}>Track a plant of this crop →</Link><p className="small-text muted">Tracking a plant does not certify its suitability.</p>
  </div>
 </article>
}
export function CropGuide({ data, onReload }) {
 const [query,setQuery]=useState(''),[filter,setFilter]=useState('all')
 const all=data.cropCatalog||[]
 const visible=all.filter(c=>(filter==='all'||(filter==='shortlist'?['candidate','conditional'].includes(c.status):c.status===filter))&&`${c.name} ${c.bengaliName}`.toLowerCase().includes(query.toLowerCase()))
 return <section className="section-block research-guide">
  <div className="research-heading"><div><span className="eyebrow">Know what you grow</span><h2>Find your next crop</h2><p>১০টি ফসলের তথ্য, পরিচর্যা ও planting guidance এক জায়গায়।</p></div><div className="research-count"><strong>{all.length}</strong><span>research entries</span></div></div>
  <p className="card-surface research-disclaimer">{data.guidance} Location changes your weather forecast; this catalog does not invent district-specific planting rules. Current date: {data.date}.</p>
  <ContainerSetup initial={data.setup} onSaved={onReload}/>
  <div className="research-toolbar"><label>Search crops<input className="form-control" value={query} onChange={e=>setQuery(e.target.value)} placeholder="বেগুন, tomato, coriander…"/></label><label>Show<select className="form-control" value={filter} onChange={e=>setFilter(e.target.value)}><option value="all">All crops</option><option value="shortlist">Planning shortlist</option><option value="needs_setup">Need pot details</option><option value="needs_evidence">Evidence incomplete</option><option value="out_of_season">Outside planting window</option><option value="not_fit">Setup does not fit</option></select></label><span>{visible.length} of {all.length} crops</span></div>
  {!visible.length?<p className="card-surface empty-state">No crops in this view. Try All crops or review your setup.</p>:<div className="research-grid">{visible.map(c=><ResearchCropCard key={c.slug} crop={c} sources={data.sources}/>)}</div>}
 </section>
}
