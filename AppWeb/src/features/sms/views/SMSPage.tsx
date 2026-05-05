import { useEffect, useState, useCallback } from "react";
import {
    MessageSquare, RefreshCw, CheckCircle, AlertTriangle, XCircle,
    Clock, ShieldCheck, ShieldOff, Database, Ban, Eye, EyeOff,
} from "lucide-react";
import clsx from "clsx";

// Usa el proxy de Vite: /sms-api → http://localhost:3001
const SMS_API = "/sms-api";

interface ParsedFields {
    MESA: string | number;
    P1: number;
    P2: number;
    P3: number;
    P4: number;
    BLANCOS: number;
    NULOS: number;
    TOTAL: number;
    OBS?: string;
}

interface ParseResult {
    valid: boolean;
    reason: string | null;
    fields: ParsedFields | null;
    calculated?: number;
    declared?: number;
    totalsMatch?: boolean;
}

interface SMSMessage {
    id: number;
    receivedAt: string;
    sender: string;
    senderNormalized: string;
    senderAuthorized: boolean;
    smsBody: string;
    parsed: ParseResult;
    status: "VALIDADA" | "PENDIENTE_REVISION" | "RECHAZADA" | "NO_AUTORIZADO";
    isDuplicate: boolean;
    srvStatus?: "GUARDADO" | "OMITIDO" | "ERROR" | "PENDIENTE";
    raw: Record<string, unknown>;
}

const estadoConfig: Record<string, { label: string; icon: React.ReactNode; cls: string }> = {
    VALIDADA: {
        label: "Validada",
        icon: <CheckCircle size={13} />,
        cls: "bg-green-100 text-green-800 dark:bg-green-900/40 dark:text-green-300",
    },
    PENDIENTE_REVISION: {
        label: "Pendiente revisión",
        icon: <AlertTriangle size={13} />,
        cls: "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/40 dark:text-yellow-300",
    },
    RECHAZADA: {
        label: "Rechazada",
        icon: <XCircle size={13} />,
        cls: "bg-red-100 text-red-800 dark:bg-red-900/40 dark:text-red-300",
    },
    NO_AUTORIZADO: {
        label: "No autorizado",
        icon: <ShieldOff size={13} />,
        cls: "bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300",
    },
};

const statColorCls: Record<string, string> = {
    blue:   "text-blue-600 dark:text-blue-400",
    green:  "text-green-600 dark:text-green-400",
    yellow: "text-yellow-600 dark:text-yellow-400",
    red:    "text-red-600 dark:text-red-400",
    indigo: "text-indigo-600 dark:text-indigo-400",
    gray:   "text-gray-500 dark:text-gray-400",
};

const srvConfig: Record<string, { label: string; cls: string }> = {
    GUARDADO: { label: "Guardado en SRV ✓", cls: "bg-blue-100 text-blue-800 dark:bg-blue-900/40 dark:text-blue-300" },
    OMITIDO:  { label: "No guardado",       cls: "bg-gray-100 text-gray-500 dark:bg-gray-800 dark:text-gray-400" },
    ERROR:    { label: "Error SRV",         cls: "bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-400" },
    PENDIENTE:{ label: "Guardando…",        cls: "bg-yellow-100 text-yellow-700 dark:bg-yellow-900/40 dark:text-yellow-300" },
};

function EstadoBadge({ estado, isDuplicate }: Readonly<{ estado: string; isDuplicate?: boolean }>) {
    const cfg = estadoConfig[estado] ?? {
        label: estado,
        icon: <Clock size={13} />,
        cls: "bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300",
    };
    return (
        <span className="inline-flex items-center gap-1.5 flex-wrap">
            <span className={clsx("inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium", cfg.cls)}>
                {cfg.icon}{cfg.label}
            </span>
            {isDuplicate && (
                <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-purple-100 text-purple-800 dark:bg-purple-900/40 dark:text-purple-300">
                    DUPLICADO
                </span>
            )}
        </span>
    );
}

function SrvBadge({ status }: Readonly<{ status?: string }>) {
    if (!status) return null;
    const cfg = srvConfig[status];
    if (!cfg) return null;
    return (
        <span className={clsx("inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium", cfg.cls)}>
            <Database size={11} />{cfg.label}
        </span>
    );
}

interface MessageRowProps {
    m: SMSMessage;
    isSelected: boolean;
    onSelect: (m: SMSMessage) => void;
}

