import React from 'react'
import { createRoot } from 'react-dom/client'
import './style.css'
import Root from "./pages/root";
import Index from './pages/index';
import {
    createBrowserRouter,
    RouterProvider,
} from "react-router-dom";
import Products from './pages/products';


const router = createBrowserRouter([
    {
        path: "/",
        element: <Root />,
        children: [
            {
                index: true,
                element: <Index />,
            },
            {
                path: "products",
                element: <Products />,
            },
        ],
    },
])

const container = document.getElementById('root')
const root = createRoot(container!)

root.render(
    <React.StrictMode>
        <RouterProvider router={router} />
    </React.StrictMode>
)
