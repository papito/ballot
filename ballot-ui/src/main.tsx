import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import Index from './Index.tsx'
import Vote from './Vote.tsx';

import './index.css'

import {
    createBrowserRouter,
    RouterProvider,
} from "react-router-dom";

const router = createBrowserRouter([
    {
        path: "/",
        element: <Index />
    },
    {
        path: "/vote",
        element: <Vote />
    },
]);
createRoot(document.getElementById('root')!).render(
  <StrictMode>
      <RouterProvider router={router} />
  </StrictMode>,
)
