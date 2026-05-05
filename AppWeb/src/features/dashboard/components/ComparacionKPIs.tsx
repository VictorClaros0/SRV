import { useComparacion } from "../hooks/useComparacion";

const fmt = (n?: number) => (n == null ? "—" : Number(n).toLocaleString("es-BO"));
const fmtSigned = (n: number) => `${n >= 0 ? "+" : ""}${fmt(n)}`;
const colorDelta = (n: number) =>
    n === 0 ? "text-gray-400" : n > 0 ? "text-blue-500" : "text-red-500";

export const ComparacionKPIs = () => {
    const c = useComparacion();

    if (c.isLoading)
        return (
            <div className="grid grid-cols-2 md:grid-cols-4 gap-3 animate-pulse">
                {[...Array(4)].map((_, i) => (
                    <div key={i} className="h-32 bg-gray-200 dark:bg-gray-800 rounded-xl" />
                ))}
            </div>
        );

    if (c.isError || !c.rrv || !c.oficial)
        return (
            <div className="bg-red-50 dark:bg-red-900/20 text-red-600 p-4 rounded-xl text-sm">
                Error: no se pudo conectar a SRRV (puerto 8081) u Oficial.
            </div>
        );

    const cards = [
        {
            label: "Actas procesadas",
            rrv: c.rrv.totalActasProcesadas,
            oficial: c.oficial.actasTranscritas,
            delta: c.deltaActas,
        },
        {
            label: "Votos válidos",
            rrv: c.rrv.votosValidos,
            oficial: c.oficial.votosValidos,
            delta: c.deltaValidos,
        },
        {
            label: "Votos nulos",
            rrv: c.rrv.votosNulos,
            oficial: c.oficial.votosNulos,
            delta: c.deltaNulos,
        },
        {
            label: "Votos blanco",
            rrv: c.rrv.votosBlancos,
            oficial: c.oficial.votosBlanco,
            delta: c.deltaBlanco,
        },
    ];

    return (
        <div>
            <div className="flex items-center justify-between mb-3">
                <h3 className="text-sm font-bold text-gray-900 dark:text-white">
                    Comparación RRV vs Cómputo Oficial
                </h3>
                <span className="text-xs text-gray-500">
                    Confiabilidad RRV: <strong>{c.confiabilidadRRV.toFixed(2)}%</strong>
                </span>
            </div>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
                {cards.map((card) => (
                    <div
                        key={card.label}
                        className="bg-white dark:bg-gray-800 rounded-xl px-4 py-3 shadow-sm border border-gray-200 dark:border-gray-700"
                    >
                        <div className="text-[11px] font-bold text-gray-500 uppercase tracking-wide mb-2">
                            {card.label}
                        </div>
                        <div className="grid grid-cols-2 gap-2 mb-2">
                            <div>
                                <div className="text-[10px] text-orange-500 font-bold">RRV</div>
                                <div className="text-lg font-extrabold text-orange-500">{fmt(card.rrv)}</div>
                            </div>
                            <div>
                                <div className="text-[10px] text-blue-500 font-bold">OFICIAL</div>
                                <div className="text-lg font-extrabold text-blue-500">{fmt(card.oficial)}</div>
                            </div>
                        </div>
                        <div className={`text-xs font-semibold ${colorDelta(card.delta)}`}>
                            Δ {fmtSigned(card.delta)}
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
};
