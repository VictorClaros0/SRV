import { useQuery } from "@tanstack/react-query";
import { dashboardService } from "../services/dashboardService";

export const AnomaliasPanel = () => {
    const { data, isLoading } = useQuery({
        queryKey: ["anomalias"],
        queryFn: async () => {
            const { call } = dashboardService.getAnomalias();
            const { data } = await call;
            return data;
        },
    });

    if (isLoading) return <div className="h-64 bg-gray-100 dark:bg-gray-800 rounded-xl animate-pulse" />;
    if (!data) return null;

    const anomalias = data.anomalias ?? [];
    const total = data.total ?? anomalias.length;

    const sevColor: Record<string, string> = {
        alta: "text-red-500 bg-red-50",
        media: "text-amber-600 bg-amber-50",
        baja: "text-blue-500 bg-blue-50",
    };

    return (
        <div className="bg-white dark:bg-gray-800 p-6 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm">
            <h3 className="text-sm font-bold text-gray-900 dark:text-white mb-3">
                Detección de anomalías ({total})
            </h3>
            {!anomalias.length ? (
                <p className="text-sm text-gray-400">Sin anomalías detectadas.</p>
            ) : (
                <div className="space-y-2 max-h-72 overflow-auto">
                    {anomalias.map((a, i) => (
                        <div key={`${a.codigoActa}-${a.tipo}-${i}`} className="border border-gray-200 dark:border-gray-700 rounded-lg p-3">
                            <div className="flex items-center justify-between mb-1">
                                <span className="font-mono text-xs text-gray-700 dark:text-gray-300">
                                    Acta #{a.codigoActa} · Mesa {a.nroMesa}
                                </span>
                                <span className={`text-[10px] font-bold uppercase px-2 py-0.5 rounded-full ${sevColor[a.severidad] || ""}`}>
                                    {a.severidad}
                                </span>
                            </div>
                            <div className="text-xs font-bold text-gray-900 dark:text-white">{a.tipo}</div>
                            <div className="text-xs text-gray-500">{a.detalle}</div>
                            {a.valoresEsperados != null && a.valoresReales != null && (
                                <div className="text-[11px] text-gray-400 mt-1">
                                    Esperado: {a.valoresEsperados} · Real: {a.valoresReales}
                                </div>
                            )}
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
};
