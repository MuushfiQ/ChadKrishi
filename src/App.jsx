import { CropGuide, ContainerSetup } from './CropGuide'
import { useEffect, useState } from 'react'
import { Link, Navigate, Route, Routes, useLocation, useNavigate } from 'react-router-dom'
import { OrdersPage, AdminPage, MessagesPage } from './Social'
import api, { apiMessage } from './api'
import { Avatar, ProfilePhotoEditor } from './Avatar'
import NotificationBell from './Notifications'
import { PlantsPage, CarePanel, CarePage, FeedPage, PersonPage } from './Garden'

const locations = [
  { name: 'Dhaka', latitude: 23.8103, longitude: 90.4125 },
  { name: 'Chattogram', latitude: 22.3569, longitude: 91.7832 },
  { name: 'Rajshahi', latitude: 24.3745, longitude: 88.6042 },
  { name: 'Khulna', latitude: 22.8456, longitude: 89.5403 },
  { name: 'Sylhet', latitude: 24.8949, longitude: 91.8687 },
  { name: 'Rangpur', latitude: 25.7439, longitude: 89.2752 },
]

const comingSoon = [
  { label: 'Community', icon: 'C', path:'/community' },
  { label: 'My Plants', icon: 'P', path:'/plants' },
  { label: 'Marketplace', icon: 'M', path:'/marketplace' },
 {label:'Daily care',icon:'✓',path:'/care'},
 {label:'Buy / sell history',icon:'↔',path:'/orders'},
 {label:'Messages',icon:'✉',path:'/messages'},
]

function Logo() {
  return <Link className="brand-mark" to="/dashboard"><img className="brand-logo" src="/logo.svg" alt=""/> ChadKrishi</Link>
}

function Sidebar({ user, onLogout }) {
  const location = useLocation()
  return (
    <aside className="app-sidebar">
      <div className="sidebar-brand"><Logo /><span className="brand-tagline">Grow with care</span></div>
      <nav className="sidebar-nav" aria-label="Main navigation">
        <Link className={`nav-item ${location.pathname === '/dashboard' ? 'active' : ''}`} to="/dashboard"><span className="nav-icon">D</span>Dashboard</Link>
        {comingSoon.map((item) => <Link className={`nav-item ${location.pathname === item.path ? 'active' : ''}`} key={item.path} to={item.path}><span className="nav-icon">{item.icon}</span>{item.label}</Link>)}
        <Link className={`nav-item ${location.pathname === '/profile' ? 'active' : ''}`} to="/profile"><span className="nav-icon">S</span>My profile</Link>
      </nav>
      {user?.isAdmin&&<Link className="nav-item" to="/admin">Admin archive</Link>}<div className="sidebar-footer">
        <Avatar name={user?.name} photo={user?.profilePhoto} />
        <div className="user-summary"><strong>{user?.name || 'Gardener'}</strong><span>{user?.district || 'Complete profile'}</span></div>
        <button className="logout-link" onClick={onLogout} aria-label="Sign out">↗</button>
      </div>
    </aside>
  )
}

function MobileNav({ user, onLogout }) {
  const location = useLocation()
  return (
    <nav className="mobile-nav" aria-label="Mobile navigation">
      <Link className={location.pathname === '/dashboard' ? 'active' : ''} to="/dashboard"><span>D</span>Home</Link>
      <Link className={location.pathname === '/profile' ? 'active' : ''} to="/profile"><span>S</span>Profile</Link>
      <Link to="/messages"><span>✉</span>Chat</Link>
      <Link to="/plants"><span>P</span>Plants</Link><Link to="/community"><span>C</span>Feed</Link><Link to="/marketplace"><span>M</span>Market</Link>
    </nav>
  )
}

