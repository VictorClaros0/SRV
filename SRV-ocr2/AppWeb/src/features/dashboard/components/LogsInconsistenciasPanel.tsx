import { useQuery } from "@tanstack/react-query";
import { rrvService } from "../services/rrvService";

export const LogsInconsistenciasPanel = () => {
    const { data, isLoading } = useQuery({
        queryKey: ["rrv-logs-inconsistencias"],
        queryFn: async () => (await rrvService.getLogInconsistencias().call).data,
    });

    if (isLoading) return <div className="h-64 bg-gray-100 dark:bg-gray-800 rounded-xl animate-pulse" />;
    if (!data) return null;

    const registros = data.registros ?? [];

    const camposCount: Record<string, number> = {};
    registros.forEach((r) => {
        (r.camposIlegibles ?? []).forEach((c) => {
            camposCount[c] = (camposCount[c] ?? 0) + 1;
        });
    });
    const topCampos = Object.entries(camposCount).sort((a, b) => b[1] - a[1]);

    return (
        <div className="bg-white dark:bg-gray-800 p-6 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm">
            <h3 className="text-sm font-bold text-gray-900 dark:text-white mb-3">
                Logs OCR — Campos ilegibles ({data.total})
            </h3>

            {topCampos.length > 0 && (
                <div className="mb-4">
                    <h4 className="text-[11px] font-bold text-gray-500 uppercase mb-2">Campos más afectados</h4>
                    <div className="flex flex-wrap gap-2">
                        {topCampos.map(([campo, n]) => (
                            <span
                                key={campo}
                                className="px-2 py-1 rounded-full bg-amber-100 text-amber-700 text-xs font-bold"
                            >
                                {campo}: {n}
                            </span>
                        ))}
                    </div>
                </div>
            )}

            {!registros.length ? (
                <p className="text-sm text-gray-400">Sin inconsistencias OCR.</p>
            ) : (
                <div className="overflow-x-auto max-h-72 overflow-y-auto">
                    <table className="w-full text-sm">
                        <thead className="sticky top-0 bg-gray-50 dark:bg-gray-900">
                            <tr className="text-[11px] uppercase text-gray-500">
                                <th className="px-3 py-2 text-left">Mesa</th>
                                <th className="px-3 py-2 text-left">Archivo</th>
                                <th className="px-3 py-2 text-left">Campos ilegibles</th>
                            </tr>
                        </thead>
                        <tbody>
                            {registros.slice(0, 100).map((r, i) => (
                                <tr key={`${r.codigoMesa}-${i}`} className="border-t border-gray-100 dark:border-gray-700">
                                    <td className="px-3 py-2 font-mono text-xs">{r.codigoMesa}</td>
                                    <td className="px-3 py-2 text-xs text-gray-500">{r.archivo}</td>
                                    <td className="px-3 py-2">
                                        <div className="flex gap-1 flex-wrap">
                                            {r.camposIlegibles?.map((c) => (
                                                <span
                                                    key={c}
                                                    className="px-1.5 py-0.5 rounded bg-red-100 text-red-700 text-[10px] font-bold"
                                                >
                                                    {c}
                                                </span>
                                            ))}
                                        </div>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            )}
        </div>
    );
};
