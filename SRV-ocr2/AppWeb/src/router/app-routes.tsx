import { Suspense } from "react";
import { useRoutes } from "react-router-dom";
import { routesConfig } from "./routes-config";

export const AppRoutes = () => {
    const routing = useRoutes(routesConfig);
    
    return (
        <Suspense fallback={
            <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900">
                <div className="w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin"></div>
            </div>
        }>
            {routing}
        </Suspense>
    );
};
