import { useEffect, useState } from 'react'
import Login from './Login'

export default function App() {
  const [authed, setAuthed] = useState<boolean | null>(null)

  useEffect(() => {
    checkAuth()
  }, [])

  function checkAuth() {
    fetch('/api/health', { credentials: 'same-origin' })
      .then(res => {
        setAuthed(res.ok)
      })
      .catch(() => setAuthed(false))
  }

  if (authed === null) return null

  if (!authed) {
    return <Login onLogin={() => setAuthed(true)} />
  }

  return (
    <div>
      <h1>Luna DNS</h1>
    </div>
  )
}
