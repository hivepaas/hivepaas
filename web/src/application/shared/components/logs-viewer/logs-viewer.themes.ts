import { useCallback, useEffect, useState } from "react";

import type { ITheme } from "@xterm/xterm";

export type TerminalThemeCategory = "gradient" | "classic";

export interface TerminalThemeDefinition {
    id: string;
    label: string;
    description: string;
    category: TerminalThemeCategory;
    accentColor: string;
    background: string;
    backgroundColor: string;
    theme: ITheme;
}

export const DEFAULT_TERMINAL_THEME: TerminalThemeDefinition = {
    id: "default",
    label: "Default Slate",
    description: "Modern slate dark theme",
    category: "classic",
    accentColor: "#38bdf8",
    background: "#0f172a",
    backgroundColor: "#0f172a",
    theme: {
        background: "#00000000",
        foreground: "#e2e8f0",
        cursor: "#38bdf8",
        cursorAccent: "#0f172a",
        selectionBackground: "#f59e0b",
        selectionForeground: "#000000",
        selectionInactiveBackground: "#d97706",
        black: "#0f172a",
        red: "#ef4444",
        green: "#22c55e",
        yellow: "#eab308",
        blue: "#3b82f6",
        magenta: "#a855f7",
        cyan: "#06b6d4",
        white: "#f8fafc",
        brightBlack: "#64748b",
        brightRed: "#f87171",
        brightGreen: "#4ade80",
        brightYellow: "#facc15",
        brightBlue: "#60a5fa",
        brightMagenta: "#c084fc",
        brightCyan: "#22d3ee",
        brightWhite: "#ffffff",
    },
};