function AppShell({ user, onLogout, children }) {
  return (
    <div className="app-shell">
      <Sidebar user={user} onLogout={onLogout} />
      <div className="app-main">
        <header className="topbar">
          <div><span className="topbar-kicker">ChadKrishi / Bangladesh</span><span className="topbar-title">A calmer way to grow</span></div>
          <NotificationBell key={user.id} /><Link className="topbar-history" to="/orders">Buy / sell history</Link><Link className="topbar-profile" to="/profile"><Avatar name={user?.name} photo={user?.profilePhoto} className="avatar-small" /><span>My profile</span></Link>
        </header>
        {children}
      </div>
      <MobileNav user={user} onLogout={onLogout} />
    </div>
  )
}

function AuthPage({ onAuthenticated }) {
  const [mode, setMode] = useState('register')
  const [form, setForm] = useState({ name: '', email: '', password: '' })
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)
  const navigate = useNavigate()

  function update(event) {
    setForm((current) => ({ ...current, [event.target.name]: event.target.value }))
    setError('')
  }

  async function submit(event) {
    event.preventDefault()
    setError('')
    if (mode === 'register' && form.name.trim().length < 2) {
      setError('Please enter your name (at least 2 characters).')
      return
    }
    if (!form.email.includes('@')) {
      setError('Please enter a valid email address.')
      return
    }
    if (form.password.length < 8) {
      setError('Password must be at least 8 characters.')
      return
    }
    setBusy(true)
    try {
      const response = await api.post(`/auth/${mode}`, mode === 'register' ? form : { email: form.email, password: form.password })
      const currentUser = await api.get('/auth/me')
      onAuthenticated(currentUser.data.user || response.data.user)
      navigate('/dashboard')
    } catch (requestError) {
      setError(apiMessage(requestError, 'Could not authenticate. Please try again.'))
    } finally {
      setBusy(false)
    }
  }

  return (
    <main className="auth-layout">
      <section className="auth-intro">
        <Logo />
        <div className="auth-intro-content"><span className="eyebrow">Your little garden companion</span><h1>Good gardens begin with small, thoughtful steps.</h1><p>Plan your rooftop or balcony garden with simple guidance made for Bangladesh.</p><div className="auth-points"><span>✦ Explainable crop ideas</span><span>✦ A private gardening profile</span><span>✦ Care guidance, at your pace</span></div></div>
      </section>
      <section className="auth-panel">
        <div className="auth-card">
          <div className="auth-switcher"><button className={mode === 'register' ? 'selected' : ''} onClick={() => setMode('register')}>Create account</button><button className={mode === 'login' ? 'selected' : ''} onClick={() => setMode('login')}>Log in</button></div>
          <h2>{mode === 'register' ? 'Welcome, gardener' : 'Welcome back'}</h2>
          <p className="muted">Your email stays private. You can update your gardening setup anytime.</p>
          <form onSubmit={submit} noValidate>
            {mode === 'register' && <label className="form-label">Full name<input className="form-control" name="name" value={form.name} onChange={update} autoComplete="name" required /></label>}
            <label className="form-label">Email address<input className="form-control" type="email" name="email" value={form.email} onChange={update} autoComplete="email" required /></label>
            <label className="form-label">Password<input className="form-control" type="password" name="password" value={form.password} onChange={update} autoComplete={mode === 'register' ? 'new-password' : 'current-password'} required /><span className="field-hint">At least 8 characters</span></label>
            {error && <div className="form-error" role="alert">{error}</div>}
            <button className="btn btn-leaf btn-lg w-100" disabled={busy}>{busy ? 'Please wait…' : mode === 'register' ? 'Create my account' : 'Log in'}</button>
          </form>
        </div>
      </section>
    </main>
  )
}

