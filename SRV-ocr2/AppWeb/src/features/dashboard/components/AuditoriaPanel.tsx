import { useQuery } from "@tanstack/react-query";
import { dashboardService } from "../services/dashboardService";

interface ActaObservada {
    id?: number;
    codigoActa?: number;
    estado?: string;
    observaciones?: string;
}

export const AuditoriaPanel = () => {
    const { data, isLoading } = useQuery({
        queryKey: ["auditoriaOficial"],
        queryFn: async () => {
            const { call } = dashboardService.getAuditoria();
            const { data } = await call;
            return data as { total: number; actas: ActaObservada[] };
        },
    });

    if (isLoading)
        return <div className="h-96 bg-gray-100 dark:bg-gray-800 rounded-xl animate-pulse" />;

    const actas: ActaObservada[] = data?.actas ?? [];
    const total = data?.total ?? actas.length;

    return (
        <div className="bg-white dark:bg-gray-800 p-6 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm hover:shadow-md transition-shadow">
            <h3 className="text-sm font-bold text-gray-900 dark:text-white mb-4">
                Auditoría / Trazabilidad oficial
            </h3>

            <div className="flex items-baseline gap-2 mb-4">
                <span className="text-3xl font-extrabold text-red-500">{total.toLocaleString("es-BO")}</span>
                <span className="text-xs text-gray-500 dark:text-gray-400">actas observadas</span>
            </div>

            {actas.length === 0 ? (
                <p className="text-sm text-gray-400">Sin actas observadas.</p>
            ) : (
                <div className="overflow-x-auto">
                    <table className="w-full text-sm">
                        <thead>
                            <tr className="bg-gray-50 dark:bg-gray-900/50 text-[11px] uppercase tracking-wide text-gray-500 dark:text-gray-400">
                                <th className="px-3 py-2 text-left font-bold">Acta</th>
                                <th className="px-3 py-2 text-left font-bold">Estado</th>
                                <th className="px-3 py-2 text-left font-bold">Observaciones</th>
                            </tr>
                        </thead>
                        <tbody>
                            {actas.slice(0, 8).map((acta, i) => (
                                <tr
                                    key={acta.id ?? acta.codigoActa ?? i}
                                    className="border-t border-gray-100 dark:border-gray-700"
                                >
                                    <td className="px-3 py-2 font-semibold text-gray-900 dark:text-white">
                                        {acta.codigoActa}
                                    </td>
                                    <td className="px-3 py-2 text-gray-700 dark:text-gray-300">
                                        {acta.estado}
                                    </td>
                                    <td className="px-3 py-2 text-gray-700 dark:text-gray-300">
                                        {acta.observaciones || "—"}
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