export const TERMINAL_THEMES: Record<string, TerminalThemeDefinition> = {
    // Gradient Themes
    neonNebula: {
        id: "neonNebula",
        label: "Neon Nebula",
        description: "Coolify-inspired glowing cosmic nebula",
        category: "gradient",
        accentColor: "#f472b6",
        background: "linear-gradient(135deg, #0b3a4f 0%, #1e1e3b 25%, #761a60 50%, #3e103e 75%, #0b0e17 100%)",
        backgroundColor: "#0b0e17",
        theme: {
            background: "#00000000",
            foreground: "#f1f5f9",
            cursor: "#34d399",
            cursorAccent: "#0b0e17",
            selectionBackground: "#f472b6",
            selectionForeground: "#ffffff",
            selectionInactiveBackground: "#9d174d",
            black: "#0b0e17",
            red: "#fb7185",
            green: "#34d399",
            yellow: "#fbbf24",
            blue: "#60a5fa",
            magenta: "#f472b6",
            cyan: "#22d3ee",
            white: "#f8fafc",
            brightBlack: "#64748b",
            brightRed: "#fda4af",
            brightGreen: "#6ee7b7",
            brightYellow: "#fde047",
            brightBlue: "#93c5fd",
            brightMagenta: "#f9a8d4",
            brightCyan: "#67e8f9",
            brightWhite: "#ffffff",
        },
    },
    cyberpunk: {
        id: "cyberpunk",
        label: "Cyberpunk Synthwave",
        description: "High-contrast neon synthwave glow",
        category: "gradient",
        accentColor: "#facc15",
        background: "linear-gradient(135deg, #0a1e3d 0%, #2f0d4e 35%, #630c69 65%, #0d061c 100%)",
        backgroundColor: "#0d061c",
        theme: {
            background: "#00000000",
            foreground: "#fdf4ff",
            cursor: "#facc15",
            cursorAccent: "#0d061c",
            selectionBackground: "#e879f9",
            selectionForeground: "#000000",
            selectionInactiveBackground: "#a21caf",
            black: "#0d061c",
            red: "#ff007f",
            green: "#00ff9f",
            yellow: "#ffe600",
            blue: "#00b8ff",
            magenta: "#d600ff",
            cyan: "#00f0ff",
            white: "#ffffff",
            brightBlack: "#701a75",
            brightRed: "#ff54a6",
            brightGreen: "#50ffba",
            brightYellow: "#fff066",
            brightBlue: "#5cd3ff",
            brightMagenta: "#e866ff",
            brightCyan: "#66f6ff",
            brightWhite: "#ffffff",
        },
    },
    midnightAurora: {
        id: "midnightAurora",
        label: "Midnight Aurora",
        description: "Deep boreal emerald night",
        category: "gradient",
        accentColor: "#2dd4bf",
        background: "linear-gradient(135deg, #042e3b 0%, #055952 35%, #073d36 70%, #02121a 100%)",
        backgroundColor: "#02121a",
        theme: {
            background: "#00000000",
            foreground: "#e6fffa",
            cursor: "#2dd4bf",
            cursorAccent: "#02121a",
            selectionBackground: "#0f766e",
            selectionForeground: "#ffffff",
            selectionInactiveBackground: "#115e59",
            black: "#02121a",
            red: "#f87171",
            green: "#34d399",
            yellow: "#fbbf24",
            blue: "#38bdf8",
            magenta: "#a78bfa",
            cyan: "#2dd4bf",
            white: "#f0fdfa",
            brightBlack: "#334e48",
            brightRed: "#fca5a5",
            brightGreen: "#6ee7b7",
            brightYellow: "#fde047",
            brightBlue: "#7dd3fc",
            brightMagenta: "#c4b5fd",
            brightCyan: "#5eead4",
            brightWhite: "#ffffff",
        },
    },
    sunsetGlow: {
        id: "sunsetGlow",
        label: "Sunset Glow",
        description: "Warm dusk embers and twilight",
        category: "gradient",
        accentColor: "#fb923c",
        background: "linear-gradient(135deg, #300c3b 0%, #681842 35%, #882b19 70%, #16071c 100%)",
        backgroundColor: "#16071c",
        theme: {
            background: "#00000000",
            foreground: "#fff7ed",
            cursor: "#fb923c",
            cursorAccent: "#16071c",
            selectionBackground: "#ea580c",
            selectionForeground: "#ffffff",
            selectionInactiveBackground: "#9a3412",
            black: "#16071c",
            red: "#ef4444",
            green: "#4ade80",
            yellow: "#facc15",
            blue: "#60a5fa",
            magenta: "#e879f9",
            cyan: "#38bdf8",
            white: "#ffedd5",
            brightBlack: "#57302a",
            brightRed: "#f87171",
            brightGreen: "#86efac",
            brightYellow: "#fde047",
            brightBlue: "#93c5fd",
            brightMagenta: "#f0abfc",
            brightCyan: "#7dd3fc",
            brightWhite: "#ffffff",
        },
    },
    abyssalDeep: {
        id: "abyssalDeep",
        label: "Abyssal Deep",
        description: "Dark oceanic sapphire abyss",
        category: "gradient",
        accentColor: "#38bdf8",
        background: "linear-gradient(135deg, #033252 0%, #0a4773 35%, #062d47 70%, #010f1a 100%)",
        backgroundColor: "#010f1a",
        theme: {
            background: "#00000000",
            foreground: "#f0f9ff",
            cursor: "#38bdf8",
            cursorAccent: "#010f1a",
            selectionBackground: "#0284c7",
            selectionForeground: "#ffffff",
            selectionInactiveBackground: "#0369a1",
            black: "#010f1a",
            red: "#f87171",
            green: "#34d399",
            yellow: "#fbbf24",
            blue: "#38bdf8",
            magenta: "#818cf8",
            cyan: "#22d3ee",
            white: "#f0f9ff",
            brightBlack: "#1e3a5f",
            brightRed: "#fca5a5",
            brightGreen: "#6ee7b7",
            brightYellow: "#fde047",
            brightBlue: "#7dd3fc",
            brightMagenta: "#a5b4fc",
            brightCyan: "#67e8f9",
            brightWhite: "#ffffff",
        },
    },
    cosmicVelvet: {
        id: "cosmicVelvet",
        label: "Cosmic Velvet",
        description: "Rich plum and royal amethyst",
        category: "gradient",
        accentColor: "#c084fc",
        background: "linear-gradient(135deg, #1f093d 0%, #4c1266 45%, #661470 70%, #120621 100%)",
        backgroundColor: "#120621",
        theme: {
            background: "#00000000",
            foreground: "#faf5ff",
            cursor: "#c084fc",
            cursorAccent: "#120621",
            selectionBackground: "#9333ea",
            selectionForeground: "#ffffff",
            selectionInactiveBackground: "#7e22ce",
            black: "#120621",
            red: "#f43f5e",
            green: "#10b981",
            yellow: "#f59e0b",
            blue: "#8b5cf6",
            magenta: "#c084fc",
            cyan: "#06b6d4",
            white: "#faf5ff",
            brightBlack: "#4c1d63",
            brightRed: "#fb7185",
            brightGreen: "#34d399",
            brightYellow: "#fbbf24",
            brightBlue: "#a78bfa",
            brightMagenta: "#d8b4fe",
            brightCyan: "#22d3ee",
            brightWhite: "#ffffff",
        },
    },

    // Classic Themes
    default: DEFAULT_TERMINAL_THEME,
    dracula: {
        id: "dracula",
        label: "Dracula",
        description: "Classic dark theme with vibrant pastels",
        category: "classic",
        accentColor: "#bd93f9",
        background: "#282a36",
        backgroundColor: "#282a36",
        theme: {
            background: "#00000000",
            foreground: "#f8f8f2",
            cursor: "#f8f8f2",
            cursorAccent: "#282a36",
            selectionBackground: "#44475a",
            selectionForeground: "#f8f8f2",
            selectionInactiveBackground: "#3a3c4e",
            black: "#21222c",
            red: "#ff5555",
            green: "#50fa7b",
            yellow: "#f1fa8c",
            blue: "#bd93f9",
            magenta: "#ff79c6",
            cyan: "#8be9fd",
            white: "#f8f8f2",
            brightBlack: "#6272a4",
            brightRed: "#ff6e6e",
            brightGreen: "#69ff94",
            brightYellow: "#ffffa5",
            brightBlue: "#d6acff",
            brightMagenta: "#ff92df",
            brightCyan: "#a4ffff",
            brightWhite: "#ffffff",
        },
    },
    tokyoNight: {
        id: "tokyoNight",
        label: "Tokyo Night",
        description: "Clean dark theme celebrating downtown Tokyo",
        category: "classic",
        accentColor: "#7aa2f7",
        background: "#1a1b26",
        backgroundColor: "#1a1b26",
        theme: {
            background: "#00000000",
            foreground: "#a9b1d6",
            cursor: "#c0caf5",
            cursorAccent: "#1a1b26",
            selectionBackground: "#33467c",
            selectionForeground: "#c0caf5",
            selectionInactiveBackground: "#28345c",
            black: "#15161e",
            red: "#f7768e",
            green: "#9ece6a",
            yellow: "#e0af68",
            blue: "#7aa2f7",
            magenta: "#bb9af7",
            cyan: "#7dcfff",
            white: "#a9b1d6",
            brightBlack: "#414868",
            brightRed: "#f7768e",
            brightGreen: "#9ece6a",
            brightYellow: "#e0af68",
            brightBlue: "#7aa2f7",
            brightMagenta: "#bb9af7",
            brightCyan: "#7dcfff",
            brightWhite: "#c0caf5",
        },
    },
    oneDark: {
        id: "oneDark",
        label: "One Dark Pro",
        description: "Atom's iconic dark theme",
        category: "classic",
        accentColor: "#61afef",
        background: "#282c34",
        backgroundColor: "#282c34",
        theme: {
            background: "#00000000",
            foreground: "#abb2bf",
            cursor: "#528bff",
            cursorAccent: "#282c34",
            selectionBackground: "#3e4451",
            selectionForeground: "#ffffff",
            selectionInactiveBackground: "#353b45",
            black: "#1e2127",
            red: "#e06c75",
            green: "#98c379",
            yellow: "#d19a66",
            blue: "#61afef",
            magenta: "#c678dd",
            cyan: "#56b6c2",
            white: "#abb2bf",
            brightBlack: "#5c6370",
            brightRed: "#e06c75",
            brightGreen: "#98c379",
            brightYellow: "#e5c07b",
            brightBlue: "#61afef",
            brightMagenta: "#c678dd",
            brightCyan: "#56b6c2",
            brightWhite: "#ffffff",
        },
    },
    nord: {
        id: "nord",
        label: "Nord",
        description: "Arctic, north-bluish color palette",
        category: "classic",
        accentColor: "#88c0d0",
        background: "#2e3440",
        backgroundColor: "#2e3440",
        theme: {
            background: "#00000000",
            foreground: "#d8dee9",
            cursor: "#eceff4",
            cursorAccent: "#2e3440",
            selectionBackground: "#434c5e",
            selectionForeground: "#eceff4",
            selectionInactiveBackground: "#3b4252",
            black: "#2e3440",
            red: "#bf616a",
            green: "#a3be8c",
            yellow: "#ebcb8b",
            blue: "#81a1c1",
            magenta: "#b48ead",
            cyan: "#88c0d0",
            white: "#e5e9f0",
            brightBlack: "#4c566a",
            brightRed: "#d08770",
            brightGreen: "#a3be8c",
            brightYellow: "#ebcb8b",
            brightBlue: "#88c0d0",
            brightMagenta: "#b48ead",
            brightCyan: "#8fbcbb",
            brightWhite: "#eceff4",
        },
    },
    catppuccin: {
        id: "catppuccin",
        label: "Catppuccin Mocha",
        description: "Soothing pastel dark theme",
        category: "classic",
        accentColor: "#cba6f7",
        background: "#1e1e2e",
        backgroundColor: "#1e1e2e",
        theme: {
            background: "#00000000",
            foreground: "#cdd6f4",
            cursor: "#f5e0dc",
            cursorAccent: "#1e1e2e",
            selectionBackground: "#585b70",
            selectionForeground: "#cdd6f4",
            selectionInactiveBackground: "#45475a",
            black: "#181825",
            red: "#f38ba8",
            green: "#a6e3a1",
            yellow: "#f9e2af",
            blue: "#89b4fa",
            magenta: "#f5c2e7",
            cyan: "#94e2d5",
            white: "#bac2de",
            brightBlack: "#585b70",
            brightRed: "#f38ba8",
            brightGreen: "#a6e3a1",
            brightYellow: "#f9e2af",
            brightBlue: "#89b4fa",
            brightMagenta: "#cba6f7",
            brightCyan: "#89dceb",
            brightWhite: "#a6adc8",
        },
    },
    githubDark: {
        id: "githubDark",
        label: "GitHub Dark",
        description: "GitHub's official dark theme",
        category: "classic",
        accentColor: "#58a6ff",
        background: "#0d1117",
        backgroundColor: "#0d1117",
        theme: {
            background: "#00000000",
            foreground: "#c9d1d9",
            cursor: "#58a6ff",
            cursorAccent: "#0d1117",
            selectionBackground: "#1f6feb",
            selectionForeground: "#ffffff",
            selectionInactiveBackground: "#163c76",
            black: "#010409",
            red: "#ff7b72",
            green: "#3fb950",
            yellow: "#d29922",
            blue: "#58a6ff",
            magenta: "#bc8cff",
            cyan: "#39c5cf",
            white: "#b1bac4",
            brightBlack: "#6e7681",
            brightRed: "#ffa198",
            brightGreen: "#56d364",
            brightYellow: "#e3b341",
            brightBlue: "#79c0ff",
            brightMagenta: "#d2a8ff",
            brightCyan: "#56d4dd",
            brightWhite: "#f0f6fc",
        },
    },
};

