import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import Join from './features/join.tsx'
import Landing from './features/landing.tsx'
import Vote from './features/vote.tsx'

import './core.css'

import { createBrowserRouter, RouterProvider } from 'react-router-dom'

const router = createBrowserRouter([
    {
        path: '/',
        element: <Landing />,
    },
    {
        path: '/vote/s/:sessionId/u/:userId',
        element: <Vote />,
    },
    {
        path: '/vote/s/:sessionId',
        element: <Join />,
    },
])

createRoot(document.getElementById('root')!).render(
    <StrictMode>
        <RouterProvider router={router} />
    </StrictMode>
)
