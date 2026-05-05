import { useQuery } from "@tanstack/react-query";
import { dashboardService } from "../services/dashboardService";

export const VelocidadPanel = () => {
    const { data, isLoading } = useQuery({
        queryKey: ["velocidad"],
        queryFn: async () => {
            const { call } = dashboardService.getResultados();
            const { data } = await call;
            return data;
        },
    });

    if (isLoading) return <div className="h-48 bg-gray-100 dark:bg-gray-800 rounded-xl animate-pulse" />;
    if (!data) return null;

    const v = data.velocidad ?? { actasUltimaHora: 0, actasUltimas24h: 0 };
    const promedio = v.actasUltimas24h > 0 ? Math.round(v.actasUltimas24h / 24) : 0;

    return (
        <div className="bg-white dark:bg-gray-800 p-6 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm h-48">
            <h3 className="text-sm font-bold text-gray-900 dark:text-white mb-4">Velocidad de procesamiento</h3>
            <div className="grid grid-cols-3 gap-3">
                <Stat label="Última hora" value={v.actasUltimaHora.toLocaleString("es-BO")} sub="actas" color="text-blue-500" />
                <Stat label="Últimas 24h" value={v.actasUltimas24h.toLocaleString("es-BO")} sub="actas" color="text-green-500" />
                <Stat label="Promedio/h" value={promedio.toLocaleString("es-BO")} sub="actas/hora" color="text-indigo-500" />
            </div>
        </div>
    );
};

const Stat = ({ label, value, sub, color }: { label: string; value: string; sub: string; color: string }) => (
    <div className="bg-gray-50 dark:bg-gray-900/40 rounded-lg p-3">
        <div className="text-[11px] uppercase font-bold text-gray-500">{label}</div>
        <div className={`text-2xl font-extrabold ${color}`}>{value}</div>
        <div className="text-[10px] text-gray-400">{sub}</div>
    </div>
);
