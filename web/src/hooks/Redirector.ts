import { useCallback } from "react";

export function useRedirector() {
    return useCallback((url: string) => {
        window.location.href = url;
    }, []);
}
