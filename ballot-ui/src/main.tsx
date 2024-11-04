import React, { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import Join from './features/join.tsx'
import Landing from './features/landing.tsx'
import Vote from './features/vote.tsx'
import { ErrorContextProvider } from './contexts/error_context.tsx'

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

function App(): React.JSX.Element {
    return (
        <StrictMode>
            <ErrorContextProvider>
                <RouterProvider router={router} />
            </ErrorContextProvider>
        </StrictMode>
    )
}

const container = document.getElementById('root')

if (container) {
    const root = createRoot(container)
    root.render(<App />)
}
