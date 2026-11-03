import { useCallback, useEffect, useRef, useState } from 'react'
import api, { apiMessage } from './api'

// Refresh the loaded prefix, not just the newest page: archived/changed older
// rows update too, while scrolling keeps its already-loaded history.
export function useLiveList(path, field, params = {}, interval = 8000) {
  const signature = JSON.stringify(params)
  const context = useRef(null)
  const [state, setState] = useState({ items: [], next: null, meta: {}, loading: true, busy: false, error: '' })

  const request = useCallback(async (more = false) => {
    const ctx = context.current
    if (!ctx || ctx.stopped) return
    if (ctx.busy) { if (!more) ctx.pending = true; return }
    if (more && !ctx.next) return
    ctx.busy = true
    setState(s => ({ ...s, busy: true }))
    try {
      const target = ctx.items.length
      const bottom = ctx.items.at(-1)?.id
      let cursor = more ? ctx.next : undefined
      let rows = [], meta = {}, next
      do {
        const response = await api.get(ctx.path, { params: { ...ctx.params, before: cursor } })
        if (ctx.stopped || context.current !== ctx) return
        const page = response.data[ctx.field] || []
        if (!rows.length) meta = response.data
        rows.push(...page)
        next = response.data.nextCursor || null
        if (more || !next || !page.length || (rows.length >= target && (!bottom || rows.at(-1).id <= bottom))) break
        if (next === cursor) break
        cursor = next
      } while (true)
      const combined = more ? [...ctx.items, ...rows] : rows
      ctx.items = [...new Map(combined.map(row => [row.id, row])).values()]
      ctx.next = next
      setState(s => ({ ...s, items: ctx.items, next, meta: more ? s.meta : meta, loading: false, error: '' }))
    } catch (err) {
      if (!ctx.stopped) setState(s => ({ ...s, loading: false, error: apiMessage(err, 'Updates paused. Retrying automatically…') }))
    } finally {
      ctx.busy = false
      if (!ctx.stopped) {
        setState(s => ({ ...s, busy: false }))
        if (ctx.pending) { ctx.pending = false; request() }
      }
    }
  }, [])

  useEffect(() => {
    const ctx = { path, field, params: JSON.parse(signature), items: [], next: null, busy: false, stopped: false }
    context.current = ctx
    setState({ items: [], next: null, meta: {}, loading: true, busy: false, error: '' })
    request()
    const resume = () => { if (!document.hidden) request() }
    const timer = setInterval(resume, interval)
    window.addEventListener('focus', resume)
    window.addEventListener('online', resume)
    window.addEventListener('chadkrishi:activity', resume)
    document.addEventListener('visibilitychange', resume)
    return () => {
      ctx.stopped = true
      clearInterval(timer)
      window.removeEventListener('focus', resume)
      window.removeEventListener('online', resume)
      window.removeEventListener('chadkrishi:activity', resume)
      document.removeEventListener('visibilitychange', resume)
    }
  }, [path, field, signature, interval, request])

  const loadMore = useCallback(() => request(true), [request])
  const refresh = useCallback(() => request(), [request])
  return { ...state, refresh, loadMore }
}

export function useInfiniteScroll(ref, enabled, onMore) {
  useEffect(() => {
    if (!enabled || !ref.current || typeof IntersectionObserver === 'undefined') return
    const observer = new IntersectionObserver(entries => {
      if (entries.some(entry => entry.isIntersecting)) onMore()
    }, { rootMargin: '300px' })
    observer.observe(ref.current)
    return () => observer.disconnect()
  }, [ref, enabled, onMore])
}

export function announceActivity() {
  window.dispatchEvent(new Event('chadkrishi:activity'))
}