function ProfilePage({ user, onAuthenticated, onSignOut }) {
  const [form, setForm] = useState({
    name: user?.name || '', district: user?.district || '', latitude: user?.latitude || '',
    longitude: user?.longitude || '', gardeningSpace: user?.gardeningSpace || '',
    sunlightHours: user?.sunlightHours ?? '', rainExposure: user?.rainExposure ?? '', bio: user?.bio || '',
  })
  const [errors, setErrors] = useState({})
  const [saved, setSaved] = useState(false)
  const [busy, setBusy] = useState(false)

  function update(event) {
    const { name, value } = event.target
    setSaved(false)
    setErrors((current) => ({ ...current, [name]: '' }))
    if (name === 'district') {
      const location = locations.find((item) => item.name === value)
      setForm((current) => ({ ...current, district: value, latitude: location?.latitude || '', longitude: location?.longitude || '' }))
      return
    }
    setForm((current) => ({ ...current, [name]: name === 'rainExposure' ? value === 'true' : value }))
  }

  function validate() {
    const next = {}
    if (form.name.trim().length < 2) next.name = 'Enter at least 2 characters.'
    if (!form.district) next.district = 'Choose your general area.'
    if (!form.gardeningSpace) next.gardeningSpace = 'Choose rooftop or balcony.'
    if (form.sunlightHours === '' || Number(form.sunlightHours) < 0 || Number(form.sunlightHours) > 24) next.sunlightHours = 'Use a value between 0 and 24.'
    if (form.rainExposure === '') next.rainExposure = 'Please choose yes or no.'
    setErrors(next)
    return Object.keys(next).length === 0
  }

  async function submit(event) {
    event.preventDefault()
    if (!validate()) return
    setBusy(true)
    setSaved(false)
    try {
      const response = await api.put('/profile', { ...form, latitude: Number(form.latitude), longitude: Number(form.longitude), sunlightHours: Number(form.sunlightHours) })
      onAuthenticated(response.data.user)
      setSaved(true)
    } catch (requestError) {
      setErrors({ form: apiMessage(requestError, 'Could not save profile.') })
    } finally {
      setBusy(false)
    }
  }

  return (
    <main className="page-content narrow-page">
      <div className="page-heading"><div><span className="eyebrow">Your garden setup</span><h1>Tell us about your growing space</h1><p className="muted">This helps ChadKrishi make recommendations that fit your real space.</p></div><button className="btn btn-outline-danger" type="button" onClick={onSignOut}>Sign out</button></div>
      <ProfilePhotoEditor user={user} onSaved={onAuthenticated} />
      <ContainerSetup />
      <form className="card-surface profile-form" onSubmit={submit} noValidate>
        <div className="form-section-title"><span className="section-number">01</span><div><h2>About you</h2><p>Keep it simple. You can change this later.</p></div></div>
        <div className="form-grid"><label className="form-label">Name<input className={`form-control ${errors.name ? 'invalid' : ''}`} name="name" value={form.name} onChange={update} />{errors.name && <span className="inline-error">{errors.name}</span>}</label><label className="form-label">General area<select className={`form-control ${errors.district ? 'invalid' : ''}`} name="district" value={form.district} onChange={update}><option value="">Select a city</option>{locations.map((location) => <option key={location.name}>{location.name}</option>)}</select>{errors.district && <span className="inline-error">{errors.district}</span>}</label></div>
        <div className="form-section-title section-divider"><span className="section-number">02</span><div><h2>Your growing space</h2><p>We use this for seasonal, practical suggestions.</p></div></div>
        <div className="form-grid"><label className="form-label">Space type<select className={`form-control ${errors.gardeningSpace ? 'invalid' : ''}`} name="gardeningSpace" value={form.gardeningSpace} onChange={update}><option value="">Select a space</option><option value="rooftop">Rooftop</option><option value="balcony">Balcony</option></select>{errors.gardeningSpace && <span className="inline-error">{errors.gardeningSpace}</span>}</label><label className="form-label">Direct sunlight per day<input className={`form-control ${errors.sunlightHours ? 'invalid' : ''}`} type="number" min="0" max="24" step="0.5" name="sunlightHours" value={form.sunlightHours} onChange={update} /><span className="field-hint">A rough estimate is enough.</span>{errors.sunlightHours && <span className="inline-error">{errors.sunlightHours}</span>}</label></div>
        <label className="form-label radio-label">Does rain directly reach your plants?<select className={`form-control ${errors.rainExposure ? 'invalid' : ''}`} name="rainExposure" value={form.rainExposure === '' ? '' : String(form.rainExposure)} onChange={update}><option value="">Select one</option><option value="true">Yes, they are exposed</option><option value="false">No, they are covered</option></select>{errors.rainExposure && <span className="inline-error">{errors.rainExposure}</span>}</label>
        <label className="form-label">Short bio <span className="optional">(optional)</span><textarea className="form-control" rows="3" maxLength="500" name="bio" value={form.bio} onChange={update} placeholder="What do you enjoy growing?" /></label>
        {errors.form && <div className="form-error" role="alert">{errors.form}</div>}{saved && <div className="form-success" role="status">Profile saved. Your dashboard is ready.</div>}
        <div className="form-actions"><button className="btn btn-leaf btn-lg" disabled={busy}>{busy ? 'Saving…' : 'Save and view dashboard'}</button></div>
      </form>
    </main>
  )
}

