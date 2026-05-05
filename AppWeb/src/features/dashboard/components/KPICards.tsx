import { useQuery } from "@tanstack/react-query";
import { dashboardService } from "../services/dashboardService";

const fmtNum = (n?: number) => (n == null ? "—" : Number(n).toLocaleString("es-BO"));
const fmtPct = (n?: number) => (n == null ? "—" : `${Number(n).toFixed(1)}%`);

export const KPICards = () => {
    const { data, isLoading, isError } = useQuery({
        queryKey: ["kpis"],
        queryFn: async () => {
            const { call } = dashboardService.getKPIs();
            const { data } = await call;
            return data;
        },
    });

    if (isLoading)
        return (
            <div className="grid grid-cols-2 md:grid-cols-4 xl:grid-cols-8 gap-3 animate-pulse">
                {[...Array(8)].map((_, i) => (
                    <div key={i} className="h-24 bg-gray-200 dark:bg-gray-800 rounded-xl" />
                ))}
            </div>
        );

    if (isError || !data)
        return <div className="text-red-500 p-4 bg-red-50 rounded-xl">Error cargando KPIs</div>;

    const cards = [
        { label: "Total Actas", value: fmtNum(data.totalActas), color: "border-blue-500", textColor: "text-blue-500" },
        { label: "Transcritas", value: fmtNum(data.actasTranscritas), color: "border-green-500", textColor: "text-green-500" },
        { label: "Observadas", value: fmtNum(data.actasObservadas), color: "border-amber-500", textColor: "text-amber-500" },
        { label: "Pendientes", value: fmtNum(data.actasPendientes), color: "border-slate-400", textColor: "text-slate-500" },
        { label: "% Procesadas", value: fmtPct(data.porcentajePublicadas), color: "border-indigo-500", textColor: "text-indigo-500" },
        { label: "Votos Válidos", value: fmtNum(data.votosValidos), color: "border-purple-500", textColor: "text-purple-500" },
        { label: "Ganador", value: `${data.ganador?.candidato ?? "—"} (${fmtPct(data.ganador?.porcentaje)})`, color: "border-emerald-600", textColor: "text-emerald-600" },
        { label: "Margen Victoria", value: fmtPct(data.margenVictoria), color: "border-red-500", textColor: "text-red-500" },
    ];

    return (
        <div className="grid grid-cols-2 md:grid-cols-4 xl:grid-cols-8 gap-3">
            {cards.map((c) => (
                <div
                    key={c.label}
                    className={`bg-white dark:bg-gray-800 rounded-xl px-5 py-4 shadow-sm hover:shadow-lg hover:-translate-y-0.5 transition-all duration-200 border-l-4 ${c.color} border-y border-r border-gray-200 dark:border-gray-700 cursor-default`}
                >
                    <div className="text-[11px] font-bold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">
                        {c.label}
                    </div>
                    <div className={`text-2xl font-extrabold ${c.textColor}`}>{c.value}</div>
                </div>
            ))}
        </div>
    );
};
