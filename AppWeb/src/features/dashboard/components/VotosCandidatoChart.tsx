import { useQuery } from "@tanstack/react-query";
import { dashboardService } from "../services/dashboardService";
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, LabelList } from "recharts";
import type { DashboardFilters } from "./FiltersBar";

const COLORS = ["#3b82f6", "#10b981", "#f59e0b", "#ef4444"];

interface Props {
    filters?: DashboardFilters;
}

export const VotosCandidatoChart = ({ filters }: Props = {}) => {
    const f = filters ?? { departamento: "", municipio: "", provincia: "", recinto: "", mesa: "" };
    const { data, isLoading } = useQuery({
        queryKey: ["votosCandidato", f.departamento, f.municipio, f.provincia, f.recinto, f.mesa],
        queryFn: async () => {
            const { call } = dashboardService.getResultados(f);
            const { data } = await call;
            return data;
        },
    });

    if (isLoading)
        return <div className="h-96 bg-gray-100 dark:bg-gray-800 rounded-xl animate-pulse" />;
    if (!data) return null;

    const comparativa = data.comparativa ?? [];
    const totalVotos = comparativa.reduce((a, b) => a + b.votos, 0);

    const chartData = comparativa.map((it, i) => ({
        candidato: it.candidato,
        votos: it.votos,
        porcentaje: it.pct,
        fill: COLORS[i % COLORS.length],
    }));

    return (
        <div className="bg-white dark:bg-gray-800 p-6 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm hover:shadow-md transition-shadow h-96 flex flex-col">
            <h3 className="text-sm font-bold text-gray-900 dark:text-white mb-1">
                Votos por candidato (oficial)
            </h3>
            <p className="text-[11px] text-gray-400 mb-2">
                Total candidatos: {totalVotos.toLocaleString("es-BO")} · Margen: {data.margen_victoria?.toFixed?.(2) ?? "0"}%
            </p>
            <div className="flex-1 min-h-0">
                <ResponsiveContainer width="100%" height="100%">
                    <BarChart data={chartData} margin={{ top: 20, right: 20, left: 0, bottom: 5 }}>
                        <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                        <XAxis dataKey="candidato" stroke="#6b7280" fontSize={12} />
                        <YAxis stroke="#6b7280" fontSize={12} />
                        <Tooltip
                            cursor={{ fill: "rgba(99, 102, 241, 0.08)" }}
                            contentStyle={{ backgroundColor: "#1F2937", borderColor: "#374151", color: "#fff", borderRadius: 8 }}
                            formatter={(v, _n, p) => {
                                const pct = (p as { payload?: { porcentaje?: number } })?.payload?.porcentaje;
                                return [
                                    `${Number(v).toLocaleString("es-BO")}${pct != null ? ` (${pct.toFixed(2)}%)` : ""}`,
                                    "Votos",
                                ];
                            }}
                        />
                        <Bar dataKey="votos" radius={[6, 6, 0, 0]} isAnimationActive animationDuration={800}>
                            <LabelList dataKey="porcentaje" position="top" formatter={(v) => `${Number(v).toFixed(2)}%`} fontSize={11} />
                        </Bar>
                    </BarChart>
                </ResponsiveContainer>
            </div>
        </div>
    );
};