function Dashboard({ user, onSessionExpired }) {
  const [data, setData] = useState(null)
  const [state, setState] = useState('loading')
  const [error, setError] = useState('')
  const navigate = useNavigate()

  async function loadRecommendations() {
    setState('loading')
    try {
      const response = await api.get('/recommendations')
      setData(response.data)
      setState('ready')
    } catch (requestError) {
      if (requestError.response?.status === 401) onSessionExpired()
      else if (requestError.response?.status === 422) setState('incomplete')
      else { setError(apiMessage(requestError, 'Could not load recommendations.')); setState('error') }
    }
  }
  useEffect(() => { loadRecommendations() }, [])

  if (state === 'loading') return <main className="page-content"><div className="loading-state"><span className="loader" />Preparing your garden dashboard…</div></main>
  if (state === 'incomplete') return <main className="page-content"><div className="empty-state"><span className="eyebrow">One small step first</span><h2>Complete your gardening profile</h2><p>We need your location, growing space, and sunlight hours before we can make useful crop suggestions.</p><Link className="btn btn-leaf" to="/profile">Complete your profile</Link></div></main>
  if (state === 'error') return <main className="page-content"><div className="empty-state"><h2>We could not load your garden</h2><p>{error}</p><button className="btn btn-leaf" onClick={loadRecommendations}>Try again</button></div></main>
  const weather = data.weather
  const weatherUnavailable = weather.status === 'unavailable'
  const weatherTitle = weather.current ? `${Math.round(weather.current.temperature)}°C · ${weather.current.description}` : 'Weather unavailable'
  return (
    <main className="page-content">
      <div className="page-heading dashboard-heading"><div><span className="eyebrow">Hello, {user.name.split(' ')[0]}</span><h1>Here is your garden today.</h1><p className="muted">Small steps, steady growth. Your setup is ready for thoughtful suggestions.</p></div><Link className="btn btn-soft" to="/profile">Edit setup <span>↗</span></Link></div>
      <section className="dashboard-grid">
        <div className="card-surface weather-card"><div className="card-topline"><span className="card-label">Current weather</span>{weather.stale && <span className="status-pill neutral">Stale cache</span>}</div>{weatherUnavailable ? <><div className="weather-main"><span className="weather-icon">—</span><div><strong>Weather unavailable</strong><span>{data.location} · live provider could not be reached</span></div></div><p className="card-note">{weather.message}</p></> : <><div className="weather-main"><span className="weather-icon">☼</span><div><strong>{weatherTitle}</strong><span>{data.location} · valid {weather.validTime}</span></div></div><div className="weather-meta"><span>Humidity {Math.round(weather.current.relativeHumidity)}%</span><span>Wind {weather.current.windSpeed} km/h</span></div><p className="card-note">{weather.provider} · fetched {weather.fetchedAt} · {weather.attribution}</p></>}</div>
        <div className="card-surface setup-card"><span className="card-label">Your setup</span><div className="setup-row"><span className="setup-icon">⌂</span><div><strong>{user.district || 'Location not set'}</strong><span>{user.gardeningSpace || 'Space not set'} · {user.sunlightHours ?? '—'} hrs direct sun</span></div></div><Link to="/profile" className="text-link">Update your setup <span>→</span></Link></div>
      </section>
      {!weatherUnavailable && weather.forecast?.length > 0 && <section className="section-block"><div className="section-heading"><div><span className="eyebrow">Seven-day forecast</span><h2>Forecast outlook</h2></div><span className="muted small-text">Separate from seasonal crop suitability</span></div><div className="forecast-grid">{weather.forecast.map((day) => <div className="forecast-day" key={day.date}><strong>{day.date}</strong><span>{Math.round(day.minTemperature)}–{Math.round(day.maxTemperature)}°C</span><span>{day.precipitationProbability == null ? 'Rain chance unavailable' : `${day.precipitationProbability}% rain chance`}</span><span>{day.precipitationSum} mm rain</span></div>)}</div></section>}
      <CarePanel compact />
      <CropGuide data={data} onReload={loadRecommendations} />
      <section className="quick-actions"><span className="card-label">Quick actions</span><div><Link to="/profile" className="quick-action"><span>✎</span>Update profile</Link><Link to="/plants" className="quick-action"><span>＋</span>Add a plant</Link><Link to="/community" className="quick-action"><span>↗</span>Share an update</Link></div></section>
    </main>
  )
}


