import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import axios from "axios";
import { ArrowLeft, AlertCircle, CheckCircle2 } from "lucide-react";
import { actasService } from "../services/actasService";
import {
    type ActaCreateRequest,
    emptyActaCreate,
} from "../models/request/acta-create-request";
import type { ActaEstado } from "../models/response/acta-response";

const SECTION_TITLE =
    "text-xs font-bold uppercase tracking-wide text-gray-500 dark:text-gray-400 mb-3 pb-2 border-b border-gray-100 dark:border-gray-700";
const LABEL = "text-xs font-semibold text-gray-600 dark:text-gray-300";
const INPUT =
    "px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-900 text-sm text-gray-900 dark:text-white outline-none focus:ring-2 focus:ring-blue-500";

interface FormState extends ActaCreateRequest {
    codigoActaText: string;
    codigoRecintoText: string;
}

const buildInitial = (): FormState => ({
    ...emptyActaCreate(),
    codigoActaText: "",
    codigoRecintoText: "",
});

const fromActa = (a: Partial<ActaCreateRequest> & { codigoActa?: number | string; codigoRecinto?: number | string }): FormState => {
    const base = emptyActaCreate();
    return {
        ...base,
        ...a,
        codigoActa: Number(a.codigoActa) || 0,
        codigoRecinto: Number(a.codigoRecinto) || 0,
        codigoActaText: a.codigoActa != null ? String(a.codigoActa) : "",
        codigoRecintoText: a.codigoRecinto != null ? String(a.codigoRecinto) : "",
    };
};

