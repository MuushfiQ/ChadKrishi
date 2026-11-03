import axios from 'axios'

const apiOrigin = (import.meta.env.VITE_API_URL || '').replace(/\/$/, '')

const api = axios.create({
  baseURL: `${apiOrigin}/api/v1`,
  withCredentials: true,
  headers: { 'Content-Type': 'application/json' },
})

export function apiMessage(error, fallback = 'Something went wrong. Please try again.') {
  return error.response?.data?.error?.message || fallback
}

export default api
