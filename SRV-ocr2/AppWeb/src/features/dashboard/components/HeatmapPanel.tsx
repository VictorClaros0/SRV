import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { dashboardService } from "../services/dashboardService";
import type { HeatmapMetric } from "../models/response/heatmap-response";

const colorScale = (v: number, max: number) => {
    if (max <= 0) return "rgba(59,130,246,0.05)";
    const t = Math.min(1, v / max);
    const r = Math.round(239 - 180 * t);
    const g = Math.round(246 - 100 * t);
    const b = Math.round(255 - 100 * t);
    return `rgb(${r},${g},${b})`;
};

export const HeatmapPanel = () => {
    const [metric, setMetric] = useState<HeatmapMetric>("ganador");
    const { data, isLoading } = useQuery({
        queryKey: ["heatmap", metric],
        queryFn: async () => {
            const { call } = dashboardService.getHeatmap(metric);
            const { data } = await call;
            return data;
        },
    });

    if (isLoading) return <div className="h-96 bg-gray-100 dark:bg-gray-800 rounded-xl animate-pulse" />;
    if (!data?.items?.length)
        return (
            <div className="bg-white dark:bg-gray-800 p-6 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm h-96">
                <h3 className="text-sm font-bold text-gray-900 dark:text-white mb-3">Mapa de calor</h3>
                <p className="text-sm text-gray-400">Sin datos.</p>
            </div>
        );

    const max = Math.max(...data.items.map((i) => i.valor), 0);

    return (
        <div className="bg-white dark:bg-gray-800 p-6 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm h-96 flex flex-col">
            <div className="flex items-center justify-between mb-3 gap-2">
                <h3 className="text-sm font-bold text-gray-900 dark:text-white">Mapa de calor</h3>
                <select
                    value={metric}
                    onChange={(e) => setMetric(e.target.value as HeatmapMetric)}
                    className="bg-white dark:bg-gray-900 border border-gray-300 dark:border-gray-700 text-sm rounded-md px-3 py-1.5"
                >
                    <option value="ganador">% Ganador</option>
                    <option value="inconsistencias">Inconsistencias</option>
                    <option value="procesamiento">Procesamiento</option>
                </select>
            </div>
            <div className="flex-1 overflow-auto grid grid-cols-2 md:grid-cols-3 gap-2 content-start">
                {data.items.map((it, i) => (
                    <div
                        key={`${it.nombre}-${i}`}
                        className="rounded-lg p-3 border border-gray-200 dark:border-gray-700"
                        style={{ backgroundColor: colorScale(it.valor, max) }}
                    >
                        <div className="text-xs font-bold text-gray-900">{it.nombre}</div>
                        <div className="text-xl font-extrabold text-gray-900">
                            {metric === "ganador" ? `${it.valor.toFixed(1)}%` : it.valor.toLocaleString("es-BO")}
                        </div>
                        <div className="text-[11px] text-gray-700">
                            Ganador: {it.ganador} · Votos: {it.votos.toLocaleString("es-BO")}
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
};
