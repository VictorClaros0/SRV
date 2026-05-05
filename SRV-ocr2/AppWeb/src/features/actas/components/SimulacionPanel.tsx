import { useEffect, useRef, useState } from "react";

export interface SimulacionSampleData {
    codigoActa: number;
    codigoRecinto: number;
    nroMesa: number;
    papeletasAnfora: number;
    papeletasNoUsadas: number;
    p1: number;
    p2: number;
    p3: number;
    p4: number;
    votosValidos: number;
    votosNulos: number;
    votosBlanco: number;
    aperturaHora: number;
    aperturaMinutos: number;
    cierreHora: number;
    cierreMinutos: number;
    observaciones: string;
}

const SIM_FIELDS: { key: keyof SimulacionSampleData; label: string }[] = [
    { key: "codigoActa", label: "Código Acta" },
    { key: "codigoRecinto", label: "Cód. Recinto" },
    { key: "nroMesa", label: "Nro Mesa" },
    { key: "papeletasAnfora", label: "Pap. Ánfora" },
    { key: "papeletasNoUsadas", label: "Pap. No Usadas" },
    { key: "p1", label: "P1" },
    { key: "p2", label: "P2" },
    { key: "p3", label: "P3" },
    { key: "p4", label: "P4" },
    { key: "votosValidos", label: "Válidos" },
    { key: "votosNulos", label: "Nulos" },
    { key: "votosBlanco", label: "Blancos" },
    { key: "aperturaHora", label: "Ap. Hora" },
    { key: "aperturaMinutos", label: "Ap. Min" },
    { key: "cierreHora", label: "Ci. Hora" },
    { key: "cierreMinutos", label: "Ci. Min" },
    { key: "observaciones", label: "Observaciones" },
];

const ANIM_SAMPLE: SimulacionSampleData[] = [
    { codigoActa: 1010200001001, codigoRecinto: 1010200001, nroMesa: 1, papeletasAnfora: 788, papeletasNoUsadas: 89, p1: 140, p2: 39, p3: 124, p4: 345, votosValidos: 648, votosNulos: 64, votosBlanco: 76, aperturaHora: 8, aperturaMinutos: 0, cierreHora: 16, cierreMinutos: 30, observaciones: "" },
    { codigoActa: 1020300002015, codigoRecinto: 1020300002, nroMesa: 15, papeletasAnfora: 652, papeletasNoUsadas: 124, p1: 287, p2: 91, p3: 55, p4: 189, votosValidos: 622, votosNulos: 18, votosBlanco: 12, aperturaHora: 8, aperturaMinutos: 0, cierreHora: 17, cierreMinutos: 0, observaciones: "" },
    { codigoActa: 2040100003032, codigoRecinto: 2040100003, nroMesa: 32, papeletasAnfora: 512, papeletasNoUsadas: 201, p1: 68, p2: 143, p3: 201, p4: 76, votosValidos: 488, votosNulos: 15, votosBlanco: 9, aperturaHora: 8, aperturaMinutos: 30, cierreHora: 16, cierreMinutos: 45, observaciones: "Tachadura en campo P3" },
    { codigoActa: 3050200004007, codigoRecinto: 3050200004, nroMesa: 7, papeletasAnfora: 891, papeletasNoUsadas: 45, p1: 412, p2: 187, p3: 93, p4: 154, votosValidos: 846, votosNulos: 32, votosBlanco: 13, aperturaHora: 8, aperturaMinutos: 0, cierreHora: 17, cierreMinutos: 30, observaciones: "" },
    { codigoActa: 4060300005020, codigoRecinto: 4060300005, nroMesa: 20, papeletasAnfora: 445, papeletasNoUsadas: 178, p1: 56, p2: 312, p3: 48, p4: 21, votosValidos: 437, votosNulos: 5, votosBlanco: 3, aperturaHora: 8, aperturaMinutos: 0, cierreHora: 16, cierreMinutos: 0, observaciones: "" },
];

interface Props {
    onComplete: (sample: SimulacionSampleData) => void;
}

interface UiState {
    displayed: Partial<Record<keyof SimulacionSampleData, string>>;
    activeField: number;
    actaIdx: number;
}

