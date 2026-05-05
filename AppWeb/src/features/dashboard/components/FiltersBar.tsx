import { useQuery } from "@tanstack/react-query";
import { dashboardService } from "../services/dashboardService";
import type { FiltroItem } from "../models/response/filtros-response";

export interface DashboardFilters {
    departamento: string;
    municipio: string;
    provincia: string;
    recinto: string;
    mesa: string;
}

interface Props {
    value: DashboardFilters;
    onChange: (next: DashboardFilters) => void;
}

const useFiltro = (key: unknown[], fn: () => Promise<FiltroItem[]>, enabled = true) =>
    useQuery({ queryKey: key, queryFn: fn, enabled, staleTime: 60_000 });

export const FiltersBar = ({ value, onChange }: Props) => {
    const deps = useFiltro(["filtro-deps"], async () => {
        const { call } = dashboardService.getDepartamentos();
        return (await call).data;
    });

    const muns = useFiltro(
        ["filtro-muns", value.departamento],
        async () => {
            const { call } = dashboardService.getMunicipios(value.departamento);
            return (await call).data;
        },
        !!value.departamento,
    );

    const provs = useFiltro(
        ["filtro-provs", value.departamento, value.municipio],
        async () => {
            const { call } = dashboardService.getProvincias(value.departamento, value.municipio);
            return (await call).data;
        },
        !!value.municipio,
    );

    const recs = useFiltro(
        ["filtro-recs", value.departamento, value.municipio, value.provincia],
        async () => {
            const { call } = dashboardService.getRecintos({
                departamento: value.departamento,
                municipio: value.municipio,
                provincia: value.provincia,
            });
            return (await call).data;
        },
        !!value.provincia,
    );

    const mesas = useFiltro(
        ["filtro-mesas", value.recinto],
        async () => {
            const { call } = dashboardService.getMesas(value.recinto);
            return (await call).data;
        },
        !!value.recinto,
    );

    const set = (patch: Partial<DashboardFilters>) => onChange({ ...value, ...patch });

    return (
        <div className="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 px-5 py-4 flex flex-wrap gap-3 items-end">
            <Field label="Departamento">
                <select
                    value={value.departamento}
                    onChange={(e) =>
                        set({
                            departamento: e.target.value,
                            municipio: "",
                            provincia: "",
                            recinto: "",
                            mesa: "",
                        })
                    }
                    className={SELECT}
                >
                    <option value="">Todos</option>
                    {deps.data?.map((d) => (
                        <option key={String(d.id)} value={d.nombre}>{d.nombre}</option>
                    ))}
                </select>
            </Field>

            <Field label="Municipio">
                <select
                    value={value.municipio}
                    onChange={(e) =>
                        set({ municipio: e.target.value, provincia: "", recinto: "", mesa: "" })
                    }
                    disabled={!value.departamento}
                    className={SELECT}
                >
                    <option value="">{value.departamento ? "Todos" : "Selecciona depto."}</option>
                    {muns.data?.map((m) => (
                        <option key={String(m.id)} value={m.nombre}>{m.nombre}</option>
                    ))}
                </select>
            </Field>

            <Field label="Provincia">
                <select
                    value={value.provincia}
                    onChange={(e) => set({ provincia: e.target.value, recinto: "", mesa: "" })}
                    disabled={!value.municipio}
                    className={SELECT}
                >
                    <option value="">{value.municipio ? "Todas" : "Selecciona municipio"}</option>
                    {provs.data?.map((p) => (
                        <option key={String(p.id)} value={p.nombre}>{p.nombre}</option>
                    ))}
                </select>
            </Field>

            <Field label="Recinto">
                <select
                    value={value.recinto}
                    onChange={(e) => set({ recinto: e.target.value, mesa: "" })}
                    disabled={!value.provincia}
                    className={SELECT}
                >
                    <option value="">{value.provincia ? "Todos" : "Selecciona provincia"}</option>
                    {recs.data?.map((r) => (
                        <option key={String(r.id)} value={String(r.id)}>{r.nombre}</option>
                    ))}
                </select>
            </Field>

            <Field label="Mesa">
                <select
                    value={value.mesa}
                    onChange={(e) => set({ mesa: e.target.value })}
                    disabled={!value.recinto}
                    className={SELECT}
                >
                    <option value="">{value.recinto ? "Todas" : "Selecciona recinto"}</option>
                    {mesas.data?.map((m) => (
                        <option key={String(m.id)} value={String(m.id)}>{m.nombre}</option>
                    ))}
                </select>
            </Field>

            <button
                onClick={() =>
                    onChange({ departamento: "", municipio: "", provincia: "", recinto: "", mesa: "" })
                }
                className="bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-200 text-sm px-4 py-2 rounded-md"
            >
                Limpiar
            </button>
        </div>
    );
};

const SELECT =
    "bg-white dark:bg-gray-900 border border-gray-300 dark:border-gray-700 rounded-md px-3 py-1.5 text-sm outline-none focus:ring-2 focus:ring-blue-500 min-w-44 disabled:opacity-50 disabled:cursor-not-allowed";

const Field = ({ label, children }: { label: string; children: React.ReactNode }) => (
    <div className="flex flex-col gap-1">
        <label className="text-[11px] font-bold text-gray-500 dark:text-gray-400 uppercase tracking-wide">
            {label}
        </label>
        {children}
    </div>
);
