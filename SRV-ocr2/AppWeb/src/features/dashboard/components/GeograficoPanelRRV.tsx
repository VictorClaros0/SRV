import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { rrvService } from "../services/rrvService";
import type { DashboardFilters } from "./FiltersBar";

type Nivel = "departamento" | "provincia" | "municipio" | "recinto";

interface Props {
    filters: DashboardFilters;
}

export const GeograficoPanelRRV = ({ filters }: Props) => {
    const [nivel, setNivel] = useState<Nivel>("departamento");

    const { data, isLoading } = useQuery({
        queryKey: ["rrv-geografico", nivel, filters.departamento, filters.municipio, filters.provincia, filters.recinto, filters.mesa],
        queryFn: async () => {
            const { call } = rrvService.getGeografico(nivel, {
                departamento: filters.departamento,
                municipio: filters.municipio,
                provincia: filters.provincia,
                recinto: filters.recinto,
                mesa: filters.mesa,
            });
            const { data } = await call;
            return data;
        },
    });

    return (
        <div className="bg-white dark:bg-gray-800 p-6 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm">
            <div className="flex items-center justify-between mb-4 gap-3">
                <h3 className="text-sm font-bold text-gray-900 dark:text-white">
                    Distribución territorial (conteo rápido)
                </h3>
                <select
                    value={nivel}
                    onChange={(e) => setNivel(e.target.value as Nivel)}
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
            ) : !data?.datos?.length ? (
                <p className="text-sm text-gray-400 py-6 text-center">Sin datos.</p>
            ) : (
                <div className="overflow-x-auto">
                    <table className="w-full text-sm">
                        <thead>
                            <tr className="bg-gray-50 dark:bg-gray-900/50 text-[11px] uppercase tracking-wide text-gray-500">
                                <th className="px-3 py-2 text-left">{nivel}</th>
                                <th className="px-3 py-2 text-left">P1</th>
                                <th className="px-3 py-2 text-left">P2</th>
                                <th className="px-3 py-2 text-left">P3</th>
                                <th className="px-3 py-2 text-left">P4</th>
                                <th className="px-3 py-2 text-left">Válidos</th>
                                <th className="px-3 py-2 text-left">Nulos</th>
                                <th className="px-3 py-2 text-left">Habilitados</th>
                                <th className="px-3 py-2 text-left">Participación</th>
                            </tr>
                        </thead>
                        <tbody>
                            {data.datos.map((it, idx) => (
                                <tr key={`${it.nombre}-${idx}`} className="border-t border-gray-100 dark:border-gray-700">
                                    <td className="px-3 py-2 font-semibold text-gray-900 dark:text-white">{it.nombre}</td>
                                    <td className="px-3 py-2">{it.votosP1.toLocaleString("es-BO")}</td>
                                    <td className="px-3 py-2">{it.votosP2.toLocaleString("es-BO")}</td>
                                    <td className="px-3 py-2">{it.votosP3.toLocaleString("es-BO")}</td>
                                    <td className="px-3 py-2">{it.votosP4.toLocaleString("es-BO")}</td>
                                    <td className="px-3 py-2">{it.votosValidos.toLocaleString("es-BO")}</td>
                                    <td className="px-3 py-2">{it.votosNulos.toLocaleString("es-BO")}</td>
                                    <td className="px-3 py-2">{it.habilitados.toLocaleString("es-BO")}</td>
                                    <td className="px-3 py-2 font-bold text-emerald-600">{it.participacion?.toFixed(2)}%</td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            )}
        </div>
    );
};