export const SimulacionPanel = ({ onComplete }: Props) => {
    const timerRef = useRef<number | null>(null);
    const stoppedRef = useRef(false);
    const stateRef = useRef({ actaIdx: 0, fieldIdx: 0, charIdx: 0 });
    const onCompleteRef = useRef(onComplete);
    onCompleteRef.current = onComplete;

    const [cursor, setCursor] = useState(true);
    const [dotOn, setDotOn] = useState(true);
    const [ui, setUi] = useState<UiState>({ displayed: {}, activeField: 0, actaIdx: 0 });

    useEffect(() => {
        stoppedRef.current = false;

        const tick = () => {
            if (stoppedRef.current) return;
            const st = stateRef.current;
            const acta = ANIM_SAMPLE[st.actaIdx % ANIM_SAMPLE.length];
            const field = SIM_FIELDS[st.fieldIdx];
            const fullVal = String(acta[field.key] ?? "");

            if (st.charIdx < fullVal.length) {
                st.charIdx++;
                const partial = fullVal.slice(0, st.charIdx);
                setUi((prev) => ({ ...prev, displayed: { ...prev.displayed, [field.key]: partial } }));
                timerRef.current = window.setTimeout(tick, 20);
            } else if (st.fieldIdx < SIM_FIELDS.length - 1) {
                st.fieldIdx++;
                st.charIdx = 0;
                setUi((prev) => ({ ...prev, activeField: st.fieldIdx }));
                timerRef.current = window.setTimeout(tick, 160);
            } else {
                onCompleteRef.current(ANIM_SAMPLE[st.actaIdx % ANIM_SAMPLE.length]);
                st.actaIdx++;
                st.fieldIdx = 0;
                st.charIdx = 0;
                setUi({ displayed: {}, activeField: 0, actaIdx: st.actaIdx });
                timerRef.current = window.setTimeout(tick, 350);
            }
        };

        timerRef.current = window.setTimeout(tick, 20);
        const blinkId = window.setInterval(() => setCursor((v) => !v), 530);
        const dotId = window.setInterval(() => setDotOn((v) => !v), 700);

        return () => {
            stoppedRef.current = true;
            if (timerRef.current) clearTimeout(timerRef.current);
            clearInterval(blinkId);
            clearInterval(dotId);
        };
    }, []);

    const acta = ANIM_SAMPLE[ui.actaIdx % ANIM_SAMPLE.length];

    return (
        <div className="bg-white dark:bg-gray-800 border border-green-500/40 rounded-xl px-5 py-4 mb-4 shadow-[0_2px_8px_rgba(39,174,96,0.08)]">
            <div className="flex items-center gap-2 mb-3">
                <span
                    className="w-2 h-2 rounded-full bg-green-500 inline-block transition-opacity duration-300"
                    style={{ opacity: dotOn ? 1 : 0.2 }}
                />
                <span className="text-sm font-bold text-gray-900 dark:text-white">
                    Transcribiendo acta #{acta.codigoActa}
                </span>
                <span className="text-xs text-gray-500 dark:text-gray-400 ml-1">
                    — operador ingresando datos
                </span>
            </div>
            <div className="grid gap-x-2.5 gap-y-2 [grid-template-columns:repeat(auto-fill,minmax(128px,1fr))]">
                {SIM_FIELDS.map((f, i) => {
                    const isActive = i === ui.activeField;
                    const isDone = i < ui.activeField;
                    const val = ui.displayed[f.key] ?? "";
                    const fieldClass = isActive
                        ? "border-blue-500 bg-blue-50 dark:bg-blue-900/20"
                        : isDone
                            ? "border-green-200 bg-green-50/40 dark:bg-green-900/10"
                            : "border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-900/30";
                    const valueClass = isActive
                        ? "text-blue-700 dark:text-blue-300"
                        : isDone
                            ? "text-green-600 dark:text-green-400"
                            : "text-gray-300 dark:text-gray-600";
                    return (
                        <div
                            key={f.key}
                            className={`rounded-md px-2.5 py-1.5 border transition-colors duration-150 ${fieldClass}`}
                        >
                            <div className="text-[10px] font-bold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-0.5">
                                {f.label}
                            </div>
                            <div className={`text-sm font-semibold font-mono min-h-[18px] break-all ${valueClass}`}>
                                {val || (!isActive && <span className="text-gray-300 dark:text-gray-600">—</span>)}
                                {isActive && (
                                    <span
                                        className="text-blue-500 font-light"
                                        style={{ opacity: cursor ? 1 : 0 }}
                                    >
                                        |
                                    </span>
                                )}
                            </div>
                        </div>
                    );
                })}
            </div>
        </div>
    );
};

export default SimulacionPanel;
