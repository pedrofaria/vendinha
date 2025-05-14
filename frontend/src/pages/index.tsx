import React, { useEffect, useState } from 'react';
import Container from '@mui/material/Container';
import Box from '@mui/material/Box';
import { List } from '../../wailsjs/go/product/Service';
import { PlacePurchase } from '../../wailsjs/go/purchase/Service';
import { model } from '../../wailsjs/go/models';
import { Button, Grid, Paper, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Typography } from '@mui/material';
import IconButton from '@mui/material/IconButton';
import DeleteIcon from '@mui/icons-material/Delete';
import AddIcon from '@mui/icons-material/Add';
import RemoveIcon from '@mui/icons-material/Remove';
import Format from '../libs/format';
import useEventListener from '@use-it/event-listener';

const Index = () => {
    const [products, setProducts] = useState(Array<model.Product>);
    const [purchase, setPurchase] = useState<model.Purchase>();

    const handleKeyDown = (ev: any) => {
        if (purchase == null) {
            return;
        }

        console.log(ev);

        if (ev.key === 'n') {
            newPurchase();
            return;
        }

        if (ev.key === 'p') {
            placePurchase();
            return;
        }

        const i = products.findIndex((row) => row.Key == ev.key);
        if (i < 0) {
            console.log("not found", i);
            return;
        }

        addItem(products[i])
    };

    useEventListener('keyup', handleKeyDown);

    const list = () => {
        List().then(data => {
            setProducts(data);
        })
    }

    const sortFunc = (a: model.Product, b: model.Product): number => {
        const aLbl = `${a.Order} - ` + ((a.Key != "") ? `${a.Key} - ` : "") + `${a.Code} (${Format(a.Price)})}`
        const bLbl = `${b.Order} - ` + ((b.Key != "") ? `${b.Key} - ` : "") + `${b.Code} (${Format(b.Price)})}`

        return aLbl > bLbl ? 1 : -1
    }

    const newPurchase = () => {
        const purchase = new model.Purchase({ Products: [] });
        console.log(purchase)
        setPurchase(purchase);
    }

    const addItem = (prod: model.Product) => {
        if (purchase == null) {
            return;
        }

        const p = new model.Purchase({ ...purchase });

        const idx = purchase.Products.findIndex((item) => item.ProductID == prod.ID);
        if (idx == -1) {
            const item = new model.PurchaseItem({
                Quantity: 1,
                ProductID: prod.ID,
                ProductCode: prod.Code,
                ProductPrice: prod.Price
            });

            p.Products.push(item);
        } else {
            p.Products[idx].Quantity++;
        }

        setPurchase(p);
    }

    const incItem = (idx: Number) => {
        const p = new model.Purchase({ ...purchase });
        const i = parseInt(idx.toString());
        p.Products[i].Quantity++;
        setPurchase(p);
    }

    const decItem = (idx: Number) => {
        const p = new model.Purchase({ ...purchase });
        const i = parseInt(idx.toString());

        p.Products[i].Quantity--;
        if (p.Products[i].Quantity <= 0) {
            p.Products.splice(i, 1);
        }
        setPurchase(p);
    }

    const delItem = (idx: Number) => {
        const p = new model.Purchase({ ...purchase });
        const i = parseInt(idx.toString());
        p.Products.splice(i, 1);
        setPurchase(p);
    }

    const calcTotal = (): number => {
        if (purchase == null) {
            return 0;
        }

        let total = 0;

        purchase.Products.map((item) => total += item.Quantity * item.ProductPrice)

        return total;
    }

    const placePurchase = () => {
        if (purchase == null) {
            return;
        }

        if (purchase.Products.length == 0) {
            window.alert("Nada para imprimir");
            return;
        }

        PlacePurchase(purchase)
            .then(() => {
                // newPurchase();
            })
            .catch((err) => {
                window.alert(err);
            });
    }

    useEffect(() => {
        list();
        newPurchase();
    }, []);

    return (
        // <div>aeeee</div>
        <Container maxWidth="xl">
            <Box>
                <Grid container spacing={3}>
                    <Grid item xs={6}>

                        <Box sx={{ my: 4 }}>
                            <Box>
                                <Button variant="contained" onClick={newPurchase}>
                                    Nova Venda
                                </Button>

                            </Box>
                        </Box>

                        <Grid container spacing={2}>
                            {purchase && products.sort(sortFunc).map((row, idx) => (
                                <Grid item xs={6}>
                                    <Button
                                        color="success"
                                        variant="contained"
                                        onClick={() => addItem(row)}
                                        key={row.ID}
                                        style={{
                                            display: 'block', marginBottom: 10,
                                            textAlign: 'left', width: '100%',
                                        }}>

                                        {(row.Key != "") ? `${row.Key} - ` : ""}
                                        {row.Code}<br />
                                        {Format(row.Price)}
                                    </Button>
                                </Grid>
                            ))}
                        </Grid>

                    </Grid>

                    <Grid item xs={6}>
                        <Box>
                            {purchase ? (<>
                                <Paper sx={{ padding: 3, marginY: 3, bgcolor: '#90a4ae' }} elevation={4}>
                                    <Typography variant='h5'>
                                        Total: {Format(calcTotal())}
                                    </Typography>
                                </Paper>

                                <Box>
                                    <TableContainer component={Paper}>
                                        <Table size="small" aria-label="simple table">
                                            <TableHead>
                                                <TableRow>
                                                    <TableCell>Item</TableCell>
                                                    <TableCell align="right">Qtd</TableCell>
                                                    <TableCell align="right">Price</TableCell>
                                                    <TableCell align="right">Total</TableCell>
                                                    <TableCell></TableCell>
                                                </TableRow>
                                            </TableHead>
                                            <TableBody>
                                                {purchase.Products.map((item, idx) => (
                                                    <TableRow
                                                        key={idx}
                                                        sx={{ '&:last-child td, &:last-child th': { border: 0 } }}
                                                    >
                                                        <TableCell component="th" scope="row">{item.ProductCode}</TableCell>
                                                        <TableCell align="right">{item.Quantity}</TableCell>
                                                        <TableCell align="right">{Format(item.ProductPrice)}</TableCell>
                                                        <TableCell align="right">{Format(item.Quantity * item.ProductPrice)}</TableCell>
                                                        <TableCell>
                                                            <IconButton aria-label="delete" size="small" onClick={() => incItem(idx)}>
                                                                <AddIcon fontSize="inherit" />
                                                            </IconButton>
                                                            <IconButton aria-label="delete" size="small" onClick={() => decItem(idx)}>
                                                                <RemoveIcon fontSize="inherit" />
                                                            </IconButton>
                                                            <IconButton aria-label="delete" size="small" onClick={() => delItem(idx)}>
                                                                <DeleteIcon fontSize="inherit" />
                                                            </IconButton>
                                                        </TableCell>
                                                    </TableRow>
                                                ))}
                                            </TableBody>
                                        </Table>
                                    </TableContainer>
                                </Box>
                                <Box>
                                    <Button variant="contained" onClick={placePurchase} sx={{ marginTop: 2 }}>
                                        Print
                                    </Button>
                                </Box>
                            </>) : ""}
                        </Box>
                    </Grid>
                </Grid>

            </Box>
        </Container>
    )
}

export default Index