export namespace main {
	
	export class Cartela {
	    reais: number;
	    preco: number;
	    nome: string;
	    conteudo: string;
	
	    static createFrom(source: any = {}) {
	        return new Cartela(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.reais = source["reais"];
	        this.preco = source["preco"];
	        this.nome = source["nome"];
	        this.conteudo = source["conteudo"];
	    }
	}
	export class Conta {
	    id: number;
	    eventoId: number;
	    nome: string;
	    totalPendente: number;
	    criadoEm: string;
	
	    static createFrom(source: any = {}) {
	        return new Conta(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.eventoId = source["eventoId"];
	        this.nome = source["nome"];
	        this.totalPendente = source["totalPendente"];
	        this.criadoEm = source["criadoEm"];
	    }
	}
	export class ContaSaldo {
	    id: number;
	    eventoId: number;
	    nome: string;
	    totalPendente: number;
	    numAberto: number;
	    totalQuitado: number;
	    criadoEm: string;
	
	    static createFrom(source: any = {}) {
	        return new ContaSaldo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.eventoId = source["eventoId"];
	        this.nome = source["nome"];
	        this.totalPendente = source["totalPendente"];
	        this.numAberto = source["numAberto"];
	        this.totalQuitado = source["totalQuitado"];
	        this.criadoEm = source["criadoEm"];
	    }
	}
	export class Evento {
	    id: number;
	    nome: string;
	    ativo: boolean;
	    vendeCartela: boolean;
	    criadoEm: string;
	
	    static createFrom(source: any = {}) {
	        return new Evento(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.nome = source["nome"];
	        this.ativo = source["ativo"];
	        this.vendeCartela = source["vendeCartela"];
	        this.criadoEm = source["criadoEm"];
	    }
	}
	export class Grupo {
	    id: number;
	    eventoId: number;
	    nome: string;
	    cor: string;
	    ordem: number;
	
	    static createFrom(source: any = {}) {
	        return new Grupo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.eventoId = source["eventoId"];
	        this.nome = source["nome"];
	        this.cor = source["cor"];
	        this.ordem = source["ordem"];
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
	    grupoId: number;
	    atalho: number;
	    ordem: number;
	
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
	        this.grupoId = source["grupoId"];
	        this.atalho = source["atalho"];
	        this.ordem = source["ordem"];
	    }
	}
	export class GrupoProdutos {
	    id: number;
	    eventoId: number;
	    nome: string;
	    cor: string;
	    ordem: number;
	    produtos: Produto[];
	
	    static createFrom(source: any = {}) {
	        return new GrupoProdutos(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.eventoId = source["eventoId"];
	        this.nome = source["nome"];
	        this.cor = source["cor"];
	        this.ordem = source["ordem"];
	        this.produtos = this.convertValues(source["produtos"], Produto);
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
	export class ProdutoVenda {
	    id: number;
	    nome: string;
	    preco: number;
	    estoqueTipo: string;
	    quantidade: number;
	    esgotado: boolean;
	    atalho: number;
	
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
	        this.atalho = source["atalho"];
	    }
	}
	export class GrupoVenda {
	    id: number;
	    eventoId: number;
	    nome: string;
	    cor: string;
	    ordem: number;
	    produtos: ProdutoVenda[];
	
	    static createFrom(source: any = {}) {
	        return new GrupoVenda(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.eventoId = source["eventoId"];
	        this.nome = source["nome"];
	        this.cor = source["cor"];
	        this.ordem = source["ordem"];
	        this.produtos = this.convertValues(source["produtos"], ProdutoVenda);
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
	export class HoraVendas {
	    hora: number;
	    vendas: number;
	
	    static createFrom(source: any = {}) {
	        return new HoraVendas(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hora = source["hora"];
	        this.vendas = source["vendas"];
	    }
	}
	export class ImpressorasInfo {
	    nomes: string[];
	    padrao: string;
	    selecionada: string;
	
	    static createFrom(source: any = {}) {
	        return new ImpressorasInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.nomes = source["nomes"];
	        this.padrao = source["padrao"];
	        this.selecionada = source["selecionada"];
	    }
	}
	export class PedidoResumo {
	    id: number;
	    eventoId: number;
	    numero: number;
	    total: number;
	    forma: string;
	    contaId: number;
	    contaNome: string;
	    quitadoEm: string;
	    canceladoEm: string;
	    criadoEm: string;
	
	    static createFrom(source: any = {}) {
	        return new PedidoResumo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.eventoId = source["eventoId"];
	        this.numero = source["numero"];
	        this.total = source["total"];
	        this.forma = source["forma"];
	        this.contaId = source["contaId"];
	        this.contaNome = source["contaNome"];
	        this.quitadoEm = source["quitadoEm"];
	        this.canceladoEm = source["canceladoEm"];
	        this.criadoEm = source["criadoEm"];
	    }
	}
	export class ListaPedidos {
	    pedidos: PedidoResumo[];
	    total: number;
	    pagina: number;
	    totalPaginas: number;
	
	    static createFrom(source: any = {}) {
	        return new ListaPedidos(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pedidos = this.convertValues(source["pedidos"], PedidoResumo);
	        this.total = source["total"];
	        this.pagina = source["pagina"];
	        this.totalPaginas = source["totalPaginas"];
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
	export class PedidoItem {
	    id: number;
	    produtoId: number;
	    cartelaReais: number;
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
	        this.cartelaReais = source["cartelaReais"];
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
	    forma: string;
	    contaId: number;
	    quitadoEm: string;
	    canceladoEm: string;
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
	        this.forma = source["forma"];
	        this.contaId = source["contaId"];
	        this.quitadoEm = source["quitadoEm"];
	        this.canceladoEm = source["canceladoEm"];
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
	
	
	
	
	export class ResumoEvento {
	    receitaTotal: number;
	    numPedidos: number;
	    numProdutosVendidos: number;
	    vendasInicioHora: number;
	    vendasPorHora: HoraVendas[];
	
	    static createFrom(source: any = {}) {
	        return new ResumoEvento(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.receitaTotal = source["receitaTotal"];
	        this.numPedidos = source["numPedidos"];
	        this.numProdutosVendidos = source["numProdutosVendidos"];
	        this.vendasInicioHora = source["vendasInicioHora"];
	        this.vendasPorHora = this.convertValues(source["vendasPorHora"], HoraVendas);
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

}

