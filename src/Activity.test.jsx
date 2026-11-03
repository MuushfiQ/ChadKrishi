// @vitest-environment jsdom
import React, { useState } from 'react'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { act, cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import '@testing-library/jest-dom/vitest'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import api from './api'
import NotificationBell from './Notifications'
import { CarePanel } from './Care'
import { FeedPage } from './Garden'
import { MessagesPage, OrdersPage } from './Social'
import { useLiveList } from './live'

vi.mock('./api', () => ({ default: { get: vi.fn(), put: vi.fn(), post: vi.fn(), delete: vi.fn() }, apiMessage: e => e.message }))
beforeEach(() => { vi.resetAllMocks(); api.put.mockResolvedValue({}); Object.defineProperty(document, 'hidden', { configurable: true, value: false }) })
afterEach(() => { cleanup(); vi.useRealTimers(); vi.unstubAllGlobals() })
const note = { id: 4, kind: 'message', title: 'Sent you a new message', link: '/messages?thread=2', actorName: 'Rahim', read: false, createdAt: '2026-10-22T00:00:00Z' }

it('notification bell shows persisted unread activity, marks the selected item and opens its chat', async () => {
  api.get.mockResolvedValue({ data: { notifications: [note], unreadCount: 1, latestId: 4 } })
  render(<MemoryRouter><NotificationBell/><Routes><Route path="/" element={null}/><Route path="/messages" element={<p>Selected chat</p>}/></Routes></MemoryRouter>)
  fireEvent.click(await screen.findByRole('button', { name: 'Notifications, 1 unread' }))
  fireEvent.click(screen.getByRole('button', { name: /Rahim Sent you a new message/ }))
  expect(await screen.findByText('Selected chat')).toBeInTheDocument()
  expect(api.put).toHaveBeenCalledWith('/notifications/4/read')
})

it('new notifications arrive by timer and mark-all uses the last known ID', async () => {
  vi.useFakeTimers()
  let fresh = false
  api.get.mockImplementation(async () => ({ data: { notifications: fresh ? [note] : [], unreadCount: fresh ? 1 : 0, latestId: fresh ? 4 : 0 } }))
  render(<MemoryRouter><NotificationBell/></MemoryRouter>)
  await act(async () => {})
  fresh = true
  await act(async () => { await vi.advanceTimersByTimeAsync(5000) })
  expect(screen.getByRole('button', { name: 'Notifications, 1 unread' })).toBeInTheDocument()
  expect(screen.getByRole('status')).toHaveTextContent('Rahim')
  fireEvent.click(screen.getByRole('button', { name: 'Notifications, 1 unread' }))
  await act(async () => { fireEvent.click(screen.getByRole('button', { name: 'Mark all read' })) })
  expect(api.put).toHaveBeenCalledWith('/notifications/read-all', { through: 4 })
})

it('care tasks group by plant ID even when two plants have the same name', async () => {
  api.get.mockResolvedValue({ data: { tasks: [
    { id: 1, plantId: 11, plantName: 'Tomato', title: 'Check pot one', completed: false },
    { id: 2, plantId: 12, plantName: 'Tomato', title: 'Check pot two', completed: true },
    { id: 3, plantId: 11, plantName: 'Tomato', title: 'Check support', completed: false },
  ] } })
  render(<MemoryRouter><CarePanel compact/></MemoryRouter>)
  const groups = await screen.findAllByRole('region', { name: 'Tomato checklist' })
  expect(groups).toHaveLength(2)
  expect(within(groups[0]).getAllByRole('checkbox')).toHaveLength(2)
  expect(within(groups[1]).getByRole('checkbox')).toBeChecked()
  expect(within(groups[0]).queryByText('Check pot two')).not.toBeInTheDocument()
})

it('creates a custom task for the chosen plant and displays it in that plant checklist', async () => {
  let saved = false
  api.get.mockImplementation(async path => ({ data: path === '/plants' ? { plants: [{ id: 12, nickname: 'Balcony tomato', cropName: 'Tomato' }] } : { tasks: saved ? [{ id: 7, plantId: 12, plantName: 'Balcony tomato', title: 'Tie support', reason: 'Use soft string', custom: true, completed: false }] : [] } }))
  api.post.mockImplementation(async () => { saved = true; return { data: { id: 7 } } })
  render(<MemoryRouter><CarePanel/></MemoryRouter>)
  fireEvent.click(await screen.findByRole('button', { name: '＋ Add custom task' }))
  await screen.findByRole('option', { name: /Balcony tomato/ })
  fireEvent.change(screen.getByLabelText('Plant'), { target: { value: '12' } })
  fireEvent.change(screen.getByLabelText('Task title'), { target: { value: 'Tie support' } })
  fireEvent.change(screen.getByLabelText('Notes (optional)'), { target: { value: 'Use soft string' } })
  fireEvent.click(screen.getByRole('button', { name: 'Save custom task' }))
  expect(await screen.findByText('Tie support')).toBeInTheDocument()
  expect(api.post).toHaveBeenCalledWith('/tasks', expect.objectContaining({ plantId: 12, title: 'Tie support', reason: 'Use soft string' }))
})

function LiveHarness() {
  const [filter, setFilter] = useState('all')
  const { items, next, error, loadMore } = useLiveList('/posts', 'posts', { type: filter }, 1000)
  return <><button onClick={() => setFilter('sale')}>Filter</button>{items.map(p => <p key={p.id}>{p.title}</p>)}{next && <button onClick={loadMore}>More</button>}{error && <p role="alert">{error}</p>}</>
}

it('live refresh preserves older loaded pages and deduplicates overlapping cursor results', async () => {
  vi.useFakeTimers()
  let updated = false
  api.get.mockImplementation(async (path, { params }) => ({ data: params.before ? { posts: [{ id: 3, title: 'Third' }, { id: 2, title: updated ? 'Updated older row' : 'Second' }], nextCursor: null } : { posts: [{ id: 4, title: 'Fourth' }, { id: 3, title: 'Third' }], nextCursor: 3 } }))
  render(<LiveHarness/>); await act(async () => {})
  await act(async () => { fireEvent.click(screen.getByText('More')) })
  expect(screen.getAllByText('Third')).toHaveLength(1)
  updated = true
  await act(async () => { await vi.advanceTimersByTimeAsync(1000) })
  expect(screen.getByText('Updated older row')).toBeInTheDocument()
  expect(screen.getByText('Fourth')).toBeInTheDocument()
})

it('a response from an old filter cannot overwrite the current feed', async () => {
  let resolveOld
  api.get.mockImplementation((path, { params }) => params.type === 'all' ? new Promise(resolve => { resolveOld = resolve }) : Promise.resolve({ data: { posts: [{ id: 2, title: 'Current sale' }] } }))
  render(<LiveHarness/>)
  fireEvent.click(screen.getByText('Filter'))
  await screen.findByText('Current sale')
  await act(async () => { resolveOld({ data: { posts: [{ id: 1, title: 'Stale post' }] } }) })
  expect(screen.queryByText('Stale post')).not.toBeInTheDocument()
})

it('scrolling to the feed sentinel loads the next page automatically', async () => {
  let intersect
  vi.stubGlobal('IntersectionObserver', class { constructor(callback) { intersect = callback } observe() {} disconnect() {} })
  const post = id => ({ id, authorId: 1, authorName: 'Grower', title: `Post ${id}`, content: 'Garden', type: 'update', createdAt: '2026-10-22', reactions: {}, commentCount: 0 })
  api.get.mockImplementation(async (path, { params }) => ({ data: params.before ? { posts: [post(1)], nextCursor: null } : { posts: [post(2)], nextCursor: 2 } }))
  render(<MemoryRouter><FeedPage/></MemoryRouter>)
  await screen.findByText('Post 2')
  await waitFor(() => expect(intersect).toBeTypeOf('function'))
  await act(async () => { intersect([{ isIntersecting: true }]) })
  expect(await screen.findByText('Post 1')).toBeInTheDocument()
  expect(screen.getByText('Post 2')).toBeInTheDocument()
})

it('connection failure preserves visible data and the next poll recovers', async () => {
  vi.useFakeTimers()
  api.get.mockResolvedValueOnce({ data: { posts: [{ id: 1, title: 'Saved post' }] } }).mockRejectedValueOnce(new Error('Offline')).mockResolvedValue({ data: { posts: [{ id: 1, title: 'Recovered post' }] } })
  render(<LiveHarness/>); await act(async () => {})
  await act(async () => { await vi.advanceTimersByTimeAsync(1000) })
  expect(screen.getByText('Saved post')).toBeInTheDocument()
  expect(screen.getByRole('alert')).toHaveTextContent('Offline')
  await act(async () => { await vi.advanceTimersByTimeAsync(1000) })
  expect(screen.getByText('Recovered post')).toBeInTheDocument()
  expect(screen.queryByRole('alert')).not.toBeInTheDocument()
})


it('incoming messages and new conversations update without reloading the page or clearing a draft', async () => {
  vi.useFakeTimers()
  let fresh = false
  api.get.mockImplementation(async path => ({ data: path === '/conversations'
    ? { conversations: [{ id: 2, userId: 3, name: 'Rahim', lastMessage: fresh ? 'New hello' : 'Hello', unread: 0 }, ...(fresh ? [{ id: 5, userId: 4, name: 'New member', lastMessage: 'Hi', unread: 1 }] : [])] }
    : { messages: [{ id: 1, content: 'Hello', mine: false, createdAt: '2026-10-22T00:00:00Z' }, ...(fresh ? [{ id: 2, content: 'New hello', mine: false, createdAt: '2026-10-22T00:01:00Z' }] : [])].sort((a,b)=>b.id-a.id) } }))
  render(<MemoryRouter initialEntries={['/messages?thread=2']}><MessagesPage/></MemoryRouter>)
  await act(async () => {})
  fireEvent.change(screen.getByLabelText('Message'), { target: { value: 'Unsent draft' } })
  fresh = true
  await act(async () => { await vi.advanceTimersByTimeAsync(4000) })
  expect(screen.getByText('New hello', { selector: 'p' })).toBeInTheDocument()
  expect(screen.getByRole('button', { name: /New member/ })).toBeInTheDocument()
  expect(screen.getByLabelText('Message')).toHaveValue('Unsent draft')
})

it('seller payment controls update when the buyer marks Paid in another session', async () => {
  vi.useFakeTimers()
  let paid = false
  api.get.mockImplementation(async () => ({ data: { orders: [{ id: 1, title: 'Seedling', amountBdt: 100, quantity: 1, buyerId: 2, buyerName: 'Buyer', sellerId: 1, sellerName: 'Seller', isBuyer: false, status: paid ? 'buyer_paid' : 'waiting_payment', createdAt: '2026-10-22' }] } }))
  render(<MemoryRouter><OrdersPage/></MemoryRouter>); await act(async () => {})
  expect(screen.getByRole('button', { name: 'Yes, payment received' })).toBeDisabled()
  paid = true
  await act(async () => { await vi.advanceTimersByTimeAsync(5000) })
  expect(screen.getByRole('button', { name: 'Yes, payment received' })).toBeEnabled()
})
