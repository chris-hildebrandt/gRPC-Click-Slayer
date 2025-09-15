import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import GrpcApp from './GrpcApp.jsx'

createRoot(document.getElementById('root')).render(
  <StrictMode>
    <GrpcApp />
  </StrictMode>,
)