function MessageRow({ m, isSelected, onSelect }: Readonly<MessageRowProps>) {
    return (
        <tr
            onClick={() => onSelect(m)}
            className={clsx(
                "cursor-pointer transition-colors",
                !m.senderAuthorized && "opacity-50",
                isSelected ? "bg-blue-50 dark:bg-blue-900/20" : "hover:bg-gray-50 dark:hover:bg-gray-800/40"
            )}
        >
            <td className="px-3 py-2.5 text-gray-400 text-xs">{m.id}</td>
            <td className="px-3 py-2.5 text-xs text-gray-500 dark:text-gray-400 whitespace-nowrap">
                {new Date(m.receivedAt).toLocaleString("es-BO")}
            </td>
            <td className={clsx(
                "px-3 py-2.5 font-mono text-xs font-semibold",
                m.senderAuthorized ? "text-green-600 dark:text-green-400" : "text-red-500 dark:text-red-400"
            )}>
                <span className="inline-flex items-center gap-1">
                    {m.senderAuthorized ? <ShieldCheck size={12} /> : <Ban size={12} />}
                    {m.senderNormalized || m.sender}
                </span>
            </td>
            <td className="px-3 py-2.5 font-medium text-gray-900 dark:text-white text-xs">
                {m.parsed?.fields?.MESA ?? "—"}
            </td>
            <td className="px-3 py-2.5 text-xs">
                {m.parsed?.fields?.OBS ? (
                    <span className="inline-flex items-center px-1.5 py-0.5 rounded bg-purple-100 text-purple-700 dark:bg-purple-900/40 dark:text-purple-300 font-medium max-w-[120px] truncate" title={m.parsed.fields.OBS}>
                        {m.parsed.fields.OBS}
                    </span>
                ) : (
                    <span className="text-gray-300 dark:text-gray-600">—</span>
                )}
            </td>
            <td className="px-3 py-2.5">
                <EstadoBadge estado={m.status} isDuplicate={m.isDuplicate} />
            </td>
            <td className="px-3 py-2.5">
                <SrvBadge status={m.srvStatus} />
            </td>
        </tr>
    );
}

interface ListBodyProps {
    loading: boolean;
    allMessages: SMSMessage[];
    visibleMessages: SMSMessage[];
    selectedId: number | undefined;
    onSelect: (m: SMSMessage) => void;
}

function ListBody({ loading, allMessages, visibleMessages, selectedId, onSelect }: Readonly<ListBodyProps>) {
    if (loading && allMessages.length === 0) {
        return (
            <div className="p-8 text-center text-gray-400">
                <RefreshCw size={24} className="animate-spin mx-auto mb-2" />
                <p className="text-sm">Cargando...</p>
            </div>
        );
    }
    if (visibleMessages.length === 0) {
        return (
            <div className="p-8 text-center text-gray-400">
                <MessageSquare size={32} className="mx-auto mb-2 opacity-40" />
                <p className="text-sm">No se han recibido mensajes de números autorizados.</p>
                <p className="text-xs mt-1 text-gray-400">
                    Webhook: <code className="bg-gray-100 dark:bg-gray-800 px-1 rounded">POST localhost:3001/webhook</code>
                </p>
            </div>
        );
    }
    return (
        <div className="overflow-x-auto">
            <table className="w-full text-sm">
                <thead>
                    <tr className="border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/50">
                        <th className="px-3 py-3 text-left font-medium text-gray-600 dark:text-gray-400 text-xs">#</th>
                        <th className="px-3 py-3 text-left font-medium text-gray-600 dark:text-gray-400 text-xs">Fecha</th>
                        <th className="px-3 py-3 text-left font-medium text-gray-600 dark:text-gray-400 text-xs">Número</th>
                        <th className="px-3 py-3 text-left font-medium text-gray-600 dark:text-gray-400 text-xs">Mesa</th>
                        <th className="px-3 py-3 text-left font-medium text-gray-600 dark:text-gray-400 text-xs">OBS</th>
                        <th className="px-3 py-3 text-left font-medium text-gray-600 dark:text-gray-400 text-xs">Estado</th>
                        <th className="px-3 py-3 text-left font-medium text-gray-600 dark:text-gray-400 text-xs">SRV</th>
                    </tr>
                </thead>
                <tbody className="divide-y divide-gray-100 dark:divide-gray-800">
                    {visibleMessages.map((m) => (
                        <MessageRow key={m.id} m={m} isSelected={m.id === selectedId} onSelect={onSelect} />
                    ))}
                </tbody>
            </table>
        </div>
    );
}

