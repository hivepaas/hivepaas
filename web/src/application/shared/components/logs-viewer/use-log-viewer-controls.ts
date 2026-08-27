import { type Dispatch, type SetStateAction, useCallback, useEffect, useState } from "react";

import { LOG_FONT_SIZES } from "./logs-viewer.constants";

export interface UseLogViewerControlsOptions {
    initialFontSizeIndex?: number;
}

export interface UseLogViewerControlsReturn {
    isFullscreen: boolean;
    setIsFullscreen: Dispatch<SetStateAction<boolean>>;
    toggleFullscreen: () => void;
    fontSizeIndex: number;
    fontSize: number;
    cycleFontSize: () => void;
}

export function useLogViewerControls(options: UseLogViewerControlsOptions = {}): UseLogViewerControlsReturn {
    const [isFullscreen, setIsFullscreen] = useState(false);
    const [fontSizeIndex, setFontSizeIndex] = useState(options.initialFontSizeIndex ?? 0);

    const fontSize = LOG_FONT_SIZES[fontSizeIndex] ?? 14;

    const toggleFullscreen = useCallback(() => {
        setIsFullscreen(current => !current);
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
        fontSizeIndex,
        fontSize,
        cycleFontSize,
    };
}
