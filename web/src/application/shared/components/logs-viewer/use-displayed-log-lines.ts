import { useMemo, useRef } from "react";

import type { LogsViewerFrame } from "./logs-viewer.types";
import { buildDisplayedLogFrames, getAnsiLogLines, getPlainLogLines } from "./logs-viewer.utils";

interface DisplayedLogLinesCache {
    framesLength: number;
    showDebugLogs: boolean;
    showTimestamps: boolean;
    plainLines: string[];
    ansiLines: string[];
}

function buildDisplayedLogLinesCache(
    frames: LogsViewerFrame[],
    showDebugLogs: boolean,
    showTimestamps: boolean,
): DisplayedLogLinesCache {
    const displayedFrames = buildDisplayedLogFrames(frames, showDebugLogs);

    return {
        framesLength: frames.length,
        showDebugLogs,
        showTimestamps,
        plainLines: displayedFrames.flatMap(frame => getPlainLogLines(frame, showTimestamps)),
        ansiLines: displayedFrames.flatMap(frame => getAnsiLogLines(frame, showTimestamps)),
    };
}

export function useDisplayedLogLines(
    frames: LogsViewerFrame[],
    showDebugLogs: boolean,
    showTimestamps: boolean,
): DisplayedLogLinesCache {
    const cacheRef = useRef<DisplayedLogLinesCache | null>(null);

    return useMemo(() => {
        const cache = cacheRef.current;

        if (!cache) {
            const next = buildDisplayedLogLinesCache(frames, showDebugLogs, showTimestamps);
            cacheRef.current = next;
            return next;
        }

        if (cache.showDebugLogs !== showDebugLogs || cache.showTimestamps !== showTimestamps) {
            const next = buildDisplayedLogLinesCache(frames, showDebugLogs, showTimestamps);
            cacheRef.current = next;
            return next;
        }

        if (frames.length < cache.framesLength) {
            const next = buildDisplayedLogLinesCache(frames, showDebugLogs, showTimestamps);
            cacheRef.current = next;
            return next;
        }

        if (frames.length === cache.framesLength) {
            return cache;
        }

        const appendedFrames = buildDisplayedLogFrames(frames.slice(cache.framesLength), showDebugLogs);
        const next: DisplayedLogLinesCache = {
            framesLength: frames.length,
            showDebugLogs,
            showTimestamps,
            plainLines: [
                ...cache.plainLines,
                ...appendedFrames.flatMap(frame => getPlainLogLines(frame, showTimestamps)),
            ],
            ansiLines: [...cache.ansiLines, ...appendedFrames.flatMap(frame => getAnsiLogLines(frame, showTimestamps))],
        };
        cacheRef.current = next;
        return next;
    }, [frames, showDebugLogs, showTimestamps]);
}
