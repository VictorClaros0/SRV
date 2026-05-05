import { useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { RefreshCcw } from "lucide-react";
import { KPICards } from "../components/KPICards";
import { VotosCandidatoChart } from "../components/VotosCandidatoChart";
import { HeatmapPanel } from "../components/HeatmapPanel";
import { GeograficoPanel } from "../components/GeograficoPanel";
import { ParticipacionChart } from "../components/ParticipacionChart";
import { AnomaliasPanel } from "../components/AnomaliasPanel";
import { AuditoriaPanel } from "../components/AuditoriaPanel";
import { VelocidadPanel } from "../components/VelocidadPanel";
import { FiltersBar, type DashboardFilters } from "../components/FiltersBar";
import { ComparacionKPIs } from "../components/ComparacionKPIs";
import { ComparacionVotosCandidato } from "../components/ComparacionVotosCandidato";
import { ComparacionGeografica } from "../components/ComparacionGeografica";
import { HeatmapComparativoPanel } from "../components/HeatmapComparativoPanel";
import { TransparenciaPanel } from "../components/TransparenciaPanel";
import { TrazabilidadPanel } from "../components/TrazabilidadPanel";
import { TecnicoPanel } from "../components/TecnicoPanel";
import { AnomaliasRRVPanel } from "../components/AnomaliasRRVPanel";
import { LogsInconsistenciasPanel } from "../components/LogsInconsistenciasPanel";
import { ParticipacionComparativa } from "../components/ParticipacionComparativa";

const DEFAULT_FILTERS: DashboardFilters = {
    departamento: "",
    municipio: "",
    provincia: "",
    recinto: "",
    mesa: "",
};

export const DashboardPage = () => {
    const queryClient = useQueryClient();
    const [lastUpdate, setLastUpdate] = useState<string>(() => new Date().toLocaleTimeString());
    const [refreshing, setRefreshing] = useState(false);
    const [filters, setFilters] = useState<DashboardFilters>(DEFAULT_FILTERS);

    const refresh = async () => {
        setRefreshing(true);
        await queryClient.invalidateQueries();
        setLastUpdate(new Date().toLocaleTimeString());
        setTimeout(() => setRefreshing(false), 600);
    };

    return (
        <div className="p-6 md:p-8">
            <div className="max-w-7xl mx-auto space-y-5">
                <div className="flex items-center gap-3">
                    <div className="flex-1">
                        <h1 className="text-2xl font-extrabold text-gray-900 dark:text-white tracking-tight">
                            Dashboard Electoral — RRV + Cómputo Oficial
                        </h1>
                        <p className="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
                            Actualizado: {lastUpdate}
                        </p>
                    </div>
                    <button
                        onClick={refresh}
                        className="inline-flex items-center gap-2 bg-gray-100 dark:bg-gray-800 hover:bg-gray-200 dark:hover:bg-gray-700 active:scale-95 text-gray-700 dark:text-gray-200 text-sm font-medium px-4 py-2 rounded-lg transition-all"
                    >
                        <RefreshCcw size={14} className={refreshing ? "animate-spin" : ""} />
                        Actualizar
                    </button>
                </div>

                <FiltersBar value={filters} onChange={setFilters} />

                <KPICards />

                <ComparacionKPIs />

                <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
                    <VotosCandidatoChart filters={filters} />
                    <ParticipacionChart filters={filters} />
                </div>

                <ComparacionVotosCandidato />

                <ParticipacionComparativa />

                <HeatmapComparativoPanel />

                <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
                    <HeatmapPanel />
                    <VelocidadPanel />
                </div>

                <ComparacionGeografica />

                <GeograficoPanel filters={filters} />

                <TransparenciaPanel />

                <TrazabilidadPanel />

                <TecnicoPanel />

                <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
                    <AnomaliasPanel />
                    <AnomaliasRRVPanel />
                </div>

                <LogsInconsistenciasPanel />

                <AuditoriaPanel />
            </div>
        </div>
    );
};

export default DashboardPage;
