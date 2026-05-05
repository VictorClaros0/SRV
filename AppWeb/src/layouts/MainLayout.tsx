import { useState } from "react";
import { Outlet, NavLink, useNavigate } from "react-router-dom";
import { Sidebar } from "../components/Sidebar";
import { LayoutDashboard, AlertTriangle, MessageSquare, User, LogOut, Sun, Moon } from "lucide-react";
import { useAuthStore } from "../features/auth/hooks/useAuthStore";
import { useDarkMode } from "../hooks/useDarkMode";
import clsx from "clsx";

export const MainLayout = () => {
    const { isDark, toggleDark } = useDarkMode();
    const { username, role, clearAuthUser } = useAuthStore();
    const navigate = useNavigate();
    const [menuOpen, setMenuOpen] = useState(false);

    const handleLogout = () => {
        clearAuthUser();
        navigate("/auth/login");
    };

    return (
        <div className="min-h-screen bg-gray-50 dark:bg-gray-950 text-gray-900 dark:text-gray-100 transition-colors duration-200 flex flex-col md:flex-row overflow-hidden relative">
            {/* Mobile Header (Top) */}
            <header className="md:hidden flex items-center justify-between p-4 bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800 shadow-sm z-30 shrink-0">
                <div className="flex items-center gap-2">
                    <div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center text-white font-bold">CE</div>
                    <span className="font-bold text-gray-900 dark:text-white">Cómputo Electoral</span>
                </div>

                <div className="flex items-center gap-3">
                    <button onClick={toggleDark} className="p-2 text-gray-500 dark:text-gray-400 bg-gray-100 dark:bg-gray-800 rounded-full">
                        {isDark ? <Sun size={18} /> : <Moon size={18} />}
                    </button>

                    <div className="relative">
                        <button onClick={() => setMenuOpen(!menuOpen)} className="w-9 h-9 bg-blue-100 dark:bg-blue-900/50 text-blue-600 dark:text-blue-300 rounded-full flex items-center justify-center">
                            <User size={18} />
                        </button>

                        {menuOpen && (
                            <>
                                <div className="fixed inset-0 z-40" onClick={() => setMenuOpen(false)}></div>
                                <div className="absolute right-0 top-full mt-2 w-48 bg-white dark:bg-gray-800 rounded-xl shadow-lg border border-gray-100 dark:border-gray-700 overflow-hidden z-50">
                                    <div className="p-4 border-b border-gray-100 dark:border-gray-700">
                                        <p className="text-sm font-semibold text-gray-900 dark:text-white truncate">{username || "Usuario"}</p>
                                        <p className="text-xs text-gray-500 dark:text-gray-400 truncate">{role || "Rol"}</p>
                                    </div>
                                    <div className="p-2">
                                        <button onClick={handleLogout} className="w-full flex items-center gap-2 px-3 py-2 text-sm text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/30 rounded-lg transition-colors">
                                            <LogOut size={16} /> Cerrar Sesión
                                        </button>
                                    </div>
                                </div>
                            </>
                        )}
                    </div>
                </div>
            </header>

            {/* Desktop Sidebar */}
            <Sidebar />

            {/* Main Content Area */}
            <main className="flex-1 overflow-y-auto overflow-x-hidden relative h-[calc(100vh-140px)] md:h-screen pb-20 md:pb-0">
                <Outlet />
            </main>

            {/* Mobile Bottom Navigation (Bottom) */}
            <nav className="md:hidden fixed bottom-0 left-0 right-0 bg-white dark:bg-gray-900 border-t border-gray-200 dark:border-gray-800 z-40 flex items-center justify-around px-2 py-2 shadow-[0_-4px_6px_-1px_rgba(0,0,0,0.05)] dark:shadow-none pb-[calc(8px+env(safe-area-inset-bottom))]">
                <NavLink
                    to="/dashboard"
                    className={({ isActive }) => clsx(
                        "flex flex-col items-center gap-1 px-3 py-2 rounded-xl transition-colors",
                        isActive ? "text-blue-600 dark:text-blue-400" : "text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200"
                    )}
                >
                    <LayoutDashboard size={22} />
                    <span className="text-[10px] font-medium">Dashboard</span>
                </NavLink>
                <NavLink
                    to="/inconsistencias"
                    className={({ isActive }) => clsx(
                        "flex flex-col items-center gap-1 px-3 py-2 rounded-xl transition-colors",
                        isActive ? "text-blue-600 dark:text-blue-400" : "text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200"
                    )}
                >
                    <AlertTriangle size={22} />
                    <span className="text-[10px] font-medium">Inconsist.</span>
                </NavLink>
                <NavLink
                    to="/sms"
                    className={({ isActive }) => clsx(
                        "flex flex-col items-center gap-1 px-3 py-2 rounded-xl transition-colors",
                        isActive ? "text-blue-600 dark:text-blue-400" : "text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200"
                    )}
                >
                    <MessageSquare size={22} />
                    <span className="text-[10px] font-medium">SMS</span>
                </NavLink>

                {/* <NavLink
                    to="/actas"
                    className={({ isActive }) => clsx(
                        "flex flex-col items-center gap-1 px-3 py-2 rounded-xl transition-colors",
                        isActive ? "text-blue-600 dark:text-blue-400" : "text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200"
                    )}
                >
                    <FileStack size={22} />
                    <span className="text-[10px] font-medium">Actas</span>
                </NavLink> */}
            </nav>
        </div>
    );
};
