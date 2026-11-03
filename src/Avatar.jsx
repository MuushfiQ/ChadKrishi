import { useState } from 'react'
import api, { apiMessage } from './api'

export function Avatar({ name = 'Gardener', photo, className = '' }) {
  const [failedPhoto, setFailedPhoto] = useState(null)
  return <span className={`avatar ${className}`}>
    {photo && failedPhoto !== photo
      ? <img src={photo} alt={`${name}'s profile picture`} onError={() => setFailedPhoto(photo)} />
      : <span aria-hidden="true">{name?.charAt(0)?.toUpperCase() || '?'}</span>}
  </span>
}

export function ProfilePhotoEditor({ user, onSaved }) {
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')
  const [saved, setSaved] = useState(false)

  async function save(photoPath) {
    const response = await api.put('/profile/photo', { photoPath })
    onSaved(response.data.user)
    setSaved(true)
  }

  async function upload(event) {
    const file = event.target.files?.[0]
    event.target.value = ''
    if (!file) return
    setError(''); setSaved(false)
    if (!['image/jpeg', 'image/png'].includes(file.type) || file.size > 5 * 1024 * 1024) {
      setError('Choose a JPEG or PNG picture up to 5 MB.'); return
    }
    setBusy(true)
    try {
      const body = new FormData()
      body.append('image', file)
      const response = await api.post('/uploads', body, { headers: { 'Content-Type': undefined } })
      await save(response.data.path)
    } catch (err) {
      setError(apiMessage(err, 'Could not save your profile picture. Please try again.'))
    } finally { setBusy(false) }
  }

  async function remove() {
    setBusy(true); setError(''); setSaved(false)
    try { await save('') }
    catch (err) { setError(apiMessage(err, 'Could not remove your profile picture.')) }
    finally { setBusy(false) }
  }

  return <section className="card-surface profile-photo-editor" aria-label="Profile picture settings">
    <Avatar name={user.name} photo={user.profilePhoto} className="avatar-large" />
    <div className="profile-photo-controls">
      <h2>Your profile picture</h2>
      <p className="muted">Shown to signed-in members on your profile, posts, comments and messages.</p>
      <label className="form-label">Profile picture
        <input type="file" accept="image/jpeg,image/png" disabled={busy} onChange={upload} />
      </label>
      <small>JPEG / PNG · up to 5 MB · saves automatically</small>
      {user.profilePhoto && <button type="button" disabled={busy} className="text-button" onClick={remove}>Remove profile picture</button>}
      {busy && <p role="status">Saving profile picture…</p>}
      {saved && <p className="form-success" role="status">Profile picture updated.</p>}
      {error && <p className="form-error" role="alert">{error}</p>}
    </div>
  </section>
}
