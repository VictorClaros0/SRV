import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { rrvService } from "../services/rrvService";
import { ExternalLink, Search } from "lucide-react";

export const TransparenciaPanel = () => {
    const [filtro, setFiltro] = useState("");

    const { data, isLoading } = useQuery({
        queryKey: ["rrv-transparencia"],
        queryFn: async () => (await rrvService.getTransparencia().call).data,
    });

    if (isLoading)
        return <div className="h-72 bg-gray-100 dark:bg-gray-800 rounded-xl animate-pulse" />;

    const actas = data?.actas ?? [];
    const filtradas = filtro
        ? actas.filter(
              (a) =>
                  a.codigoMesa?.toLowerCase().includes(filtro.toLowerCase()) ||
                  a.recinto?.toLowerCase().includes(filtro.toLowerCase()) ||
                  a.departamento?.toLowerCase().includes(filtro.toLowerCase()),
          )
        : actas;

    return (
        <div className="bg-white dark:bg-gray-800 p-6 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm">
            <div className="flex items-center justify-between mb-3 gap-2">
                <h3 className="text-sm font-bold text-gray-900 dark:text-white">
                    Transparencia — Acceso a actas digitalizadas (SRRV)
                </h3>
                <div className="relative">
                    <Search size={14} className="absolute left-2 top-2 text-gray-400" />
                    <input
                        value={filtro}
                        onChange={(e) => setFiltro(e.target.value)}
                        placeholder="Buscar mesa/recinto/depto"
                        className="bg-white dark:bg-gray-900 border border-gray-300 dark:border-gray-700 text-sm rounded-md pl-7 pr-3 py-1.5 min-w-56"
                    />
                </div>
            </div>
            <p className="text-[11px] text-gray-400 mb-2">
                Total actas con imagen: {actas.length.toLocaleString("es-BO")} · Mostrando: {filtradas.length}
            </p>
            <div className="overflow-x-auto max-h-96 overflow-y-auto">
                <table className="w-full text-sm">
                    <thead className="sticky top-0 bg-gray-50 dark:bg-gray-900">
                        <tr className="text-[11px] uppercase text-gray-500">
                            <th className="px-3 py-2 text-left">Mesa</th>
                            <th className="px-3 py-2 text-left">Departamento</th>
                            <th className="px-3 py-2 text-left">Recinto</th>
                            <th className="px-3 py-2 text-left">Válidos</th>
                            <th className="px-3 py-2 text-left">Nulos</th>
                            <th className="px-3 py-2 text-left">Blanco</th>
                            <th className="px-3 py-2 text-left">Imagen</th>
                        </tr>
                    </thead>
                    <tbody>
                        {filtradas.slice(0, 200).map((a, i) => (
                            <tr key={`${a.codigoMesa}-${i}`} className="border-t border-gray-100 dark:border-gray-700">
                                <td className="px-3 py-2 font-mono text-xs">{a.codigoMesa}</td>
                                <td className="px-3 py-2">{a.departamento}</td>
                                <td className="px-3 py-2 text-xs">{a.recinto}</td>
                                <td className="px-3 py-2">{a.votosValidos}</td>
                                <td className="px-3 py-2">{a.votosNulos}</td>
                                <td className="px-3 py-2">{a.votosBlanco}</td>
                                <td className="px-3 py-2">
                                    {a.urlImagen ? (
                                        <a
                                            href={a.urlImagen}
                                            target="_blank"
                                            rel="noreferrer"
                                            className="inline-flex items-center gap-1 text-blue-500 text-xs hover:underline"
                                        >
                                            Ver <ExternalLink size={10} />
                                        </a>
                                    ) : (
                                        <span className="text-gray-400 text-xs">—</span>
                                    )}
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
                {filtradas.length > 200 && (
                    <p className="text-[10px] text-gray-400 mt-2 text-center">
                        Mostrando primeras 200 de {filtradas.length}. Refina filtro para ver más.
                    </p>
                )}
            </div>
        </div>
    );
};
