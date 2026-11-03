import { type Dispatch, type SetStateAction, useCallback, useEffect, useState } from "react";

import { LOG_FONT_SIZES } from "./logs-viewer.constants";
import { type TerminalThemeDefinition, useTerminalTheme } from "./logs-viewer.themes";

export interface UseLogViewerControlsOptions {
    initialFontSizeIndex?: number;
    initialThemeId?: string;
}

export interface UseLogViewerControlsReturn {
    isFullscreen: boolean;
    setIsFullscreen: Dispatch<SetStateAction<boolean>>;
    toggleFullscreen: () => void;
    isFullView: boolean;
    setIsFullView: Dispatch<SetStateAction<boolean>>;
    toggleFullView: () => void;
    isFullHeight: boolean;
    setIsFullHeight: Dispatch<SetStateAction<boolean>>;
    toggleFullHeight: () => void;
    fontSizeIndex: number;
    fontSize: number;
    cycleFontSize: () => void;
    themeId: string;
    currentTheme: TerminalThemeDefinition;
    changeTheme: (themeId: string) => void;
}

export function useLogViewerControls(options: UseLogViewerControlsOptions = {}): UseLogViewerControlsReturn {
    const { themeId, currentTheme, changeTheme } = useTerminalTheme(options.initialThemeId);
    const [isFullscreen, setIsFullscreen] = useState(false);
    const [isFullView, setIsFullView] = useState(false);
    const [isFullHeight, setIsFullHeight] = useState(() => {
        if (typeof window !== "undefined") {
            return window.innerWidth < 768;
        }
        return false;
    });
    const [fontSizeIndex, setFontSizeIndex] = useState(options.initialFontSizeIndex ?? 0);

    const fontSize = LOG_FONT_SIZES[fontSizeIndex] ?? 14;

    const toggleFullscreen = useCallback(() => {
        setIsFullscreen(current => !current);
    }, []);

    const toggleFullView = useCallback(() => {
        setIsFullView(current => !current);
    }, []);

    const toggleFullHeight = useCallback(() => {
        setIsFullHeight(current => !current);
    }, []);

    const cycleFontSize = useCallback(() => {
        setFontSizeIndex(current => (current + 1) % LOG_FONT_SIZES.length);
    }, []);

    useEffect(() => {
        if (!isFullscreen) {
            return;
        }

        function handleKeyDown(event: KeyboardEvent) {
            if (event.key === "Escape") {
                setIsFullscreen(false);
            }
        }

        window.addEventListener("keydown", handleKeyDown);

        return () => {
            window.removeEventListener("keydown", handleKeyDown);
        };
    }, [isFullscreen]);

    return {
        isFullscreen,
        setIsFullscreen,
        toggleFullscreen,
        isFullView,
        setIsFullView,
        toggleFullView,
        isFullHeight,
        setIsFullHeight,
        toggleFullHeight,
        fontSizeIndex,
        fontSize,
        cycleFontSize,
        themeId,
        currentTheme,
        changeTheme,
    };
}
