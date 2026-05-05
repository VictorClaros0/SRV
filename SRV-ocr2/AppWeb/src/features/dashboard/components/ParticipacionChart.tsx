import { useQuery } from "@tanstack/react-query";
import { dashboardService } from "../services/dashboardService";
import { PieChart, Pie, Cell, Tooltip, ResponsiveContainer, Legend } from "recharts";
import type { DashboardFilters } from "./FiltersBar";

const COLORS = ["#10b981", "#ef4444", "#94a3b8"];

interface Props {
    filters?: DashboardFilters;
}

export const ParticipacionChart = ({ filters }: Props = {}) => {
    const f = filters ?? { departamento: "", municipio: "", provincia: "", recinto: "", mesa: "" };
    const { data, isLoading } = useQuery({
        queryKey: ["participacion", f.departamento, f.municipio, f.provincia, f.recinto, f.mesa],
        queryFn: async () => {
            const { call } = dashboardService.getResultados(f);
            const { data } = await call;
            return data;
        },
    });

    if (isLoading)
        return <div className="h-96 bg-gray-100 dark:bg-gray-800 rounded-xl animate-pulse" />;
    if (!data) return null;

    const desglose = data.desglose ?? [];
    const validos = desglose.reduce((a, b) => a + (b.votosValidos ?? 0), 0);
    const nulos = desglose.reduce((a, b) => a + (b.votosNulos ?? 0), 0);
    const blanco = desglose.reduce((a, b) => a + (b.votosBlanco ?? 0), 0);
    const total = validos + nulos + blanco;

    const pie = [
        { name: "Válidos", value: validos },
        { name: "Nulos", value: nulos },
        { name: "Blanco", value: blanco },
    ];

    return (
        <div className="bg-white dark:bg-gray-800 p-6 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm hover:shadow-md transition-shadow h-96 flex flex-col">
            <h3 className="text-sm font-bold text-gray-900 dark:text-white mb-1">Participación electoral</h3>
            <p className="text-[11px] text-gray-400 mb-2">
                Total emitidos: {total.toLocaleString("es-BO")} · Válidos {validos.toLocaleString("es-BO")} · Nulos {nulos.toLocaleString("es-BO")} · Blanco {blanco.toLocaleString("es-BO")}
            </p>
            <div className="flex-1 min-h-0">
                <ResponsiveContainer width="100%" height="100%">
                    <PieChart>
                        <Pie data={pie} dataKey="value" nameKey="name" outerRadius={100} label={(e) => `${e.name}: ${Number(e.value).toLocaleString("es-BO")}`}>
                            {pie.map((_, i) => (
                                <Cell key={i} fill={COLORS[i % COLORS.length]} />
                            ))}
                        </Pie>
                        <Tooltip formatter={(v) => Number(v).toLocaleString("es-BO")} />
                        <Legend />
                    </PieChart>
                </ResponsiveContainer>
            </div>
        </div>
    );
};
