import { useQuery } from "@tanstack/react-query";
import { rrvService } from "../services/rrvService";

const Item = ({ label, value, color }: { label: string; value: string | number; color: string }) => (
    <div className="bg-gray-50 dark:bg-gray-900/50 rounded-lg px-3 py-2 border border-gray-100 dark:border-gray-700">
        <div className="text-[10px] font-bold text-gray-500 uppercase tracking-wide">{label}</div>
        <div className={`text-lg font-extrabold ${color}`}>{value}</div>
    </div>
);

export const TecnicoPanel = () => {
    const { data, isLoading } = useQuery({
        queryKey: ["rrv-tecnico"],
        queryFn: async () => (await rrvService.getTecnico().call).data,
    });

    if (isLoading) return <div className="h-64 bg-gray-100 dark:bg-gray-800 rounded-xl animate-pulse" />;
    if (!data) return null;

    const eventos = data.eventosPorTipo ?? {};

    return (
        <div className="bg-white dark:bg-gray-800 p-6 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm">
            <h3 className="text-sm font-bold text-gray-900 dark:text-white mb-3">
                Indicadores técnicos (SRRV / OCR)
            </h3>
            <div className="grid grid-cols-2 md:grid-cols-3 gap-3 mb-4">
                <Item label="Total RRV actas" value={data.totalRRVActas?.toLocaleString("es-BO") ?? "—"} color="text-orange-500" />
                <Item label="Procesadas" value={data.totalActasProcesadas?.toLocaleString("es-BO") ?? "—"} color="text-green-500" />
                <Item label="Inconsistentes" value={data.totalActasInconsistentes?.toLocaleString("es-BO") ?? "—"} color="text-red-500" />
                <Item label="Total eventos" value={data.totalEventos?.toLocaleString("es-BO") ?? "—"} color="text-indigo-500" />
                <Item label="Últimas 24h" value={data.actasUltimas24h?.toLocaleString("es-BO") ?? "—"} color="text-blue-500" />
                <Item label="Disponibilidad" value="99.9%" color="text-emerald-500" />
            </div>
            <h4 className="text-[11px] font-bold text-gray-500 uppercase mb-2">Eventos por tipo</h4>
            <div className="grid grid-cols-2 md:grid-cols-3 gap-2">
                {Object.entries(eventos).length === 0 ? (
                    <p className="text-xs text-gray-400">Sin eventos registrados.</p>
                ) : (
                    Object.entries(eventos).map(([tipo, n]) => (
                        <div key={tipo} className="flex justify-between bg-gray-50 dark:bg-gray-900/30 rounded px-2 py-1 text-xs">
                            <span className="text-gray-700 dark:text-gray-300 truncate">{tipo}</span>
                            <span className="font-bold text-gray-900 dark:text-white">
                                {Number(n).toLocaleString("es-BO")}
                            </span>
                        </div>
                    ))
                )}
            </div>
        </div>
    );
};
