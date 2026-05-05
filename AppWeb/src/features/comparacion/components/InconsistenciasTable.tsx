import type { DataActaAudit } from "../models/response/auditoria-oficial-response";

const ESTADO_COLOR: Record<string, string> = {
    CONSISTENTE: "#22c55e",
    INCONSISTENTE: "#ef4444",
    SOLO_RRV: "#f97316",
    SOLO_OFICIAL: "#94a3b8",
};


const Pill = ({ value, color }: { value: string; color: string }) => (
    <span
        className="inline-block px-2 py-0.5 rounded-full text-xs font-bold"
        style={{ backgroundColor: `${color}20`, color }}
    >
        {value}
    </span>
);

interface Props {
    rows: DataActaAudit[];

}

export const InconsistenciasTable = ({
    rows,
}: Props) => {


    return (
        <div className="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm overflow-hidden">
            <div className="overflow-x-auto">
                <table className="w-full text-sm">
                    <thead>
                        <tr className="bg-gray-50 dark:bg-gray-900/50 text-[11px] uppercase tracking-wide text-gray-500 dark:text-gray-400">
                            <th className="px-4 py-3 text-left font-bold">Acta ID</th>
                            <th className="px-4 py-3 text-left font-bold">Codigo Acta</th>
                            <th className="px-4 py-3 text-left font-bold">Codigo Recinto</th>
                            <th className="px-4 py-3 text-left font-bold">Nro Mesa</th>
                            <th className="px-4 py-3 text-left font-bold">Estado</th>
                            <th className="px-4 py-3 text-left font-bold">Observaciones</th>
                            <th className="px-4 py-3 text-left font-bold">Fecha de creacion</th>
                        </tr>
                    </thead>
                    <tbody>
                        {rows.map((row, i) => {
                            const color = ESTADO_COLOR[row.estado] ?? "#888";
                            return (
                                <tr
                                    key={row.id || i}
                                    className="border-t border-gray-100 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-900/30"
                                >
                                    <td className="px-4 py-3 font-semibold text-gray-900 dark:text-white">
                                        {row.id}
                                    </td>
                                    <td className="px-4 py-3 text-gray-700 dark:text-gray-300">
                                        {row.codigoActa || "—"}
                                    </td>
                                    <td className="px-4 py-3 text-gray-700 dark:text-gray-300">
                                        {row.codigoRecinto || "—"}
                                    </td>

                                    <td className="px-4 py-3 text-gray-700 dark:text-gray-300">
                                        {row.nroMesa || "—"}
                                    </td>
                                    <td className="px-4 py-3">
                                        <Pill value={row.estado} color={color} />
                                    </td>
                                    <td className="px-4 py-3 text-gray-700 dark:text-gray-300">
                                        {row.observaciones || "—"}
                                    </td>
                                    <td className="px-4 py-3 text-gray-700 dark:text-gray-300">
                                        {row.fechaCreacion || "—"}
                                    </td>
                                </tr>
                            );
                        })}
                    </tbody>
                </table>
            </div>


        </div>
    );
};
