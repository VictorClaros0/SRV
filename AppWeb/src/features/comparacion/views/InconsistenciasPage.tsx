import { useEffect, useState } from "react";
import { InconsistenciasTable } from "../components/InconsistenciasTable";
import { comparacionService } from "../services/comparacionService";
import type { AuditoriaOficialResponse } from "../models/response/auditoria-oficial-response";


export const InconsistenciasPage = () => {

    const [data, setData] = useState<AuditoriaOficialResponse | null>(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const load = async () => {
        setLoading(true);
        setError(null);
        try {


            const { call } = comparacionService.list();
            const { data: res } = await call;
            setData(res);
        } catch (e: unknown) {
            const msg = e instanceof Error ? e.message : "Error al cargar datos.";
            setError(msg);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        load();
    }, []);


    return (
        <div className="p-6 md:p-8">
            <div className="max-w-7xl mx-auto space-y-5">
                <h1 className="text-2xl font-extrabold text-gray-900 dark:text-white tracking-tight">
                    Inconsistencias
                </h1>

                {loading && (
                    <div className="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm p-10 text-center text-gray-500">
                        Cargando
                    </div>
                )}

                {!loading && error && (
                    <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-300 rounded-xl p-4 text-sm">
                        {error}
                    </div>
                )}

                {!loading && !error && (
                    <InconsistenciasTable
                        rows={data?.actas ?? []}
                    />
                )}
            </div>
        </div>
    );
};

export default InconsistenciasPage;
