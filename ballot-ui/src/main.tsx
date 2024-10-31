import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import Join from './views/join.tsx'
import Landing from './views/landing.tsx'
import Vote from './views/vote.tsx'

import './core.css'

import { createBrowserRouter, Navigate, RouterProvider } from 'react-router-dom'

const router = createBrowserRouter([
    {
        // This is only needed for the React dev server as the Go server will go directly to /p
        path: '/',
        element: <Navigate to="/p" replace />,
    },
    {
        path: '/p',
        element: <Landing />,
    },
    {
        path: '/p/vote/s/:sessionId/u/:userId',
        element: <Vote />,
    },
    {
        path: '/p/vote/s/:sessionId',
        element: <Join />,
    },
])

createRoot(document.getElementById('root')!).render(
    <StrictMode>
        <RouterProvider router={router} />
    </StrictMode>
)
