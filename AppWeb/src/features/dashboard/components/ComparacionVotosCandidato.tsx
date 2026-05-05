import { useQueries } from "@tanstack/react-query";
import { rrvService } from "../services/rrvService";
import { dashboardService } from "../services/dashboardService";
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from "recharts";
import type { RrvVotosCandidatoResponse } from "../models/response/rrv-responses";
import type { ResultadosResponse } from "../models/response/resultados-response";

interface Row {
    candidato: string;
    rrv: number;
    oficial: number;
    delta: number;
}

export const ComparacionVotosCandidato = () => {
    const [rrvQ, oficialQ] = useQueries({
        queries: [
            {
                queryKey: ["rrv-votos"],
                queryFn: async () => {
                    const { call } = rrvService.getVotosCandidato();
                    return (await call).data as RrvVotosCandidatoResponse;
                },
            },
            {
                queryKey: ["oficial-resultados-comp"],
                queryFn: async () => {
                    const { call } = dashboardService.getResultados();
                    return (await call).data as ResultadosResponse;
                },
            },
        ],
    });

    if (rrvQ.isLoading || oficialQ.isLoading)
        return <div className="h-96 bg-gray-100 dark:bg-gray-800 rounded-xl animate-pulse" />;

    const rrvCands = rrvQ.data?.candidatos ?? [];
    const oficialComp = oficialQ.data?.comparativa ?? [];

    const oficialMap: Record<string, number> = {};
    oficialComp.forEach((c) => {
        oficialMap[c.candidato] = c.votos;
    });

    const rows: Row[] = rrvCands.map((c) => {
        const key = c.campo?.toUpperCase() ?? c.nombre;
        const oficial = oficialMap[key] ?? 0;
        return {
            candidato: c.nombre || key,
            rrv: c.votos,
            oficial,
            delta: c.votos - oficial,
        };
    });

    return (
        <div className="bg-white dark:bg-gray-800 p-6 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm h-96 flex flex-col">
            <h3 className="text-sm font-bold text-gray-900 dark:text-white mb-1">
                Votos por candidato — RRV vs Oficial
            </h3>
            <p className="text-[11px] text-gray-400 mb-2">
                Diferencia total: {rows.reduce((a, b) => a + b.delta, 0).toLocaleString("es-BO")}
            </p>
            <div className="flex-1 min-h-0">
                <ResponsiveContainer width="100%" height="100%">
                    <BarChart data={rows} margin={{ top: 10, right: 20, left: 0, bottom: 5 }}>
                        <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                        <XAxis dataKey="candidato" stroke="#6b7280" fontSize={12} />
                        <YAxis stroke="#6b7280" fontSize={12} />
                        <Tooltip
                            contentStyle={{ backgroundColor: "#1F2937", borderColor: "#374151", color: "#fff", borderRadius: 8 }}
                            formatter={(v) => Number(v).toLocaleString("es-BO")}
                        />
                        <Legend />
                        <Bar dataKey="rrv" name="RRV" fill="#f97316" radius={[6, 6, 0, 0]} />
                        <Bar dataKey="oficial" name="Oficial" fill="#3b82f6" radius={[6, 6, 0, 0]} />
                    </BarChart>
                </ResponsiveContainer>
            </div>
        </div>
    );
};
