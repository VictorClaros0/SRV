import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuthStore } from "../hooks/useAuthStore";
import { authService } from "../services/authService";
import { Lock, User, AlertCircle } from "lucide-react";
import axios from "axios";

export const LoginPage = () => {
    const navigate = useNavigate();
    const { setAuthUser } = useAuthStore();
    const [username, setUsername] = useState("admin");
    const [password, setPassword] = useState("admin123");
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const handleLogin = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        setError(null);
        try {
            const { call } = authService.login({
                usuario: username,
                contrasena: password,
            });
            const { data: response } = await call;
            setAuthUser({ response, username });
            navigate("/dashboard");
        } catch (err) {
            if (axios.isAxiosError(err)) {
                const apiMsg = (err.response?.data as { error?: string })?.error;
                setError(apiMsg ?? "No se pudo iniciar sesión. Verifica que la API esté disponible.");
            } else {
                setError("Error inesperado al iniciar sesión.");
            }
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900 p-4">
            <div className="w-full max-w-md bg-white dark:bg-gray-800 rounded-2xl shadow-xl border border-gray-100 dark:border-gray-700 p-8">
                <div className="text-center mb-8">
                    <div className="w-16 h-16 bg-blue-600 rounded-2xl flex items-center justify-center text-white font-bold text-2xl mx-auto mb-4">
                        CE
                    </div>
                    <h2 className="text-2xl font-bold text-gray-900 dark:text-white">Cómputo Electoral</h2>
                    <p className="text-gray-500 dark:text-gray-400 mt-2 text-sm">Ingresa tus credenciales para acceder al sistema</p>
                </div>

                <form onSubmit={handleLogin} className="space-y-6">
                    <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Usuario</label>
                        <div className="relative">
                            <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                                <User size={18} className="text-gray-400" />
                            </div>
                            <input
                                type="text"
                                required
                                value={username}
                                onChange={(e) => setUsername(e.target.value)}
                                className="pl-10 bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white"
                                placeholder="tu.usuario"
                            />
                        </div>
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Contraseña</label>
                        <div className="relative">
                            <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                                <Lock size={18} className="text-gray-400" />
                            </div>
                            <input
                                type="password"
                                required
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                                className="pl-10 bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white"
                                placeholder="••••••••"
                            />
                        </div>
                    </div>

                    {error && (
                        <div className="flex items-start gap-2 p-3 rounded-lg bg-red-50 border border-red-200 text-red-700 text-sm dark:bg-red-900/20 dark:border-red-800 dark:text-red-300">
                            <AlertCircle size={16} className="mt-0.5 shrink-0" />
                            <span>{error}</span>
                        </div>
                    )}

                    <button
                        type="submit"
                        disabled={loading}
                        className="w-full text-white bg-blue-600 hover:bg-blue-700 focus:ring-4 focus:ring-blue-300 font-medium rounded-lg text-sm px-5 py-3 transition-colors dark:bg-blue-600 dark:hover:bg-blue-700 focus:outline-none flex justify-center items-center gap-2 disabled:opacity-60 disabled:cursor-not-allowed"
                    >
                        {loading ? <span className="animate-pulse">Verificando...</span> : "Ingresar al Sistema"}
                    </button>
                </form>
            </div>
        </div>
    );
};
