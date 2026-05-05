import { useQuery } from "@tanstack/react-query";
import { rrvService } from "../services/rrvService";

export const AnomaliasRRVPanel = () => {
    const { data, isLoading } = useQuery({
        queryKey: ["rrv-anomalias"],
        queryFn: async () => (await rrvService.getAnomalias().call).data,
    });

    if (isLoading) return <div className="h-64 bg-gray-100 dark:bg-gray-800 rounded-xl animate-pulse" />;
    if (!data) return null;

    const anomalias = data.anomalias ?? [];
    const total = data.total ?? anomalias.length;

    return (
        <div className="bg-white dark:bg-gray-800 p-6 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm">
            <h3 className="text-sm font-bold text-gray-900 dark:text-white mb-3">
                Anomalías SRRV ({total})
            </h3>
            {!anomalias.length ? (
                <p className="text-sm text-gray-400">Sin eventos anómalos.</p>
            ) : (
                <div className="space-y-2 max-h-72 overflow-auto">
                    {anomalias.slice(0, 50).map((e, i) => (
                        <div
                            key={(e._id as string) ?? i}
                            className="border border-gray-200 dark:border-gray-700 rounded-lg p-3"
                        >
                            <div className="flex items-center justify-between mb-1">
                                <span className="text-xs font-bold text-orange-600">
                                    {(e.tipo as string) ?? "EVENTO"}
                                </span>
                                <span className="text-[10px] text-gray-500 font-mono">
                                    {(e.fecha as string) ?? (e.timestamp as string) ?? ""}
                                </span>
                            </div>
                            {e.actaId && (
                                <div className="text-xs text-gray-700 dark:text-gray-300">
                                    Acta: <span className="font-mono">{e.actaId as string}</span>
                                </div>
                            )}
                            {e.payload != null && (
                                <pre className="text-[10px] text-gray-500 mt-1 overflow-auto max-h-20">
                                    {typeof e.payload === "string"
                                        ? e.payload
                                        : JSON.stringify(e.payload, null, 2)}
                                </pre>
                            )}
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
};