function App() {
  const [user, setUser] = useState(null)
  const [checkingSession, setCheckingSession] = useState(true)
  const [sessionError, setSessionError] = useState('')
  const navigate = useNavigate()
  useEffect(() => {
    api.get('/auth/me')
      .then((response) => setUser(response.data.user))
      .catch((requestError) => {
        if (requestError.response?.status !== 401) setSessionError(apiMessage(requestError, 'Could not verify your session.'))
      })
      .finally(() => setCheckingSession(false))
  }, [])
  useEffect(() => {
    const interceptor = api.interceptors.response.use(r => r, error => {
      if (error.response?.status === 401 && !error.config?.url?.includes('/auth/')) {
        setUser(null); navigate('/auth', { replace:true })
      }
      return Promise.reject(error)
    })
    return () => api.interceptors.response.eject(interceptor)
  }, [navigate])
  async function logout() {
    try {
      await api.post('/auth/logout')
    } catch (e) {
      if (e.response?.status !== 401) { window.alert(apiMessage(e, 'Sign out failed. Please retry.')); return }
    }
    {
      setUser(null)
      setSessionError('')
      navigate('/auth', { replace: true })
    }
  }
  function sessionExpired() {
    setUser(null)
    navigate('/auth', { replace: true })
  }
  if (checkingSession) return <div className="loading-state full-screen"><span className="loader" />Loading your garden…</div>
  if (sessionError) return <main className="loading-state full-screen"><div className="empty-state"><h2>We could not verify your session</h2><p>{sessionError}</p><button className="btn btn-leaf" onClick={() => window.location.reload()}>Try again</button></div></main>
  return user ? <AppShell user={user} onLogout={logout}><Routes><Route path="/dashboard" element={<Dashboard user={user} onSessionExpired={sessionExpired} />} /><Route path="/plants" element={<PlantsPage />} /><Route path="/care" element={<CarePage />} /><Route path="/community" element={<FeedPage key="community" />} /><Route path="/marketplace" element={<FeedPage key="marketplace" market />} /><Route path="/orders" element={<OrdersPage />} /><Route path="/messages" element={<MessagesPage />} /><Route path="/admin" element={user.isAdmin?<AdminPage />:<Navigate to="/dashboard" replace />} /><Route path="/people/:id" element={<PersonPage />} /><Route path="/profile" element={<ProfilePage user={user} onAuthenticated={setUser} onSignOut={logout} />} /><Route path="*" element={<Navigate to="/dashboard" replace />} /></Routes></AppShell> : <Routes><Route path="/auth" element={<AuthPage onAuthenticated={setUser} />} /><Route path="*" element={<Navigate to="/auth" replace />} /></Routes>
}

export default App
