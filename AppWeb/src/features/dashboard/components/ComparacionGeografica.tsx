import { useQueries } from "@tanstack/react-query";
import { rrvService } from "../services/rrvService";
import { dashboardService } from "../services/dashboardService";
import type { RrvGeograficoResponse } from "../models/response/rrv-responses";
import type { GeograficoResponse } from "../models/response/geografico-response";

interface Row {
    nombre: string;
    rrvValidos: number;
    oficialValidos: number;
    delta: number;
    deltaPct: number;
    estado: "CONSISTENTE" | "INCONSISTENTE" | "SOLO_RRV" | "SOLO_OFICIAL";
}

const calcRow = (nombre: string, rrv: number, oficial: number): Row => {
    const delta = rrv - oficial;
    const base = Math.max(rrv, oficial);
    const deltaPct = base === 0 ? 0 : (Math.abs(delta) / base) * 100;
    let estado: Row["estado"] = "CONSISTENTE";
    if (rrv === 0 && oficial > 0) estado = "SOLO_OFICIAL";
    else if (oficial === 0 && rrv > 0) estado = "SOLO_RRV";
    else if (deltaPct > 1) estado = "INCONSISTENTE";
    return {
        nombre,
        rrvValidos: rrv,
        oficialValidos: oficial,
        delta,
        deltaPct: Math.round(deltaPct * 100) / 100,
        estado,
    };
};

const badge: Record<Row["estado"], string> = {
    CONSISTENTE: "bg-green-100 text-green-700",
    INCONSISTENTE: "bg-red-100 text-red-700",
    SOLO_RRV: "bg-orange-100 text-orange-700",
    SOLO_OFICIAL: "bg-slate-100 text-slate-700",
};

export const ComparacionGeografica = () => {
    const [rrvQ, oficialQ] = useQueries({
        queries: [
            {
                queryKey: ["rrv-geo-comp"],
                queryFn: async () => {
                    const { call } = rrvService.getGeografico("departamento");
                    return (await call).data as RrvGeograficoResponse;
                },
            },
            {
                queryKey: ["oficial-geo-comp"],
                queryFn: async () => {
                    const { call } = dashboardService.getGeografico("departamento");
                    return (await call).data as GeograficoResponse;
                },
            },
        ],
    });

    if (rrvQ.isLoading || oficialQ.isLoading)
        return <div className="h-64 bg-gray-100 dark:bg-gray-800 rounded-xl animate-pulse" />;

    const rrvItems = rrvQ.data?.datos ?? [];
    const oficialItems = oficialQ.data?.items ?? [];

    const oficialMap: Record<string, number> = {};
    oficialItems.forEach((i) => {
        oficialMap[i.nombre] = i.votosValidos;
    });

    const todos = new Set<string>();
    rrvItems.forEach((i) => todos.add(i.nombre));
    oficialItems.forEach((i) => todos.add(i.nombre));

    const rows: Row[] = Array.from(todos).map((nombre) => {
        const rrvI = rrvItems.find((i) => i.nombre === nombre);
        return calcRow(nombre, rrvI?.votosValidos ?? 0, oficialMap[nombre] ?? 0);
    });

    rows.sort((a, b) => Math.abs(b.delta) - Math.abs(a.delta));

    const totales = rows.reduce(
        (acc, r) => {
            acc[r.estado]++;
            return acc;
        },
        { CONSISTENTE: 0, INCONSISTENTE: 0, SOLO_RRV: 0, SOLO_OFICIAL: 0 } as Record<Row["estado"], number>,
    );

    return (
        <div className="bg-white dark:bg-gray-800 p-6 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm">
            <h3 className="text-sm font-bold text-gray-900 dark:text-white mb-3">
                Inconsistencias por departamento (RRV vs Oficial)
            </h3>
            <div className="flex gap-2 mb-3 text-xs">
                <span className={`px-2 py-1 rounded-full ${badge.CONSISTENTE}`}>Consist. {totales.CONSISTENTE}</span>
                <span className={`px-2 py-1 rounded-full ${badge.INCONSISTENTE}`}>Inconsist. {totales.INCONSISTENTE}</span>
                <span className={`px-2 py-1 rounded-full ${badge.SOLO_RRV}`}>Solo RRV {totales.SOLO_RRV}</span>
                <span className={`px-2 py-1 rounded-full ${badge.SOLO_OFICIAL}`}>Solo Oficial {totales.SOLO_OFICIAL}</span>
            </div>
            <div className="overflow-x-auto">
                <table className="w-full text-sm">
                    <thead>
                        <tr className="bg-gray-50 dark:bg-gray-900/50 text-[11px] uppercase text-gray-500">
                            <th className="px-3 py-2 text-left">Departamento</th>
                            <th className="px-3 py-2 text-left">RRV</th>
                            <th className="px-3 py-2 text-left">Oficial</th>
                            <th className="px-3 py-2 text-left">Δ</th>
                            <th className="px-3 py-2 text-left">Δ %</th>
                            <th className="px-3 py-2 text-left">Estado</th>
                        </tr>
                    </thead>
                    <tbody>
                        {rows.map((r) => (
                            <tr key={r.nombre} className="border-t border-gray-100 dark:border-gray-700">
                                <td className="px-3 py-2 font-semibold">{r.nombre}</td>
                                <td className="px-3 py-2 text-orange-500 font-semibold">{r.rrvValidos.toLocaleString("es-BO")}</td>
                                <td className="px-3 py-2 text-blue-500 font-semibold">{r.oficialValidos.toLocaleString("es-BO")}</td>
                                <td className={`px-3 py-2 font-semibold ${r.delta >= 0 ? "text-blue-500" : "text-red-500"}`}>
                                    {r.delta >= 0 ? "+" : ""}{r.delta.toLocaleString("es-BO")}
                                </td>
                                <td className="px-3 py-2">{r.deltaPct.toFixed(2)}%</td>
                                <td className="px-3 py-2">
                                    <span className={`text-[10px] px-2 py-0.5 rounded-full font-bold ${badge[r.estado]}`}>
                                        {r.estado}
                                    </span>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </div>
    );
};