export const ActaFormPage = () => {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const editing = Boolean(id);

    const [form, setForm] = useState<FormState>(buildInitial);
    const [error, setError] = useState<string | null>(null);
    const [success, setSuccess] = useState<string | null>(null);
    const [loading, setLoading] = useState(false);
    const [fetchLoading, setFetchLoading] = useState(editing);

    useEffect(() => {
        if (!editing || !id) return;
        const { call, controller } = actasService.get(id);
        call
            .then(({ data }) => setForm(fromActa(data as unknown as ActaCreateRequest)))
            .catch((err) => {
                if (axios.isCancel(err)) return;
                const apiMsg =
                    axios.isAxiosError(err) &&
                    (err.response?.data as { error?: string })?.error;
                setError(apiMsg || "No se pudo cargar el acta.");
            })
            .finally(() => setFetchLoading(false));
        return () => controller.abort();
    }, [id, editing]);

    const setText = <K extends "codigoActaText" | "codigoRecintoText" | "observaciones">(field: K) =>
        (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
            const value = e.target.value;
            setForm((prev) => ({ ...prev, [field]: value } as FormState));
        };

    const setEstado = (e: React.ChangeEvent<HTMLSelectElement>) => {
        setForm((prev) => ({ ...prev, estado: e.target.value as ActaEstado }));
    };

    const setNum = (field: keyof ActaCreateRequest) =>
        (e: React.ChangeEvent<HTMLInputElement>) => {
            const value = Number(e.target.value) || 0;
            setForm((prev) => ({ ...prev, [field]: value } as FormState));
        };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError(null);
        setSuccess(null);
        setLoading(true);

        const payload: ActaCreateRequest = {
            codigoActa: Number(form.codigoActaText) || 0,
            codigoRecinto: Number(form.codigoRecintoText) || 0,
            nroMesa: Number(form.nroMesa) || 0,
            estado: form.estado,
            papeletasAnfora: form.papeletasAnfora,
            papeletasNoUsadas: form.papeletasNoUsadas,
            p1: form.p1,
            p2: form.p2,
            p3: form.p3,
            p4: form.p4,
            votosValidos: form.votosValidos,
            votosNulos: form.votosNulos,
            votosBlanco: form.votosBlanco,
            observaciones: form.observaciones,
            aperturaHora: form.aperturaHora,
            aperturaMinutos: form.aperturaMinutos,
            cierreHora: form.cierreHora,
            cierreMinutos: form.cierreMinutos,
        };

        try {
            if (editing && id) {
                const { call } = actasService.update(id, payload);
                await call;
                setSuccess("Acta actualizada correctamente");
            } else {
                const { call } = actasService.create(payload);
                await call;
                setSuccess("Acta creada correctamente");
                setTimeout(() => navigate("/actas"), 1200);
            }
        } catch (err) {
            const apiMsg =
                axios.isAxiosError(err) &&
                (err.response?.data as { error?: string })?.error;
            setError(apiMsg || "No se pudo guardar el acta.");
        } finally {
            setLoading(false);
        }
    };

    if (fetchLoading) {
        return (
            <div className="p-6 md:p-8">
                <div className="max-w-4xl mx-auto text-gray-500">Cargando...</div>
            </div>
        );
    }

    return (
        <div className="p-6 md:p-8">
            <div className="max-w-4xl mx-auto space-y-4">
                <div className="flex items-center gap-3">
                    <button
                        type="button"
                        onClick={() => navigate("/actas")}
                        className="inline-flex items-center gap-1 px-3 py-1.5 border border-gray-300 dark:border-gray-600 text-sm rounded-md text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-800"
                    >
                        <ArrowLeft size={16} /> Volver
                    </button>
                    <h1 className="text-xl font-extrabold text-gray-900 dark:text-white tracking-tight">
                        {editing ? "Editar acta" : "Nueva acta"}
                    </h1>
                </div>

                <form
                    onSubmit={handleSubmit}
                    className="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm p-6 space-y-6"
                >
                    {error && (
                        <div className="flex items-start gap-2 p-3 rounded-md bg-red-50 border border-red-200 text-red-700 text-sm dark:bg-red-900/20 dark:border-red-800 dark:text-red-300">
                            <AlertCircle size={16} className="mt-0.5 shrink-0" />
                            <span>{error}</span>
                        </div>
                    )}
                    {success && (
                        <div className="flex items-start gap-2 p-3 rounded-md bg-green-50 border border-green-200 text-green-700 text-sm dark:bg-green-900/20 dark:border-green-800 dark:text-green-300">
                            <CheckCircle2 size={16} className="mt-0.5 shrink-0" />
                            <span>{success}</span>
                        </div>
                    )}

                    <section>
                        <div className={SECTION_TITLE}>Identificación</div>
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                            <div className="flex flex-col gap-1">
                                <label className={LABEL}>Código Acta</label>
                                <input
                                    className={INPUT}
                                    value={form.codigoActaText}
                                    onChange={setText("codigoActaText")}
                                    required
                                />
                            </div>
                            <div className="flex flex-col gap-1">
                                <label className={LABEL}>Código Recinto</label>
                                <input
                                    className={INPUT}
                                    value={form.codigoRecintoText}
                                    onChange={setText("codigoRecintoText")}
                                    required
                                />
                            </div>
                            <div className="flex flex-col gap-1">
                                <label className={LABEL}>Nro Mesa</label>
                                <input
                                    className={INPUT}
                                    type="number"
                                    min={1}
                                    value={form.nroMesa}
                                    onChange={setNum("nroMesa")}
                                    required
                                />
                            </div>
                            <div className="flex flex-col gap-1">
                                <label className={LABEL}>Estado</label>
                                <select
                                    className={INPUT}
                                    value={form.estado}
                                    onChange={setEstado}
                                >
                                    <option value="impresa">Impresa</option>
                                    <option value="transcrita">Transcrita</option>
                                    <option value="observada">Observada</option>
                                </select>
                            </div>
                        </div>
                    </section>

                    <section>
                        <div className={SECTION_TITLE}>Papeletas</div>
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            <div className="flex flex-col gap-1">
                                <label className={LABEL}>Papeletas en Ánfora</label>
                                <input
                                    className={INPUT}
                                    type="number"
                                    min={0}
                                    value={form.papeletasAnfora}
                                    onChange={setNum("papeletasAnfora")}
                                />
                            </div>
                            <div className="flex flex-col gap-1">
                                <label className={LABEL}>Papeletas No Usadas</label>
                                <input
                                    className={INPUT}
                                    type="number"
                                    min={0}
                                    value={form.papeletasNoUsadas}
                                    onChange={setNum("papeletasNoUsadas")}
                                />
                            </div>
                        </div>
                    </section>

                    <section>
                        <div className={SECTION_TITLE}>Votos por partido</div>
                        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                            {(["p1", "p2", "p3", "p4"] as const).map((p) => (
                                <div key={p} className="flex flex-col gap-1">
                                    <label className={LABEL}>{p.toUpperCase()}</label>
                                    <input
                                        className={INPUT}
                                        type="number"
                                        min={0}
                                        value={form[p]}
                                        onChange={setNum(p)}
                                    />
                                </div>
                            ))}
                        </div>
                    </section>

                    <section>
                        <div className={SECTION_TITLE}>Totales</div>
                        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                            <div className="flex flex-col gap-1">
                                <label className={LABEL}>Votos Válidos</label>
                                <input
                                    className={INPUT}
                                    type="number"
                                    min={0}
                                    value={form.votosValidos}
                                    onChange={setNum("votosValidos")}
                                />
                            </div>
                            <div className="flex flex-col gap-1">
                                <label className={LABEL}>Votos Nulos</label>
                                <input
                                    className={INPUT}
                                    type="number"
                                    min={0}
                                    value={form.votosNulos}
                                    onChange={setNum("votosNulos")}
                                />
                            </div>
                            <div className="flex flex-col gap-1">
                                <label className={LABEL}>Votos en Blanco</label>
                                <input
                                    className={INPUT}
                                    type="number"
                                    min={0}
                                    value={form.votosBlanco}
                                    onChange={setNum("votosBlanco")}
                                />
                            </div>
                        </div>
                    </section>

                    <section>
                        <div className={SECTION_TITLE}>Horario</div>
                        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                            <div className="flex flex-col gap-1">
                                <label className={LABEL}>Apertura — Hora</label>
                                <input
                                    className={INPUT}
                                    type="number"
                                    min={0}
                                    max={23}
                                    value={form.aperturaHora}
                                    onChange={setNum("aperturaHora")}
                                />
                            </div>
                            <div className="flex flex-col gap-1">
                                <label className={LABEL}>Apertura — Minutos</label>
                                <input
                                    className={INPUT}
                                    type="number"
                                    min={0}
                                    max={59}
                                    value={form.aperturaMinutos}
                                    onChange={setNum("aperturaMinutos")}
                                />
                            </div>
                            <div className="flex flex-col gap-1">
                                <label className={LABEL}>Cierre — Hora</label>
                                <input
                                    className={INPUT}
                                    type="number"
                                    min={0}
                                    max={23}
                                    value={form.cierreHora}
                                    onChange={setNum("cierreHora")}
                                />
                            </div>
                            <div className="flex flex-col gap-1">
                                <label className={LABEL}>Cierre — Minutos</label>
                                <input
                                    className={INPUT}
                                    type="number"
                                    min={0}
                                    max={59}
                                    value={form.cierreMinutos}
                                    onChange={setNum("cierreMinutos")}
                                />
                            </div>
                        </div>
                    </section>

                    <section>
                        <div className={SECTION_TITLE}>Observaciones</div>
                        <textarea
                            className={`${INPUT} w-full min-h-[80px] resize-y`}
                            value={form.observaciones}
                            onChange={setText("observaciones")}
                            placeholder="Observaciones opcionales..."
                        />
                    </section>

                    <div className="flex gap-3">
                        <button
                            type="submit"
                            disabled={loading}
                            className="bg-gray-900 hover:bg-black text-white text-sm font-semibold px-5 py-2 rounded-md disabled:opacity-60 disabled:cursor-not-allowed"
                        >
                            {loading ? "Guardando..." : editing ? "Guardar cambios" : "Crear acta"}
                        </button>
                        <button
                            type="button"
                            onClick={() => navigate("/actas")}
                            className="bg-gray-100 hover:bg-gray-200 text-gray-700 dark:bg-gray-700 dark:text-gray-200 dark:hover:bg-gray-600 text-sm px-5 py-2 rounded-md"
                        >
                            Cancelar
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};

export default ActaFormPage;
