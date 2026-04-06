import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import App from './App'

window.addEventListener('unhandledrejection', (event) => {
  const err = event.reason
  if (err?.response?.status === 401 || err?.response?.status === 500) {
    event.preventDefault()
  }
})

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
