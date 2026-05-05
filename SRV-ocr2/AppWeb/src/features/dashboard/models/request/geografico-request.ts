export type GeoNivel = "departamento" | "municipio" | "provincia" | "recinto";

export interface GeoFilters {
    departamento?: string;
    municipio?: string;
    provincia?: string;
    recinto?: string;
    mesa?: string;
}

export interface GeograficoParams extends GeoFilters {
    nivel: GeoNivel;
}

export type GeograficoGroupBy = GeoNivel;
