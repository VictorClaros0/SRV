import type { ComparacionListParams } from "../models/request/comparacion-list-request";
import { DEPARTAMENTOS, getMunicipiosByDepartamento } from "../data/territorios";

const ESTADOS = ["", "CONSISTENTE", "INCONSISTENTE", "SOLO_RRV", "SOLO_OFICIAL"] as const;
const PER_PAGE = [10, 20, 50, 100];

interface Props {
    filters: ComparacionListParams;
    onChange: (next: ComparacionListParams) => void;
    onApply: () => void;
    onClear: () => void;
}

export const FiltersBar = ({ filters, onChange, onApply, onClear }: Props) => {
    const handleEstado = (e: React.ChangeEvent<HTMLSelectElement>) =>
        onChange({ ...filters, estado: e.target.value, pagina: 1 });

    const handleDepto = (e: React.ChangeEvent<HTMLSelectElement>) =>
        // Al cambiar departamento, limpiar municipio (cascada)
        onChange({ ...filters, departamento: e.target.value, municipio: "", pagina: 1 });

    const handleMunicipio = (e: React.ChangeEvent<HTMLSelectElement>) =>
        onChange({ ...filters, municipio: e.target.value, pagina: 1 });

    const handleRecinto = (e: React.ChangeEvent<HTMLInputElement>) =>
        onChange({ ...filters, recinto: e.target.value, pagina: 1 });

    const handlePerPage = (e: React.ChangeEvent<HTMLSelectElement>) =>
        onChange({ ...filters, por_pagina: Number(e.target.value), pagina: 1 });

    const municipios = getMunicipiosByDepartamento(filters.departamento ?? "");

    return (
        <div className="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 px-5 py-4 flex flex-wrap gap-3 items-end mb-4">
            <Field label="Estado">
                <select
                    value={filters.estado ?? ""}
                    onChange={handleEstado}
                    className="bg-white dark:bg-gray-900 border border-gray-300 dark:border-gray-700 rounded-md px-3 py-1.5 text-sm outline-none focus:ring-2 focus:ring-blue-500 min-w-36"
                >
                    {ESTADOS.map((e) => (
                        <option key={e} value={e}>
                            {e || "Todos"}
                        </option>
                    ))}
                </select>
            </Field>

            <Field label="Departamento">
                <select
                    value={filters.departamento ?? ""}
                    onChange={handleDepto}
                    className="bg-white dark:bg-gray-900 border border-gray-300 dark:border-gray-700 rounded-md px-3 py-1.5 text-sm outline-none focus:ring-2 focus:ring-blue-500 min-w-44"
                >
                    <option value="">Todos</option>
                    {DEPARTAMENTOS.map((d) => (
                        <option key={d} value={d}>
                            {d}
                        </option>
                    ))}
                </select>
            </Field>

            <Field label="Municipio">
                <select
                    value={filters.municipio ?? ""}
                    onChange={handleMunicipio}
                    disabled={!filters.departamento}
                    className="bg-white dark:bg-gray-900 border border-gray-300 dark:border-gray-700 rounded-md px-3 py-1.5 text-sm outline-none focus:ring-2 focus:ring-blue-500 min-w-44 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                    <option value="">{filters.departamento ? "Todos" : "Selecciona depto."}</option>
                    {municipios.map((m) => (
                        <option key={m} value={m}>
                            {m}
                        </option>
                    ))}
                </select>
            </Field>

            <Field label="Recinto">
                <input
                    value={filters.recinto ?? ""}
                    onChange={handleRecinto}
                    placeholder="Ej. Colegio Ayacucho"
                    className="bg-white dark:bg-gray-900 border border-gray-300 dark:border-gray-700 rounded-md px-3 py-1.5 text-sm outline-none focus:ring-2 focus:ring-blue-500 min-w-44"
                />
            </Field>

            <Field label="Resultados por página">
                <select
                    value={filters.por_pagina ?? 20}
                    onChange={handlePerPage}
                    className="bg-white dark:bg-gray-900 border border-gray-300 dark:border-gray-700 rounded-md px-3 py-1.5 text-sm outline-none focus:ring-2 focus:ring-blue-500 min-w-24"
                >
                    {PER_PAGE.map((n) => (
                        <option key={n} value={n}>
                            {n}
                        </option>
                    ))}
                </select>
            </Field>

            <button
                onClick={onApply}
                className="bg-gray-900 hover:bg-black text-white text-sm font-semibold px-5 py-2 rounded-md"
            >
                Filtrar
            </button>
            <button
                onClick={onClear}
                className="bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-200 text-sm px-4 py-2 rounded-md"
            >
                Limpiar
            </button>
        </div>
    );
};

const Field = ({ label, children }: { label: string; children: React.ReactNode }) => (
    <div className="flex flex-col gap-1">
        <label className="text-[11px] font-bold text-gray-500 dark:text-gray-400 uppercase tracking-wide">
            {label}
        </label>
        {children}
    </div>
);
