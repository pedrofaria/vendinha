import React, { useEffect, useState } from 'react';
import Container from '@mui/material/Container';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import DeleteIcon from '@mui/icons-material/Delete';
import EditIcon from '@mui/icons-material/Edit';
import { model } from '../../wailsjs/go/models';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Paper from '@mui/material/Paper';
import { Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, FormControl, IconButton, InputAdornment, InputLabel, OutlinedInput } from '@mui/material';
import { AddProduct, DelProduct, List, UpdateProduct } from '../../wailsjs/go/product/Service';

interface AddProductFormState {
    id?: number;
    code: string;
    price: number;
    order: number;
    key: string;
}

interface ListTableProps { list: model.Product[]; onUpdate: () => void; }

const ListTable = (props: ListTableProps) => {
    const [openEditDialog, setOpenEditDialog] = useState(false);
    const [currentFormData, setCurrentFormData] = useState<AddProductFormState>();

    const handleDelete = (id: number) => {
        DelProduct(id).then(() => props.onUpdate());
    }

    const edit = (prod: model.Product) => {
        setCurrentFormData({
            id: prod.ID,
            code: prod.Code,
            price: prod.Price,
            order: prod.Order,
            key: prod.Key
        })
        setOpenEditDialog(true)
    }

    const handleUpdate = (formData: AddProductFormState) => {
        console.log(formData);
        UpdateProduct(formData.id!!, formData.code, formData.price, formData.order, formData.key)
            .then(() => {
                setOpenEditDialog(false);
                props.onUpdate();
            })
    }

    return (<>
        <FormDialog
            formData={currentFormData}
            handleSave={handleUpdate}
            handleClose={() => setOpenEditDialog(false)}
            title="Editar produto"
            open={openEditDialog} />

        <TableContainer component={Paper}>
            <Table sx={{ minWidth: 650 }} size='small' aria-label="simple table">
                <TableHead>
                    <TableRow>
                        <TableCell>ID</TableCell>
                        <TableCell align="right">Code</TableCell>
                        <TableCell align="right">Price</TableCell>
                        <TableCell align="center">Ord</TableCell>
                        <TableCell align="center">Key</TableCell>
                        <TableCell></TableCell>
                    </TableRow>
                </TableHead>
                <TableBody>
                    {props.list.map((row) => (
                        <TableRow
                            key={row.ID}
                            sx={{ '&:last-child td, &:last-child th': { border: 0 } }}
                        >
                            <TableCell component="th" scope="row">{row.ID}</TableCell>
                            <TableCell align="right">{row.Code}</TableCell>
                            <TableCell align="right">{row.Price}</TableCell>
                            <TableCell align="center">{row.Order}</TableCell>
                            <TableCell align="center">{row.Key}</TableCell>
                            <TableCell>
                                <IconButton size="small" onClick={() => edit(row)}>
                                    <EditIcon color="primary" />
                                </IconButton>

                                <IconButton size="small" onClick={() => handleDelete(row.ID)}>
                                    <DeleteIcon color="primary" />
                                </IconButton>
                            </TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </TableContainer>
    </>)
}

type FormDialogProps = {
    open: boolean;
    title: string;
    handleSave: (formData: AddProductFormState) => void;
    handleClose: () => void;
    formData?: AddProductFormState;
};

const FormDialog = (props: FormDialogProps) => {
    const [formData, setFormData] = useState<AddProductFormState>({
        code: '',
        price: 0,
        order: 0,
        key: '',
    });

    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        let { name, value } = e.target;

        if (name === 'price' || name === 'order') {
            setFormData(prevData => ({ ...prevData, [name]: parseInt(value) }))
            return;
        }

        setFormData(prevData => ({ ...prevData, [name]: value }))
    }

    const handleSave = () => {
        props.handleSave(formData);
    }

    const handleClose = () => {
        props.handleClose();
    };

    useEffect(() => {
        if (props.formData !== null && props.formData !== undefined) {
            setFormData(props.formData)
        }
    }, [props.formData]);

    return (
        <React.Fragment>
            <Dialog open={props.open} onClose={handleClose}>
                <DialogTitle>{props.title}</DialogTitle>
                <DialogContent>
                    <DialogContentText>
                        Preencha os dados do novo produto.
                    </DialogContentText>
                    <FormControl fullWidth sx={{ m: 1 }}>
                        <InputLabel htmlFor="outlined-adornment-code">Codigo</InputLabel>
                        <OutlinedInput
                            id="outlined-adornment-code"
                            label="Codigo"
                            name="code"
                            value={formData.code}
                            onChange={handleChange}
                        />
                    </FormControl>
                    <FormControl fullWidth sx={{ m: 1 }}>
                        <InputLabel htmlFor="outlined-adornment-amount">Preco</InputLabel>
                        <OutlinedInput
                            id="outlined-adornment-amount"
                            startAdornment={<InputAdornment position="start">R$</InputAdornment>}
                            label="Preco"
                            name="price"
                            value={formData.price}
                            onChange={handleChange}
                        />
                    </FormControl>
                    <FormControl fullWidth sx={{ m: 1 }}>
                        <InputLabel htmlFor="outlined-adornment-order">Ordem</InputLabel>
                        <OutlinedInput
                            id="outlined-adornment-order"
                            label="Ordem"
                            name="order"
                            defaultValue="0"
                            value={formData.order}
                            onChange={handleChange}
                        />
                    </FormControl>
                    <FormControl fullWidth sx={{ m: 1 }}>
                        <InputLabel htmlFor="outlined-adornment-key">Shortcut</InputLabel>
                        <OutlinedInput
                            id="outlined-adornment-key"
                            label="Shortcut Key"
                            name="key"
                            value={formData.key}
                            onChange={handleChange}
                        />
                    </FormControl>
                </DialogContent>
                <DialogActions>
                    <Button onClick={handleClose}>Cancelar</Button>
                    <Button onClick={handleSave}>Salvar</Button>
                </DialogActions>
            </Dialog>
        </React.Fragment>
    );
}

const Products = () => {
    const [products, setProducts] = useState(Array<model.Product>);
    const [openSaveDialog, setOpenSaveDialog] = useState(false);


    const list = () => {
        List().then(data => {
            setProducts(data);
        })
    }

    useEffect(() => {
        list();
    }, []);

    const onDelete = () => {
        list();
    }

    const handleSave = (formData: AddProductFormState) => {
        console.log(formData);
        AddProduct(formData.code, formData.price, formData.order, formData.key)
            .then(() => list())
            .finally(() => setOpenSaveDialog(false))
    }



    return (
        <Container>
            <Box sx={{ my: 4 }}>
                <Typography variant="h4" component="h1" gutterBottom>
                    Lista de produtos
                </Typography>

                <Button variant="outlined" onClick={() => setOpenSaveDialog(true)}>
                    Novo Produto
                </Button>

                <FormDialog
                    handleSave={handleSave}
                    handleClose={() => setOpenSaveDialog(false)}
                    title="Criar novo produto"
                    open={openSaveDialog} />


            </Box>

            <Box>
                <ListTable list={products} onUpdate={onDelete} />
            </Box>
        </Container>
    )
}

export default Products