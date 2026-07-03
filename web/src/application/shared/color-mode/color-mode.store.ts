import { type Draft, produce } from "immer";
import { create } from "zustand";

export const COLOR_MODES = ["dark", "light", "system"] as const;
export type ColorMode = (typeof COLOR_MODES)[number];

const STORAGE_KEY = "hivepaas-color-mode";
const DEFAULT_COLOR_MODE: ColorMode = "system";

function isColorMode(value: unknown): value is ColorMode {
    return typeof value === "string" && COLOR_MODES.includes(value as ColorMode);
}

class ColorModeStorage {
    persisted(): ColorMode {
        if (typeof window === "undefined") {
            return DEFAULT_COLOR_MODE;
        }

        const value = window.localStorage.getItem(STORAGE_KEY);
        if (isColorMode(value)) {
            return value;
        }

        if (value !== null) {
            window.localStorage.removeItem(STORAGE_KEY);
        }

        return DEFAULT_COLOR_MODE;
    }

    update(mode: ColorMode): void {
        if (typeof window === "undefined") {
            return;
        }

        window.localStorage.setItem(STORAGE_KEY, mode);
    }
}

interface State {
    mode: ColorMode;
}

interface Actions {
    setMode: (mode: ColorMode) => void;
}

const storage = new ColorModeStorage();

export const useColorModeContext = create<State & Actions>()(set => ({
    mode: storage.persisted(),

    setMode: mode => {
        set(
            produce((draft: Draft<State>) => {
                draft.mode = mode;
                storage.update(mode);
            }),
        );
    },
}));
