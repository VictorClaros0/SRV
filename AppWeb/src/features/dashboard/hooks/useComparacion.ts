import { useQueries } from "@tanstack/react-query";
import { dashboardService } from "../services/dashboardService";
import { rrvService } from "../services/rrvService";
import type { KpisResponse } from "../models/response/kpis-response";
import type { RrvKpis } from "../models/response/rrv-responses";

export interface CandidatoDelta {
    candidato: string;
    rrv: number;
    oficial: number;
    delta: number;
    deltaPct: number;
}

export interface ComparacionResumen {
    rrv: RrvKpis | null;
    oficial: KpisResponse | null;
    deltaActas: number;
    deltaValidos: number;
    deltaNulos: number;
    deltaBlanco: number;
    candidatos: CandidatoDelta[];
    confiabilidadRRV: number;
    isLoading: boolean;
    isError: boolean;
}

const pct = (a: number, b: number): number => {
    const base = Math.max(Math.abs(a), Math.abs(b));
    if (base === 0) return 0;
    return Math.round(((a - b) / base) * 10000) / 100;
};

export const useComparacion = (): ComparacionResumen => {
    const results = useQueries({
        queries: [
            {
                queryKey: ["rrv-kpis"],
                queryFn: async () => {
                    const { call } = rrvService.getKPIs();
                    const { data } = await call;
                    return data;
                },
            },
            {
                queryKey: ["oficial-kpis"],
                queryFn: async () => {
                    const { call } = dashboardService.getKPIs();
                    const { data } = await call;
                    return data;
                },
            },
        ],
    });

    const rrv = (results[0].data as RrvKpis) ?? null;
    const oficial = (results[1].data as KpisResponse) ?? null;

    const isLoading = results.some((r) => r.isLoading);
    const isError = results.some((r) => r.isError);

    if (!rrv || !oficial) {
        return {
            rrv,
            oficial,
            deltaActas: 0,
            deltaValidos: 0,
            deltaNulos: 0,
            deltaBlanco: 0,
            candidatos: [],
            confiabilidadRRV: 0,
            isLoading,
            isError,
        };
    }

    const deltaActas = rrv.totalActasProcesadas - oficial.actasTranscritas;
    const deltaValidos = rrv.votosValidos - oficial.votosValidos;
    const deltaNulos = rrv.votosNulos - oficial.votosNulos;
    const deltaBlanco = rrv.votosBlancos - oficial.votosBlanco;

    const candidatos: CandidatoDelta[] = [
        { candidato: "P1", rrv: rrv.votosP1, oficial: 0, delta: 0, deltaPct: 0 },
        { candidato: "P2", rrv: rrv.votosP2, oficial: 0, delta: 0, deltaPct: 0 },
        { candidato: "P3", rrv: rrv.votosP3, oficial: 0, delta: 0, deltaPct: 0 },
        { candidato: "P4", rrv: rrv.votosP4, oficial: 0, delta: 0, deltaPct: 0 },
    ];

    const oficialP: Record<string, number> = {
        P1: 0, P2: 0, P3: 0, P4: 0,
    };
    if (oficial.ganador) oficialP[oficial.ganador.candidato] = oficial.ganador.votos;
    if (oficial.segundo) oficialP[oficial.segundo.candidato] = oficial.segundo.votos;

    candidatos.forEach((c) => {
        c.oficial = oficialP[c.candidato] ?? 0;
        c.delta = c.rrv - c.oficial;
        c.deltaPct = pct(c.rrv, c.oficial);
    });

    const totalRrv = rrv.votosValidos + rrv.votosNulos + rrv.votosBlancos;
    const totalOficial = oficial.votosValidos + oficial.votosNulos + oficial.votosBlanco;
    const confiabilidadRRV =
        totalOficial > 0
            ? Math.max(0, 100 - (Math.abs(totalRrv - totalOficial) / totalOficial) * 100)
            : 0;

    return {
        rrv,
        oficial,
        deltaActas,
        deltaValidos,
        deltaNulos,
        deltaBlanco,
        candidatos,
        confiabilidadRRV: Math.round(confiabilidadRRV * 100) / 100,
        isLoading,
        isError,
    };
};
