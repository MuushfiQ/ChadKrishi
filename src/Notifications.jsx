import { useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import api, { apiMessage } from './api'
import { Avatar } from './Avatar'
import { useLiveList } from './live'

export default function NotificationBell() {
  const { items, next, meta, loading, busy, error, refresh, loadMore } = useLiveList('/notifications', 'notifications', {}, 5000)
  const [open, setOpen] = useState(false)
  const [actionError, setActionError] = useState('')
  const [working, setWorking] = useState(false)
  const [toast, setToast] = useState(null)
  const lastID = useRef(null), panel = useRef(null)
  const navigate = useNavigate()
  const unread = meta.unreadCount || 0

  useEffect(() => {
    if (meta.latestId === undefined) return
    const previous = lastID.current
    lastID.current = meta.latestId
    if (previous !== null && meta.latestId > previous) {
      const newest = items.find(n => n.id > previous && !n.read)
      if (newest) setToast(newest)
    }
  }, [meta.latestId, items])
  useEffect(() => {
    if (!toast) return
    const timer = setTimeout(() => setToast(null), 6500)
    return () => clearTimeout(timer)
  }, [toast])
  useEffect(() => {
    if (!open) return
    const close = event => { if (!panel.current?.contains(event.target)) setOpen(false) }
    const escape = event => { if (event.key === 'Escape') setOpen(false) }
    document.addEventListener('pointerdown', close)
    document.addEventListener('keydown', escape)
    return () => { document.removeEventListener('pointerdown', close); document.removeEventListener('keydown', escape) }
  }, [open])

  async function visit(n) {
    setWorking(true); setActionError('')
    try {
      await api.put(`/notifications/${n.id}/read`)
      setOpen(false); setToast(null)
      navigate(n.link); refresh()
    } catch (err) { setActionError(apiMessage(err)) }
    finally { setWorking(false) }
  }
  async function markAll() {
    setWorking(true); setActionError('')
    try { await api.put('/notifications/read-all', { through: meta.latestId }); refresh() }
    catch (err) { setActionError(apiMessage(err)) }
    finally { setWorking(false) }
  }

  return <div className="notification-root" ref={panel}>
    <button type="button" className="notification-bell" aria-label={`Notifications, ${unread} unread`} aria-expanded={open} aria-controls="notification-panel" onClick={() => setOpen(v => !v)}>
      <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" aria-hidden="true"><path d="M18 8a6 6 0 0 0-12 0c0 7-3 7-3 9h18c0-2-3-2-3-9ZM10 21h4"/></svg>
      {unread > 0 && <span className="notification-count">{unread > 99 ? '99+' : unread}</span>}
    </button>
    {open && <section id="notification-panel" className="notification-panel" aria-label="Notifications">
      <div className="notification-heading"><h2>Notifications</h2><button className="text-button" disabled={working || !unread} onClick={markAll}>Mark all read</button></div>
      {(error || actionError) && <p role="alert" className="form-error">{actionError || error}</p>}
      {loading ? <p role="status">Loading notifications…</p> : !items.length && <p className="muted">Messages and purchase updates will appear here.</p>}
      <div className="notification-list">{items.map(n => <button key={n.id} className={`notification-item ${n.read ? '' : 'unread'}`} disabled={working} onClick={() => visit(n)}>
        <Avatar name={n.actorName} photo={n.actorPhoto}/><span><strong>{n.actorName}</strong><span>{n.title}</span><small>{new Date(n.createdAt).toLocaleString()}</small></span>{!n.read && <span className="notification-dot" aria-label="Unread"/>}
      </button>)}</div>
      {next && <button className="btn btn-soft" disabled={busy} onClick={loadMore}>Older notifications</button>}
      <small className="muted">Updates automatically while this app is open.</small>
    </section>}
    {toast && !open && <div className="notification-toast" role="status"><button disabled={working} onClick={() => visit(toast)}><strong>{toast.actorName}</strong><span>{toast.title}</span></button><button aria-label="Dismiss notification" onClick={() => setToast(null)}>×</button>{actionError && <p role="alert">{actionError}</p>}</div>}
  </div>
}
