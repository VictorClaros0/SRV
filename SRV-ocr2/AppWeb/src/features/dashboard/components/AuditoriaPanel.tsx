import { useQuery } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { dashboardService } from "../services/dashboardService";

interface ActaObservada {
    id?: number;
    codigoActa?: number;
    estado?: string;
    observaciones?: string;
}

const PAGE_SIZE = 8;

export const AuditoriaPanel = () => {
    const { data, isLoading } = useQuery({
        queryKey: ["auditoriaOficial"],
        queryFn: async () => {
            const { call } = dashboardService.getAuditoria();
            const { data } = await call;
            return data as { total: number; actas: ActaObservada[] };
        },
    });

    const [page, setPage] = useState(1);

    const actas: ActaObservada[] = data?.actas ?? [];
    const total = data?.total ?? actas.length;

    const totalPages = Math.max(1, Math.ceil(actas.length / PAGE_SIZE));
    const safePage = Math.min(page, totalPages);

    const pageItems = useMemo(() => {
        const start = (safePage - 1) * PAGE_SIZE;
        return actas.slice(start, start + PAGE_SIZE);
    }, [actas, safePage]);

    if (isLoading)
        return <div className="h-96 bg-gray-100 dark:bg-gray-800 rounded-xl animate-pulse" />;

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
                <>
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
                                {pageItems.map((acta, i) => (
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

                    <div className="flex items-center justify-between mt-4 text-xs text-gray-500 dark:text-gray-400">
                        <span>
                            Mostrando {(safePage - 1) * PAGE_SIZE + 1}–
                            {Math.min(safePage * PAGE_SIZE, actas.length)} de {actas.length}
                        </span>
                        <div className="flex items-center gap-1">
                            <button
                                onClick={() => setPage((p) => Math.max(1, p - 1))}
                                disabled={safePage <= 1}
                                className="p-1.5 rounded-md bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 disabled:opacity-40 disabled:cursor-not-allowed"
                            >
                                <ChevronLeft size={14} />
                            </button>
                            <span className="px-2 font-semibold text-gray-700 dark:text-gray-200">
                                {safePage} / {totalPages}
                            </span>
                            <button
                                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                                disabled={safePage >= totalPages}
                                className="p-1.5 rounded-md bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 disabled:opacity-40 disabled:cursor-not-allowed"
                            >
                                <ChevronRight size={14} />
                            </button>
                        </div>
                    </div>
                </>
            )}
        </div>
    );
};