import { useEffect } from "react";

import { type ColorMode, useColorModeContext } from "./color-mode.store";

const DARK_CLASS = "dark";
const DARK_QUERY = "(prefers-color-scheme: dark)";

/**
 * Applies the app's own dark palette by toggling the `dark` class that the
 * Tailwind variant keys off - see `@custom-variant dark` in index.css.
 *
 * This replaces Darkreader, which used to synthesise dark mode at runtime from
 * the light stylesheet. None of the fixes it carried had to be ported: each one
 * - the logo, the tab indicator, the terminal frames - was an instruction to
 * leave those colours alone, and without Darkreader they simply keep the colours
 * they are authored with.
 *
 * The class is also set by an inline script in index.html before React mounts,
 * so a dark session does not flash light on load. This effect keeps it in sync
 * afterwards.
 */
function applyColorMode(mode: ColorMode, prefersDark: boolean) {
    const isDark = mode === "dark" || (mode === "system" && prefersDark);
    document.documentElement.classList.toggle(DARK_CLASS, isDark);
}

export function ColorModeEffect() {
    const mode = useColorModeContext(state => state.mode);

    useEffect(() => {
        const query = window.matchMedia(DARK_QUERY);

        applyColorMode(mode, query.matches);

        // Only "system" follows the OS, and it has to keep following it - the
        // user can flip their OS theme while this tab is open.
        if (mode !== "system") {
            return;
        }

        const onChange = (event: MediaQueryListEvent) => {
            applyColorMode(mode, event.matches);
        };
        query.addEventListener("change", onChange);

        return () => {
            query.removeEventListener("change", onChange);
        };
    }, [mode]);

    return null;
}
