import { Navigate, Outlet } from "react-router-dom";
import { useAuthStore } from "../features/auth/hooks/useAuthStore";

export const PrivateRoutes = () => {
    const { jwt } = useAuthStore();
    const redirectTo = "/auth/login";

    if (!jwt) {
        return <Navigate to={redirectTo} replace />
    }
    return <Outlet />;
};
