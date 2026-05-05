import { useState } from "react";
import { httpClientRrv } from "../../../config/axios-rrv";
import { Search } from "lucide-react";

interface TrazabilidadData {
    acta?: Record<string, unknown>;
    rrvActa?: Record<string, unknown>;
    eventos?: Array<{
        _id?: string;
        tipo?: string;
        fecha?: string;
        timestamp?: string;
        actaId?: string;
        payload?: unknown;
        [k: string]: unknown;
    }>;
}

export const TrazabilidadPanel = () => {
    const [codigo, setCodigo] = useState("");
    const [data, setData] = useState<TrazabilidadData | null>(null);
    const [loading, setLoading] = useState(false);
    const [err, setErr] = useState<string | null>(null);

    const buscar = async () => {
        if (!codigo.trim()) return;
        setLoading(true);
        setErr(null);
        try {
            const res = await httpClientRrv.get<TrazabilidadData>(
                `dashboard/trazabilidad/${codigo.trim()}`,
            );
            setData(res.data);
        } catch (e) {
            setErr("Acta no encontrada o error al consultar");
            setData(null);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="bg-white dark:bg-gray-800 p-6 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm">
            <h3 className="text-sm font-bold text-gray-900 dark:text-white mb-3">
                Trazabilidad de un acta (SRRV)
            </h3>
            <div className="flex gap-2 mb-3">
                <div className="relative flex-1">
                    <Search size={14} className="absolute left-2 top-2.5 text-gray-400" />
                    <input
                        value={codigo}
                        onChange={(e) => setCodigo(e.target.value)}
                        onKeyDown={(e) => e.key === "Enter" && buscar()}
                        placeholder="Código de mesa/acta (ej: 1060200261005)"
                        className="w-full bg-white dark:bg-gray-900 border border-gray-300 dark:border-gray-700 text-sm rounded-md pl-7 pr-3 py-2"
                    />
                </div>
                <button
                    onClick={buscar}
                    disabled={loading || !codigo.trim()}
                    className="bg-gray-900 hover:bg-black text-white text-sm font-semibold px-5 py-2 rounded-md disabled:opacity-50"
                >
                    {loading ? "..." : "Buscar"}
                </button>
            </div>

            {err && <p className="text-sm text-red-500">{err}</p>}

            {data && (
                <div className="space-y-3">
                    <div className="grid grid-cols-2 gap-3">
                        <div className="bg-orange-50 dark:bg-orange-900/20 rounded-lg p-3">
                            <div className="text-[11px] uppercase font-bold text-orange-700">RRV</div>
                            <pre className="text-[10px] mt-1 overflow-auto max-h-40">
                                {JSON.stringify(data.rrvActa ?? "—", null, 2)}
                            </pre>
                        </div>
                        <div className="bg-blue-50 dark:bg-blue-900/20 rounded-lg p-3">
                            <div className="text-[11px] uppercase font-bold text-blue-700">Acta base</div>
                            <pre className="text-[10px] mt-1 overflow-auto max-h-40">
                                {JSON.stringify(data.acta ?? "—", null, 2)}
                            </pre>
                        </div>
                    </div>

                    <div>
                        <h4 className="text-xs font-bold text-gray-500 uppercase mb-2">
                            Timeline de eventos ({(data.eventos ?? []).length})
                        </h4>
                        <div className="space-y-1 max-h-60 overflow-auto">
                            {(data.eventos ?? []).map((e, i) => (
                                <div
                                    key={(e._id as string) ?? i}
                                    className="border-l-2 border-emerald-500 pl-3 py-1 text-xs"
                                >
                                    <div className="font-bold text-gray-900 dark:text-white">{e.tipo}</div>
                                    <div className="text-gray-500">
                                        {(e.fecha as string) ?? (e.timestamp as string) ?? "—"}
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};
