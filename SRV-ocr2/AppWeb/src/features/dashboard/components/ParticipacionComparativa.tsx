import { useQueries } from "@tanstack/react-query";
import { rrvService } from "../services/rrvService";
import { dashboardService } from "../services/dashboardService";
import { PieChart, Pie, Cell, Tooltip, ResponsiveContainer, Legend } from "recharts";

const COLORS = ["#10b981", "#ef4444", "#94a3b8"];

const Pie3 = ({ title, validos, nulos, blanco }: { title: string; validos: number; nulos: number; blanco: number }) => {
    const data = [
        { name: "Válidos", value: validos },
        { name: "Nulos", value: nulos },
        { name: "Blanco", value: blanco },
    ];
    const total = validos + nulos + blanco;
    return (
        <div className="flex flex-col h-72">
            <h4 className="text-xs font-bold text-gray-700 dark:text-gray-300 mb-1">{title}</h4>
            <p className="text-[10px] text-gray-400 mb-1">Total: {total.toLocaleString("es-BO")}</p>
            <div className="flex-1 min-h-0">
                <ResponsiveContainer width="100%" height="100%">
                    <PieChart>
                        <Pie data={data} dataKey="value" nameKey="name" outerRadius={70} label={(e) => `${e.name}: ${Number(e.value).toLocaleString("es-BO")}`}>
                            {data.map((_, i) => <Cell key={i} fill={COLORS[i]} />)}
                        </Pie>
                        <Tooltip formatter={(v) => Number(v).toLocaleString("es-BO")} />
                        <Legend wrapperStyle={{ fontSize: 11 }} />
                    </PieChart>
                </ResponsiveContainer>
            </div>
        </div>
    );
};

export const ParticipacionComparativa = () => {
    const [rrvQ, ofQ] = useQueries({
        queries: [
            {
                queryKey: ["rrv-kpis-part"],
                queryFn: async () => (await rrvService.getKPIs().call).data,
            },
            {
                queryKey: ["of-kpis-part"],
                queryFn: async () => (await dashboardService.getKPIs().call).data,
            },
        ],
    });

    if (rrvQ.isLoading || ofQ.isLoading)
        return <div className="h-80 bg-gray-100 dark:bg-gray-800 rounded-xl animate-pulse" />;

    return (
        <div className="bg-white dark:bg-gray-800 p-6 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm">
            <h3 className="text-sm font-bold text-gray-900 dark:text-white mb-3">
                Participación electoral — RRV vs Oficial
            </h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <Pie3
                    title="SRRV (rápido)"
                    validos={rrvQ.data?.votosValidos ?? 0}
                    nulos={rrvQ.data?.votosNulos ?? 0}
                    blanco={rrvQ.data?.votosBlancos ?? 0}
                />
                <Pie3
                    title="Cómputo Oficial"
                    validos={ofQ.data?.votosValidos ?? 0}
                    nulos={ofQ.data?.votosNulos ?? 0}
                    blanco={ofQ.data?.votosBlanco ?? 0}
                />
            </div>
        </div>
    );
};
