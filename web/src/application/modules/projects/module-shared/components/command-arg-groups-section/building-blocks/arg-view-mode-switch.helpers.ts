export type ArgViewMode = "grid" | "list";

export interface ArgViewModeSwitchProps {
    value: ArgViewMode;
    onChange: (mode: ArgViewMode) => void;
    disabled?: boolean;
    className?: string;
}

export const ARG_VIEW_MODE_STORAGE_KEY = "hivepaas_command_args_view_mode";

export function getStoredArgViewMode(): ArgViewMode {
    if (typeof window === "undefined") {
        return "grid";
    }
    try {
        const saved = window.localStorage.getItem(ARG_VIEW_MODE_STORAGE_KEY);
        if (saved === "list" || saved === "grid") {
            return saved;
        }
    } catch {
        // Ignore storage read error
    }
    return "grid";
}

export function setStoredArgViewMode(mode: ArgViewMode): void {
    if (typeof window === "undefined") {
        return;
    }
    try {
        window.localStorage.setItem(ARG_VIEW_MODE_STORAGE_KEY, mode);
    } catch {
        // Ignore storage write error
    }
}
