import { useState } from "react";
import { NavLink, useNavigate } from "react-router-dom";
import { Sun, Moon, User, LogOut, ChevronLeft, ChevronRight, LayoutDashboard, AlertTriangle, MessageSquare } from "lucide-react";
import { useAuthStore } from "../features/auth/hooks/useAuthStore";
import { useDarkMode } from "../hooks/useDarkMode";
import clsx from "clsx";

export const Sidebar = () => {
    const { isDark, toggleDark } = useDarkMode();
    const { username, role, clearAuthUser } = useAuthStore();
    const navigate = useNavigate();
    const [collapsed, setCollapsed] = useState(false);
    const [menuOpen, setMenuOpen] = useState(false);

    const handleLogout = () => {
        clearAuthUser();
        navigate("/auth/login");
    };

    return (
        <aside className={clsx(
            "hidden md:flex bg-white dark:bg-gray-900 border-r border-gray-200 dark:border-gray-800 transition-all duration-300 flex-col relative z-50 h-screen",
            collapsed ? "w-20" : "w-64"
        )}>
            {/* Toggle Button for Desktop */}
            <button
                onClick={() => setCollapsed(!collapsed)}
                className="absolute -right-3 top-8 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-full p-1 text-gray-500 hover:text-blue-600 shadow-sm"
            >
                {collapsed ? <ChevronRight size={16} /> : <ChevronLeft size={16} />}
            </button>

            {/* Logo */}
            <div className="flex items-center justify-between p-6">
                <div className="flex items-center gap-3">
                    <div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center text-white font-bold shrink-0">
                        CE
                    </div>
                    {!collapsed && (
                        <span className="font-bold text-lg text-gray-900 dark:text-white truncate">
                            Cómputo Electoral
                        </span>
                    )}
                </div>
            </div>

            {/* Navigation */}
            <nav className="flex-1 px-4 space-y-2 mt-4">
                <NavLink
                    to="/dashboard"
                    className={({ isActive }) => clsx(
                        "flex items-center gap-3 px-3 py-3 rounded-lg transition-colors group relative",
                        isActive ? "bg-blue-50 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400 font-medium" : "text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800 hover:text-gray-900 dark:hover:text-gray-200"
                    )}
                >
                    <LayoutDashboard size={20} className="shrink-0" />
                    {!collapsed && <span>Dashboard</span>}
                    {collapsed && (
                        <div className="absolute left-full ml-4 px-2 py-1 bg-gray-800 text-white text-xs rounded opacity-0 group-hover:opacity-100 pointer-events-none whitespace-nowrap z-50 transition-opacity">
                            Dashboard
                        </div>
                    )}
                </NavLink>
                <NavLink
                    to="/inconsistencias"
                    className={({ isActive }) => clsx(
                        "flex items-center gap-3 px-3 py-3 rounded-lg transition-colors group relative",
                        isActive ? "bg-blue-50 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400 font-medium" : "text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800 hover:text-gray-900 dark:hover:text-gray-200"
                    )}
                >
                    <AlertTriangle size={20} className="shrink-0" />
                    {!collapsed && <span>Inconsistencias</span>}
                    {collapsed && (
                        <div className="absolute left-full ml-4 px-2 py-1 bg-gray-800 text-white text-xs rounded opacity-0 group-hover:opacity-100 pointer-events-none whitespace-nowrap z-50 transition-opacity">
                            Inconsistencias
                        </div>
                    )}
                </NavLink>
                <NavLink
                    to="/sms"
                    className={({ isActive }) => clsx(
                        "flex items-center gap-3 px-3 py-3 rounded-lg transition-colors group relative",
                        isActive ? "bg-blue-50 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400 font-medium" : "text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800 hover:text-gray-900 dark:hover:text-gray-200"
                    )}
                >
                    <MessageSquare size={20} className="shrink-0" />
                    {!collapsed && <span>SMS</span>}
                    {collapsed && (
                        <div className="absolute left-full ml-4 px-2 py-1 bg-gray-800 text-white text-xs rounded opacity-0 group-hover:opacity-100 pointer-events-none whitespace-nowrap z-50 transition-opacity">
                            SMS
                        </div>
                    )}
                </NavLink>

                {/* <NavLink
                    to="/actas"
                    className={({ isActive }) => clsx(
                        "flex items-center gap-3 px-3 py-3 rounded-lg transition-colors group relative",
                        isActive ? "bg-blue-50 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400 font-medium" : "text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800 hover:text-gray-900 dark:hover:text-gray-200"
                    )}
                >
                    <FileStack size={20} className="shrink-0" />
                    {!collapsed && <span>Actas</span>}
                    {collapsed && (
                        <div className="absolute left-full ml-4 px-2 py-1 bg-gray-800 text-white text-xs rounded opacity-0 group-hover:opacity-100 pointer-events-none whitespace-nowrap z-50 transition-opacity">
                            Actas
                        </div>
                    )}
                </NavLink> */}
            </nav>

            {/* Footer / Actions */}
            <div className="p-4 border-t border-gray-200 dark:border-gray-800 space-y-4">
                {/* Theme Toggle */}
                <button
                    onClick={toggleDark}
                    className={clsx(
                        "flex items-center gap-3 px-3 py-2 w-full text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors group relative",
                        collapsed && "justify-center"
                    )}
                >
                    {isDark ? <Sun size={20} className="shrink-0" /> : <Moon size={20} className="shrink-0" />}
                    {!collapsed && <span className="text-sm font-medium">Cambiar Tema</span>}
                    {collapsed && (
                        <div className="absolute left-full ml-4 px-2 py-1 bg-gray-800 text-white text-xs rounded opacity-0 group-hover:opacity-100 pointer-events-none whitespace-nowrap z-50 transition-opacity">
                            Tema {isDark ? "Claro" : "Oscuro"}
                        </div>
                    )}
                </button>

                {/* User Profile */}
                <div className="relative">
                    <button
                        onClick={() => setMenuOpen(!menuOpen)}
                        className={clsx(
                            "flex items-center gap-3 px-3 py-2 w-full hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors group relative",
                            collapsed && "justify-center"
                        )}
                    >
                        <div className="w-8 h-8 bg-blue-100 dark:bg-blue-900 text-blue-600 dark:text-blue-300 rounded-full flex items-center justify-center shrink-0">
                            <User size={18} />
                        </div>
                        {!collapsed && (
                            <div className="text-left flex-1 min-w-0">
                                <p className="text-sm font-semibold text-gray-900 dark:text-white truncate">{username || "Usuario"}</p>
                                <p className="text-xs text-gray-500 dark:text-gray-400 truncate">{role || "Rol"}</p>
                            </div>
                        )}
                        {collapsed && (
                            <div className="absolute left-full ml-4 px-2 py-1 bg-gray-800 text-white text-xs rounded opacity-0 group-hover:opacity-100 pointer-events-none whitespace-nowrap z-50 transition-opacity">
                                Perfil
                            </div>
                        )}
                    </button>

                    {menuOpen && (
                        <>
                            <div className="fixed inset-0 z-40" onClick={() => setMenuOpen(false)}></div>
                            <div className={clsx(
                                "absolute bottom-full mb-2 bg-white dark:bg-gray-800 rounded-xl shadow-lg border border-gray-100 dark:border-gray-700 overflow-hidden z-50",
                                collapsed ? "left-12 w-48" : "left-0 w-full"
                            )}>
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
        </aside>
    );
};
