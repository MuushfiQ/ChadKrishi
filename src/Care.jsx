import { useEffect, useRef, useState } from 'react'
import { Link } from 'react-router-dom'
import api, { apiMessage } from './api'

const today = () => new Intl.DateTimeFormat('en-CA', { timeZone: 'Asia/Dhaka', year: 'numeric', month: '2-digit', day: '2-digit' }).format(new Date())

export function CarePanel({ compact = false }) {
  const [date, setDate] = useState(today())
  const [data, setData] = useState(null), [plants, setPlants] = useState([])
  const [error, setError] = useState(''), [busy, setBusy] = useState(false), [loading, setLoading] = useState(true)
  const [open, setOpen] = useState(false), [form, setForm] = useState({ plantId: '', title: '', reason: '', date: today() })
  const requestID = useRef(0)

  async function load() {
    const id = ++requestID.current
    setLoading(true); setError('')
    try { const r = await api.get('/tasks', { params: { date } }); if (requestID.current === id) setData(r.data) }
    catch (e) { if (requestID.current === id) setError(apiMessage(e)) }
    finally { if (requestID.current === id) setLoading(false) }
  }
  useEffect(() => { setData(null); load(); return () => { requestID.current++ } }, [date])
  useEffect(() => {
    if (compact) return
    let stopped = false
    api.get('/plants').then(r => { if (!stopped) setPlants((r.data.plants || []).filter(p => !p.archived)) }).catch(e => { if (!stopped) setError(apiMessage(e)) })
    return () => { stopped = true }
  }, [compact])

  async function toggle(task) {
    setBusy(true); setError('')
    try {
      await api.put(`/tasks/${task.id}`, { completed: !task.completed })
      setData(d => ({ ...d, tasks: d.tasks.map(t => t.id === task.id ? { ...t, completed: !t.completed } : t) }))
    } catch (e) { setError(apiMessage(e)) }
    finally { setBusy(false) }
  }
  async function create(event) {
    event.preventDefault(); setBusy(true); setError('')
    try {
      await api.post('/tasks', { ...form, plantId: Number(form.plantId) })
      setOpen(false); setForm(f => ({ ...f, title: '', reason: '' }))
      if (form.date !== date) setDate(form.date); else await load()
    } catch (e) { setError(apiMessage(e)) }
    finally { setBusy(false) }
  }
  async function remove(task) {
    if (!window.confirm('Delete this custom task?')) return
    setBusy(true); setError('')
    try { await api.delete(`/tasks/${task.id}`); setData(d => ({ ...d, tasks: d.tasks.filter(t => t.id !== task.id) })) }
    catch (e) { setError(apiMessage(e)) }
    finally { setBusy(false) }
  }
  function addFor(plantId = '') { setForm({ plantId: String(plantId), title: '', reason: '', date }); setOpen(true) }
  const groups = new Map()
  for (const task of data?.tasks || []) {
    const key = task.plantId ?? task.plantName
    if (!groups.has(key)) groups.set(key, { id: task.plantId, name: task.plantName, crop: task.cropName, tasks: [] })
    groups.get(key).tasks.push(task)
  }

  return <section className="section-block">
    <div className="section-heading"><div><span className="eyebrow">Small steps, daily</span><h2>{compact ? 'Today’s care' : 'Care checklist'}</h2></div>
      {compact ? <Link className="text-link" to="/care">Manage tasks →</Link> : <div className="care-tools"><label>Checklist date<input aria-label="Checklist date" type="date" value={date} disabled={busy} onChange={e => setDate(e.target.value)} /></label><button className="btn btn-leaf" onClick={() => addFor()}>＋ Add custom task</button></div>}
    </div>
    {error && <div className="form-error" role="alert">{error}</div>}
    {open && <form className="card-surface garden-form" onSubmit={create}>
      <div className="section-heading"><h3>Your custom task</h3><button type="button" className="btn btn-soft" disabled={busy} onClick={() => setOpen(false)}>Cancel</button></div>
      <div className="form-grid"><label>Plant<select required value={form.plantId} onChange={e => setForm({ ...form, plantId: e.target.value })}><option value="">Choose your plant</option>{plants.map(p => <option key={p.id} value={p.id}>{p.nickname} · {p.cropName} (#{p.id})</option>)}</select></label><label>Task date<input required type="date" value={form.date} onChange={e => setForm({ ...form, date: e.target.value })}/></label></div>
      <label>Task title<input required maxLength={160} value={form.title} onChange={e => setForm({ ...form, title: e.target.value })} placeholder="Check the support for my tomato plant"/></label>
      <label>Notes (optional)<textarea maxLength={300} value={form.reason} onChange={e => setForm({ ...form, reason: e.target.value })}/></label>
      <p className="muted">A one-time task for this plant and date. Your own note, saved alongside the daily checklist.</p>
      {!plants.length && <p>Add an active plant in <Link to="/plants">My Plants</Link> first.</p>}
      <button className="btn btn-leaf" disabled={busy || !plants.length}>{busy ? 'Saving…' : 'Save custom task'}</button>
    </form>}
    {loading ? <p role="status">Preparing checklist…</p> : data && <>
      <p className="muted small-text">{data.guidance}</p>
      {date > today() && <p className="market-note">Future dates show your scheduled tasks. Automatic daily guidance is generated on that day.</p>}
      {!data.tasks.length ? <div className="card-surface empty-state compact"><h3>Nothing on your checklist yet</h3><p>Choose a date or add a task for one of your plants.</p><Link to="/plants">My plants →</Link></div> : <>
        <p className="eyebrow">{data.tasks.filter(t => t.completed).length} / {data.tasks.length} complete</p>
        <div className="plant-care-groups">{[...groups.values()].map(group => <section className="card-surface plant-care-group" key={group.id ?? group.name} aria-label={`${group.name} checklist`}>
          <header className="plant-care-heading"><div><span className="eyebrow">{group.crop || 'Your plant'}{group.id ? ` · #${group.id}` : ''}</span><h3>{group.name}</h3><small>{group.tasks.filter(t => t.completed).length} of {group.tasks.length} tasks done</small></div>{!compact && plants.some(p => p.id === group.id) && <button className="btn btn-soft" onClick={() => addFor(group.id)}>＋ Add task</button>}</header>
          <div className="task-list">{group.tasks.map(task => <div className="care-task-item" key={task.id}><label className={`care-row ${task.completed ? 'completed' : ''}`}><input type="checkbox" checked={task.completed} disabled={busy} onChange={() => toggle(task)}/><span><strong>{task.title}</strong>{task.custom && <small className="custom-task-label">Your custom task</small>}<span>{task.reason}</span></span></label>{task.custom && !compact && <button className="text-button danger-text" disabled={busy} aria-label={`Delete task: ${task.title}`} onClick={() => remove(task)}>Delete</button>}</div>)}</div>
        </section>)}</div>
      </>}
    </>}
  </section>
}
