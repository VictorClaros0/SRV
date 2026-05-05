import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { dashboardService } from "../services/dashboardService";
import type { GeoNivel } from "../models/request/geografico-request";
import type { DashboardFilters } from "./FiltersBar";

interface Props {
    filters: DashboardFilters;
}

export const GeograficoPanel = ({ filters }: Props) => {
    const [nivel, setNivel] = useState<GeoNivel>("departamento");

    const { data, isLoading } = useQuery({
        queryKey: ["geografico", nivel, filters.departamento, filters.municipio, filters.provincia],
        queryFn: async () => {
            const { call } = dashboardService.getGeografico(nivel, {
                departamento: filters.departamento,
                municipio: filters.municipio,
                provincia: filters.provincia,
            });
            const { data } = await call;
            return data;
        },
    });

    return (
        <div className="bg-white dark:bg-gray-800 p-6 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm">
            <div className="flex items-center justify-between mb-4 gap-3">
                <h3 className="text-sm font-bold text-gray-900 dark:text-white">Distribución territorial</h3>
                <select
                    value={nivel}
                    onChange={(e) => setNivel(e.target.value as GeoNivel)}
                    className="bg-white dark:bg-gray-900 border border-gray-300 dark:border-gray-700 text-sm rounded-md px-3 py-1.5"
                >
                    <option value="departamento">Departamento</option>
                    <option value="municipio">Municipio</option>
                    <option value="provincia">Provincia</option>
                    <option value="recinto">Recinto</option>
                </select>
            </div>

            {isLoading ? (
                <div className="h-48 bg-gray-100 dark:bg-gray-900 rounded-lg animate-pulse" />
            ) : !data?.items?.length ? (
                <p className="text-sm text-gray-400 py-6 text-center">Sin datos.</p>
            ) : (
                <div className="overflow-x-auto">
                    <table className="w-full text-sm">
                        <thead>
                            <tr className="bg-gray-50 dark:bg-gray-900/50 text-[11px] uppercase tracking-wide text-gray-500">
                                <th className="px-3 py-2 text-left">{nivel}</th>
                                <th className="px-3 py-2 text-left">Actas</th>
                                <th className="px-3 py-2 text-left">P1</th>
                                <th className="px-3 py-2 text-left">P2</th>
                                <th className="px-3 py-2 text-left">P3</th>
                                <th className="px-3 py-2 text-left">P4</th>
                                <th className="px-3 py-2 text-left">Válidos</th>
                                <th className="px-3 py-2 text-left">Nulos</th>
                                <th className="px-3 py-2 text-left">Blanco</th>
                                <th className="px-3 py-2 text-left">Ganador</th>
                                <th className="px-3 py-2 text-left">Margen</th>
                            </tr>
                        </thead>
                        <tbody>
                            {data.items.map((it, idx) => (
                                <tr key={`${it.nombre}-${idx}`} className="border-t border-gray-100 dark:border-gray-700">
                                    <td className="px-3 py-2 font-semibold text-gray-900 dark:text-white">{it.nombre}</td>
                                    <td className="px-3 py-2">{it.totalActas}</td>
                                    <td className="px-3 py-2">{it.p1.toLocaleString("es-BO")} <span className="text-[10px] text-gray-400">({it.p1_pct.toFixed(1)}%)</span></td>
                                    <td className="px-3 py-2">{it.p2.toLocaleString("es-BO")} <span className="text-[10px] text-gray-400">({it.p2_pct.toFixed(1)}%)</span></td>
                                    <td className="px-3 py-2">{it.p3.toLocaleString("es-BO")} <span className="text-[10px] text-gray-400">({it.p3_pct.toFixed(1)}%)</span></td>
                                    <td className="px-3 py-2">{it.p4.toLocaleString("es-BO")} <span className="text-[10px] text-gray-400">({it.p4_pct.toFixed(1)}%)</span></td>
                                    <td className="px-3 py-2">{it.votosValidos.toLocaleString("es-BO")}</td>
                                    <td className="px-3 py-2">{it.votosNulos.toLocaleString("es-BO")}</td>
                                    <td className="px-3 py-2">{it.votosBlanco.toLocaleString("es-BO")}</td>
                                    <td className="px-3 py-2 font-bold text-emerald-600">{it.ganador}</td>
                                    <td className="px-3 py-2">{it.margen?.toFixed(2)}%</td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            )}
        </div>
    );
};