export const DEFAULT_TERMINAL_THEME_ID = "default";
export const TERMINAL_THEME_STORAGE_KEY = "hivepaas_terminal_theme";
const THEME_CHANGE_EVENT = "hivepaas_terminal_theme_change";

/**
 * Safely retrieve theme definition with graceful fallback to DEFAULT_TERMINAL_THEME.
 * Guaranteed to never return undefined or throw.
 */
export function getTerminalTheme(id?: string | null): TerminalThemeDefinition {
    if (id && TERMINAL_THEMES[id]) {
        return TERMINAL_THEMES[id];
    }
    return DEFAULT_TERMINAL_THEME;
}

/**
 * Safely read stored theme ID from localStorage with validation and error protection.
 * If data is corrupted, missing, or outdated, it falls back to DEFAULT_TERMINAL_THEME_ID.
 */
export function getStoredTerminalThemeId(): string {
    if (typeof window === "undefined") {
        return DEFAULT_TERMINAL_THEME_ID;
    }

    try {
        const saved = window.localStorage.getItem(TERMINAL_THEME_STORAGE_KEY);
        if (saved && TERMINAL_THEMES[saved]) {
            return saved;
        }

        // If stored value is outdated / invalid, clean it up safely
        if (saved && !TERMINAL_THEMES[saved]) {
            try {
                window.localStorage.removeItem(TERMINAL_THEME_STORAGE_KEY);
            } catch {
                // Ignore storage cleanup error
            }
        }
    } catch {
        // Catch Private Browsing, disabled cookies/storage, or SecurityError
    }

    return DEFAULT_TERMINAL_THEME_ID;
}

