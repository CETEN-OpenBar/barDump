export interface Product {
    nom: string;
    quantite: number;
}

export interface MonthlyData {
    nombre_transactions: number;
    produit_le_plus_achete: string | null;
}

export interface MonthlyArticles {
    [key: string]: MonthlyData;
}

export interface TopConsumerProduct {
    nom: string;
    quantite: number;
    total: number;
}

export interface User {
    account_id: string;
    account_name: string;
    email: string;
    total_depense: number;
    top_3_produits: Product[];
    nombre_produits_differents: number;
    articles_par_mois: MonthlyArticles;
    top_consommateur_produit: TopConsumerProduct | null;
    rang: number;
    rang_pourcentage: number;
    rang_produits_differents: number;
    rang_pourcentage_produits_differents: number;
    nombre_transaction_remote: number;
    nombre_transaction_non_remote: number;
}

export interface DumpInfo {
    title: string;
    type: string;
    file: string;
    logo1?: string;
    logo2?: string;
    start_date?: string;
    end_date?: string;
}

export interface AllData {
    utilisateurs: User[];
}
