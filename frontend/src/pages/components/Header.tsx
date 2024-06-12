import AppBar from '@mui/material/AppBar';
import CameraIcon from '@mui/icons-material/PhotoCamera';
import Toolbar from '@mui/material/Toolbar';
import { Box, Button } from '@mui/material';
import Typography from '@mui/material/Typography';
import { Link } from 'react-router-dom';

export default function Header() {
    return (
        <header>
            <AppBar position="relative">
                <Toolbar>
                    <CameraIcon sx={{ mr: 2 }} />
                    <Typography variant="h6" color="inherit" noWrap>
                        Vendinha
                    </Typography>

                    <Box sx={{ marginLeft: 10, display: { xs: 'none', md: 'flex' } }}>
                        <Button to='/' component={Link} color='inherit'>Inicio</Button>
                        <Button to="/products" component={Link} color='inherit'>Produtos</Button>
                    </Box>
                </Toolbar>
            </AppBar>
        </header>
    )
}