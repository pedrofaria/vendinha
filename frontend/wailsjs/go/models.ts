export namespace main {
	
	export class Evento {
	    id: number;
	    nome: string;
	    ativo: boolean;
	    criadoEm: string;
	
	    static createFrom(source: any = {}) {
	        return new Evento(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.nome = source["nome"];
	        this.ativo = source["ativo"];
	        this.criadoEm = source["criadoEm"];
	    }
	}
	export class PedidoItem {
	    id: number;
	    produtoId: number;
	    nome: string;
	    precoUnit: number;
	    qtd: number;
	    subtotal: number;
	
	    static createFrom(source: any = {}) {
	        return new PedidoItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.produtoId = source["produtoId"];
	        this.nome = source["nome"];
	        this.precoUnit = source["precoUnit"];
	        this.qtd = source["qtd"];
	        this.subtotal = source["subtotal"];
	    }
	}
	export class Pedido {
	    id: number;
	    eventoId: number;
	    numero: number;
	    total: number;
	    criadoEm: string;
	    itens: PedidoItem[];
	
	    static createFrom(source: any = {}) {
	        return new Pedido(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.eventoId = source["eventoId"];
	        this.numero = source["numero"];
	        this.total = source["total"];
	        this.criadoEm = source["criadoEm"];
	        this.itens = this.convertValues(source["itens"], PedidoItem);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class Produto {
	    id: number;
	    eventoId: number;
	    nome: string;
	    preco: number;
	    estoqueTipo: string;
	    quantidade: number;
	    ativo: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Produto(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.eventoId = source["eventoId"];
	        this.nome = source["nome"];
	        this.preco = source["preco"];
	        this.estoqueTipo = source["estoqueTipo"];
	        this.quantidade = source["quantidade"];
	        this.ativo = source["ativo"];
	    }
	}
	export class ProdutoVenda {
	    id: number;
	    nome: string;
	    preco: number;
	    estoqueTipo: string;
	    quantidade: number;
	    esgotado: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ProdutoVenda(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.nome = source["nome"];
	        this.preco = source["preco"];
	        this.estoqueTipo = source["estoqueTipo"];
	        this.quantidade = source["quantidade"];
	        this.esgotado = source["esgotado"];
	    }
	}

}

