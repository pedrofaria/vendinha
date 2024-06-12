import { ThemeProvider } from "@mui/material/styles"
import { Outlet } from "react-router-dom"
import React from 'react'
import theme from '../theme/theme'
import CssBaseline from '@mui/material/CssBaseline'
import Header from './components/Header'

function Root() {
    return (
        <ThemeProvider theme={theme}>
            <CssBaseline />
            <React.StrictMode>
                <Header />

                <main>
                    <Outlet />
                </main>
            </React.StrictMode>
        </ThemeProvider>
    )
}

export default Root
