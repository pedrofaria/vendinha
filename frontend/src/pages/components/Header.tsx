import AppBar from '@mui/material/AppBar';
import ShoppingBagIcon from '@mui/icons-material/ShoppingBag';
import Toolbar from '@mui/material/Toolbar';
import { Box, Button } from '@mui/material';
import Typography from '@mui/material/Typography';
import { Link } from 'react-router-dom';
import { PrintReport } from '../../../wailsjs/go/report/Service';

export default function Header() {
    const printReport = () => {
        PrintReport().catch(err => console.log(err))
    }

    return (
        <header>
            <AppBar position="relative">
                <Toolbar>
                    <ShoppingBagIcon sx={{ mr: 2 }} />
                    <Typography variant="h6" color="inherit" noWrap>
                        Vendinha
                    </Typography>

                    <Box sx={{ marginLeft: 10, display: { xs: 'none', md: 'flex' } }}>
                        <Button to='/' component={Link} color='inherit'>Inicio</Button>
                        <Button to="/products" component={Link} color='inherit'>Produtos</Button>
                        <Button color='inherit' onClick={() => printReport()}>Relatorio</Button>
                    </Box>
                </Toolbar>
            </AppBar>
        </header>
    )
}