/**
 * Safely persist theme ID to localStorage and broadcast change to other components/tabs.
 */
export function setStoredTerminalThemeId(themeId: string): boolean {
    if (typeof window === "undefined" || !TERMINAL_THEMES[themeId]) {
        return false;
    }

    try {
        window.localStorage.setItem(TERMINAL_THEME_STORAGE_KEY, themeId);
        window.dispatchEvent(new CustomEvent(THEME_CHANGE_EVENT, { detail: themeId }));
        return true;
    } catch {
        // Handle QuotaExceededError or storage disabled without throwing
        return false;
    }
}

export interface UseTerminalThemeReturn {
    themeId: string;
    currentTheme: TerminalThemeDefinition;
    changeTheme: (newThemeId: string) => void;
    themes: TerminalThemeDefinition[];
}

/**
 * Global hook to manage terminal theme across the application.
 * Automatically synchronizes across components and browser tabs with zero crash guarantee.
 */
export function useTerminalTheme(initialThemeId?: string): UseTerminalThemeReturn {
    const [themeId, setThemeId] = useState<string>(() => {
        if (initialThemeId && TERMINAL_THEMES[initialThemeId]) {
            return initialThemeId;
        }
        return getStoredTerminalThemeId();
    });

    const changeTheme = useCallback((newThemeId: string) => {
        if (TERMINAL_THEMES[newThemeId]) {
            setThemeId(newThemeId);
            setStoredTerminalThemeId(newThemeId);
        }
    }, []);

    useEffect(() => {
        if (typeof window === "undefined") {
            return undefined;
        }

        function handleStorageChange(event: StorageEvent) {
            if (event.key === TERMINAL_THEME_STORAGE_KEY) {
                const newThemeId =
                    event.newValue && TERMINAL_THEMES[event.newValue] ? event.newValue : DEFAULT_TERMINAL_THEME_ID;
                setThemeId(newThemeId);
            }
        }

        function handleCustomThemeChange(event: Event) {
            const customEvent = event as CustomEvent<string>;
            if (customEvent.detail && TERMINAL_THEMES[customEvent.detail]) {
                setThemeId(customEvent.detail);
            }
        }

        window.addEventListener("storage", handleStorageChange);
        window.addEventListener(THEME_CHANGE_EVENT, handleCustomThemeChange);

        return () => {
            window.removeEventListener("storage", handleStorageChange);
            window.removeEventListener(THEME_CHANGE_EVENT, handleCustomThemeChange);
        };
    }, []);

    const currentTheme: TerminalThemeDefinition = getTerminalTheme(themeId);

    return {
        themeId,
        currentTheme,
        changeTheme,
        themes: Object.values(TERMINAL_THEMES),
    };
}
