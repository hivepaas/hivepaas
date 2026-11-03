export const LOG_FONT_SIZES = [14, 16, 18, 20] as const;
export const DEFAULT_LOG_VIEWER_HEIGHT = "clamp(350px, calc(100vh - 400px), 2000px)";
export const DEFAULT_DOWNLOAD_FILE_NAME = "logs.txt";
export const TERMINAL_SCROLLBACK = 50_000;

export const LOG_SEARCH_DECORATIONS = {
    matchBackground: "#1e3a8a",
    activeMatchBackground: "#eab308",
    matchBorder: "#3b82f6",
    activeMatchBorder: "#fde047",
    matchOverviewRuler: "#3b82f6",
    activeMatchColorOverviewRuler: "#eab308",
} as const;
