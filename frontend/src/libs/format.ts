const Format = (val: number): string => {
    // const x = val.toString();
    // if (val < 100) {
    //     return val.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' });
    // }

    // return "R$ " + x.slice(0, x.length - 2) + "," + x.slice(x.length - 2, x.length);

    return (val / 100).toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' });
}

export default Format;