export default function SMSPage() {
    const [allMessages, setAllMessages] = useState<SMSMessage[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState("");
    const [selected, setSelected] = useState<SMSMessage | null>(null);
    const [showUnauthorized, setShowUnauthorized] = useState(false);

    const fetchMessages = useCallback(async () => {
        setLoading(true);
        setError("");
        try {
            const res = await fetch(`${SMS_API}/messages`);
            if (!res.ok) throw new Error(`HTTP ${res.status}`);
            const data: SMSMessage[] = await res.json();
            setAllMessages(data);
        } catch {
            setError("No se pudo conectar al servidor SMS (localhost:3001). ¿Está corriendo?");
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        fetchMessages();
        const interval = setInterval(fetchMessages, 3000);
        return () => clearInterval(interval);
    }, [fetchMessages]);

    // ── Seguridad: separar autorizados de no autorizados ──────────────────────
    const authorizedMessages = allMessages.filter((m) => m.senderAuthorized);
    const unauthorizedMessages = allMessages.filter((m) => !m.senderAuthorized);
    const visibleMessages = showUnauthorized ? allMessages : authorizedMessages;

    const stats = {
        total:      authorizedMessages.length,
        validadas:  authorizedMessages.filter((m) => m.status === "VALIDADA").length,
        pendientes: authorizedMessages.filter((m) => m.status === "PENDIENTE_REVISION").length,
        rechazadas: authorizedMessages.filter((m) => m.status === "RECHAZADA").length,
        guardados:  authorizedMessages.filter((m) => m.srvStatus === "GUARDADO").length,
        noAuth:     unauthorizedMessages.length,
    };

    return (
        <div className="p-4 md:p-6 space-y-5 max-w-6xl mx-auto">

            {/* ── Header ─────────────────────────────────────────────────────── */}
            <div className="flex items-center justify-between flex-wrap gap-3">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900 dark:text-white flex items-center gap-2">
                        <MessageSquare size={24} className="text-blue-600" />
                        Mensajes SMS
                    </h1>
                    <p className="text-sm text-gray-500 dark:text-gray-400 mt-0.5">
                        Actas recibidas por SMS — conectado a sms-test (localhost:3001)
                    </p>
                </div>
                <div className="flex items-center gap-2 flex-wrap">
                    <button
                        onClick={() => setShowUnauthorized((v) => !v)}
                        className={clsx(
                            "flex items-center gap-1.5 px-3 py-2 rounded-lg text-xs font-medium border transition-colors",
                            showUnauthorized
                                ? "bg-red-50 border-red-300 text-red-700 dark:bg-red-900/20 dark:border-red-700 dark:text-red-300"
                                : "bg-white border-gray-200 text-gray-600 dark:bg-gray-900 dark:border-gray-700 dark:text-gray-400"
                        )}
                    >
                        {showUnauthorized ? <EyeOff size={13} /> : <Eye size={13} />}
                        {showUnauthorized ? "Ocultar no autorizados" : "Ver no autorizados"}
                        {unauthorizedMessages.length > 0 && (
                            <span className="ml-1 bg-red-500 text-white rounded-full w-4 h-4 text-[10px] flex items-center justify-center">
                                {unauthorizedMessages.length}
                            </span>
                        )}
                    </button>
                    <button
                        onClick={fetchMessages}
                        disabled={loading}
                        className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm font-medium transition-colors disabled:opacity-50"
                    >
                        <RefreshCw size={15} className={loading ? "animate-spin" : ""} />
                        Actualizar
                    </button>
                </div>
            </div>

            {/* ── Error ──────────────────────────────────────────────────────── */}
            {error && (
                <div className="p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg text-sm text-red-700 dark:text-red-300">
                    {error}
                </div>
            )}

            {/* ── Alerta de no autorizados ────────────────────────────────────── */}
            {!showUnauthorized && unauthorizedMessages.length > 0 && (
                <div className="flex items-center gap-2 p-3 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg text-sm text-yellow-800 dark:text-yellow-300">
                    <ShieldOff size={15} />
                    <span>
                        <strong>{unauthorizedMessages.length}</strong> mensaje(s) de números <strong>no autorizados</strong> bloqueado(s) — no se guardaron en el SRV.
                    </span>
                </div>
            )}

            {/* ── Stats ──────────────────────────────────────────────────────── */}
            <div className="grid grid-cols-2 md:grid-cols-6 gap-3">
                {[
                    { label: "Autorizados",     value: stats.total,      color: "blue"   },
                    { label: "Validadas",        value: stats.validadas,  color: "green"  },
                    { label: "Pendiente rev.",   value: stats.pendientes, color: "yellow" },
                    { label: "Rechazadas",       value: stats.rechazadas, color: "red"    },
                    { label: "Guardadas SRV",    value: stats.guardados,  color: "indigo" },
                    { label: "No autorizados",   value: stats.noAuth,     color: "gray"   },
                ].map((s) => (
                    <div key={s.label} className="bg-white dark:bg-gray-900 rounded-xl p-3 border border-gray-200 dark:border-gray-800">
                        <p className="text-xs text-gray-500 dark:text-gray-400 leading-tight">{s.label}</p>
                        <p className={clsx("text-2xl font-bold mt-1", statColorCls[s.color] ?? "text-gray-500 dark:text-gray-400")}>
                            {s.value}
                        </p>
                    </div>
                ))}
            </div>

            {/* ── Lista + detalle ─────────────────────────────────────────────── */}
            <div className={clsx("grid gap-4", selected ? "md:grid-cols-2" : "grid-cols-1")}>

                {/* Lista */}
                <div className="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
                    <ListBody
                        loading={loading}
                        allMessages={allMessages}
                        visibleMessages={visibleMessages}
                        selectedId={selected?.id}
                        onSelect={(m) => setSelected(selected?.id === m.id ? null : m)}
                    />
                </div>

                {/* Panel de detalle */}
                {selected && (
                    <div className="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4 space-y-4">
                        <div className="flex items-center justify-between">
                            <h3 className="font-semibold text-gray-800 dark:text-white text-sm">Detalle #{selected.id}</h3>
                            <button onClick={() => setSelected(null)} className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 text-xs">
                                ✕ cerrar
                            </button>
                        </div>

                        <div className="space-y-1.5 text-sm">
                            <div className="flex justify-between items-center">
                                <span className="text-gray-500 text-xs">Remitente</span>
                                <span className={clsx(
                                    "font-mono font-semibold text-xs flex items-center gap-1",
                                    selected.senderAuthorized ? "text-green-600 dark:text-green-400" : "text-red-500"
                                )}>
                                    {selected.senderAuthorized ? <ShieldCheck size={12} /> : <Ban size={12} />}
                                    {selected.sender}
                                    {selected.senderAuthorized ? " ✓ Autorizado" : " ✗ No autorizado"}
                                </span>
                            </div>
                            <div className="flex justify-between items-center">
                                <span className="text-gray-500 text-xs">Estado</span>
                                <EstadoBadge estado={selected.status} isDuplicate={selected.isDuplicate} />
                            </div>
                            <div className="flex justify-between items-center">
                                <span className="text-gray-500 text-xs">Guardado SRV</span>
                                {selected.srvStatus
                                    ? <SrvBadge status={selected.srvStatus} />
                                    : <span className="text-xs text-gray-400">—</span>}
                            </div>
                            <div className="flex justify-between items-center">
                                <span className="text-gray-500 text-xs">Recibido</span>
                                <span className="text-gray-700 dark:text-gray-300 text-xs">
                                    {new Date(selected.receivedAt).toLocaleString("es-BO")}
                                </span>
                            </div>
                        </div>

                        <div>
                            <p className="text-xs text-gray-500 mb-1">SMS raw</p>
                            <code className="block bg-gray-50 dark:bg-gray-800 rounded p-2 text-xs text-green-700 dark:text-green-300 break-all">
                                {selected.smsBody}
                            </code>
                        </div>

                        {selected.parsed?.valid && selected.parsed.fields && (
                            <div>
                                <p className="text-xs text-gray-500 mb-2">Votos</p>
                                <div className="grid grid-cols-4 gap-1.5">
                                    {(["MESA", "P1", "P2", "P3", "P4", "BLANCOS", "NULOS"] as const).map((k) => (
                                        <div key={k} className="bg-gray-50 dark:bg-gray-800 rounded-lg p-2 text-center">
                                            <p className="text-[10px] text-gray-400 uppercase">{k}</p>
                                            <p className="font-bold text-gray-900 dark:text-white text-sm">{selected.parsed.fields?.[k]}</p>
                                        </div>
                                    ))}
                                </div>
                                <div className="mt-1.5 grid grid-cols-2 gap-1.5">
                                    <div className={clsx(
                                        "rounded-lg p-2 text-center border",
                                        selected.parsed.totalsMatch
                                            ? "bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800"
                                            : "bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800"
                                    )}>
                                        <p className="text-[10px] text-gray-400">TOTAL declarado</p>
                                        <p className="font-bold text-sm">{selected.parsed.declared}</p>
                                    </div>
                                    <div className={clsx(
                                        "rounded-lg p-2 text-center border",
                                        selected.parsed.totalsMatch
                                            ? "bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800"
                                            : "bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800"
                                    )}>
                                        <p className="text-[10px] text-gray-400">TOTAL calculado</p>
                                        <p className="font-bold text-sm">
                                            {selected.parsed.calculated} {selected.parsed.totalsMatch ? "✓" : "✗"}
                                        </p>
                                    </div>
                                </div>
                                {selected.parsed.fields.OBS && (
                                    <div className="mt-1.5 bg-purple-50 dark:bg-purple-900/20 border border-purple-200 dark:border-purple-800 rounded-lg p-2">
                                        <p className="text-[10px] text-gray-400 mb-0.5">Observaciones</p>
                                        <p className="text-sm text-purple-800 dark:text-purple-300">{selected.parsed.fields.OBS}</p>
                                    </div>
                                )}
                            </div>
                        )}

                        {selected.parsed?.reason && (
                            <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-2 text-xs text-red-700 dark:text-red-300">
                                ⚠ {selected.parsed.reason}
                            </div>
                        )}
                    </div>
                )}
            </div>

            {/* ── Reglas de seguridad ─────────────────────────────────────────── */}
            <div className="grid md:grid-cols-2 gap-3">
                <div className="bg-green-50 dark:bg-green-900/10 border border-green-200 dark:border-green-800 rounded-xl p-4 text-sm">
                    <p className="font-semibold text-green-800 dark:text-green-300 mb-2 flex items-center gap-1.5">
                        <ShieldCheck size={15} /> Mecanismos de seguridad activos
                    </p>
                    <ul className="space-y-1 text-xs text-green-700 dark:text-green-400">
                        <li>✓ Solo se procesan mensajes de números registrados</li>
                        <li>✓ Formato obligatorio: <code className="bg-green-100 dark:bg-green-900/40 px-1 rounded">RRV|MESA|P1..P4|BLANCOS|NULOS|TOTAL</code></li>
                        <li>✓ Validación de totales: P1+P2+P3+P4+BLANCOS+NULOS = TOTAL</li>
                        <li>✓ Se acepta campo opcional <code className="bg-green-100 dark:bg-green-900/40 px-1 rounded">OBS=…</code></li>
                        <li>✓ Solo se guarda en SRV si el mensaje es válido y los totales cuadran</li>
                        <li>✓ Detección de actas duplicadas por número de mesa</li>
                    </ul>
                </div>
                <div className="bg-blue-50 dark:bg-blue-900/10 border border-blue-200 dark:border-blue-800 rounded-xl p-4 text-sm">
                    <p className="font-semibold text-blue-800 dark:text-blue-300 mb-2">Formato SMS esperado</p>
                    <code className="block bg-blue-100 dark:bg-blue-900/40 rounded p-2 font-mono text-xs text-blue-900 dark:text-blue-200 break-all">
                        RRV|MESA=001|P1=45|P2=32|P3=18|P4=5|BLANCOS=3|NULOS=2|TOTAL=105
                    </code>
                    <code className="block bg-blue-100 dark:bg-blue-900/40 rounded p-2 font-mono text-xs text-blue-900 dark:text-blue-200 break-all mt-1">
                        RRV|MESA=001|…|TOTAL=105|OBS=acta con manchas
                    </code>
                    <p className="mt-2 text-xs text-blue-700 dark:text-blue-400">
                        Webhook ngrok → <code className="bg-blue-100 dark:bg-blue-900/40 px-1 rounded">POST localhost:3001/webhook</code>
                    </p>
                </div>
            </div>
        </div>
    );
}
