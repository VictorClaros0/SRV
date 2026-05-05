import { useQueries } from "@tanstack/react-query";
import { useState } from "react";
import { rrvService } from "../services/rrvService";
import { dashboardService } from "../services/dashboardService";
import type { HeatmapMetric } from "../models/response/heatmap-response";

const colorScale = (v: number, max: number, hue: number) => {
    if (max <= 0) return "rgba(200,200,200,0.05)";
    const t = Math.min(1, v / max);
    const lightness = 90 - 50 * t;
    return `hsl(${hue}, 70%, ${lightness}%)`;
};

export const HeatmapComparativoPanel = () => {
    const [metric, setMetric] = useState<HeatmapMetric>("ganador");
    const rrvMetric = metric === "ganador" ? "votos" : "participacion";

    const [oficialQ, rrvQ] = useQueries({
        queries: [
            {
                queryKey: ["heatmap-of", metric],
                queryFn: async () => (await dashboardService.getHeatmap(metric).call).data,
            },
            {
                queryKey: ["heatmap-rrv", rrvMetric],
                queryFn: async () => (await rrvService.getHeatmap(rrvMetric).call).data,
            },
        ],
    });

    const ofItems = oficialQ.data?.items ?? [];
    const rrvItems = rrvQ.data?.datos ?? [];
    const maxOf = Math.max(...ofItems.map((i) => i.valor), 0);
    const maxRrv = Math.max(...rrvItems.map((i) => i.valor), 0);

    const Cell = ({
        nombre,
        valor,
        max,
        hue,
        sub,
    }: {
        nombre: string;
        valor: number;
        max: number;
        hue: number;
        sub?: string;
    }) => (
        <div
            className="rounded-lg p-2 border border-gray-200 dark:border-gray-700"
            style={{ backgroundColor: colorScale(valor, max, hue) }}
        >
            <div className="text-[11px] font-bold text-gray-900">{nombre}</div>
            <div className="text-base font-extrabold text-gray-900">
                {Number(valor).toLocaleString("es-BO")}
            </div>
            {sub && <div className="text-[10px] text-gray-700">{sub}</div>}
        </div>
    );

    return (
        <div className="bg-white dark:bg-gray-800 p-6 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm">
            <div className="flex items-center justify-between mb-3 gap-2">
                <h3 className="text-sm font-bold text-gray-900 dark:text-white">
                    Mapa de calor — RRV vs Oficial
                </h3>
                <select
                    value={metric}
                    onChange={(e) => setMetric(e.target.value as HeatmapMetric)}
                    className="bg-white dark:bg-gray-900 border border-gray-300 dark:border-gray-700 text-sm rounded-md px-3 py-1.5"
                >
                    <option value="ganador">% Ganador / Votos</option>
                    <option value="inconsistencias">Inconsistencias / Participación</option>
                </select>
            </div>
            <div className="grid grid-cols-2 gap-4">
                <div>
                    <h4 className="text-xs font-bold text-orange-500 mb-2">SRRV (rápido)</h4>
                    {rrvQ.isLoading ? (
                        <div className="h-48 bg-gray-100 dark:bg-gray-900 rounded-lg animate-pulse" />
                    ) : (
                        <div className="grid grid-cols-3 gap-2">
                            {rrvItems.map((it, i) => (
                                <Cell key={`r-${i}`} nombre={it.nombre} valor={it.valor} max={maxRrv} hue={25} />
                            ))}
                        </div>
                    )}
                </div>
                <div>
                    <h4 className="text-xs font-bold text-blue-500 mb-2">Oficial (SROV)</h4>
                    {oficialQ.isLoading ? (
                        <div className="h-48 bg-gray-100 dark:bg-gray-900 rounded-lg animate-pulse" />
                    ) : (
                        <div className="grid grid-cols-3 gap-2">
                            {ofItems.map((it, i) => (
                                <Cell
                                    key={`o-${i}`}
                                    nombre={it.nombre}
                                    valor={it.valor}
                                    max={maxOf}
                                    hue={210}
                                    sub={`Ganador: ${it.ganador}`}
                                />
                            ))}
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
};
