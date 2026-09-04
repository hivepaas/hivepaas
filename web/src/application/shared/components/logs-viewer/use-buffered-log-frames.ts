import { type SetStateAction, useCallback, useEffect, useRef, useState } from "react";

import type { LogsViewerFrame } from "./logs-viewer.types";

const DEFAULT_MAX_FRAMES = 10_000;
const DEFAULT_FLUSH_INTERVAL_MS = 100;

function capFrames(frames: LogsViewerFrame[], maxFrames: number): LogsViewerFrame[] {
    if (frames.length <= maxFrames) {
        return frames;
    }

    return frames.slice(frames.length - maxFrames);
}

function resolveSetStateAction<T>(action: SetStateAction<T>, current: T): T {
    return typeof action === "function" ? (action as (value: T) => T)(current) : action;
}

interface UseBufferedLogFramesOptions {
    maxFrames?: number;
    flushIntervalMs?: number;
    frames?: LogsViewerFrame[];
    onFramesChange?: (action: SetStateAction<LogsViewerFrame[]>) => void;
}

interface BufferedLogFramesResult {
    frames: LogsViewerFrame[];
    appendFrames: (frames: LogsViewerFrame[]) => void;
    reset: () => void;
    replaceFrames: (frames: LogsViewerFrame[]) => void;
    flush: () => void;
}

export function useBufferedLogFrames(options: UseBufferedLogFramesOptions = {}): BufferedLogFramesResult {
    const maxFrames = options.maxFrames ?? DEFAULT_MAX_FRAMES;
    const flushIntervalMs = options.flushIntervalMs ?? DEFAULT_FLUSH_INTERVAL_MS;
    const usesExternalState = options.onFramesChange !== undefined;

    const [internalFrames, setInternalFrames] = useState<LogsViewerFrame[]>([]);
    const bufferRef = useRef<LogsViewerFrame[]>([]);
    const intervalRef = useRef<number | null>(null);
    const onFramesChangeRef = useRef(options.onFramesChange);

    onFramesChangeRef.current = options.onFramesChange;

    const applyFramesUpdate = useCallback(
        (action: SetStateAction<LogsViewerFrame[]>) => {
            if (usesExternalState) {
                onFramesChangeRef.current?.(current => capFrames(resolveSetStateAction(action, current), maxFrames));
                return;
            }

            setInternalFrames(current => capFrames(resolveSetStateAction(action, current), maxFrames));
        },
        [maxFrames, usesExternalState],
    );

    const flush = useCallback(() => {
        if (bufferRef.current.length === 0) {
            return;
        }

        const pending = bufferRef.current;
        bufferRef.current = [];
        applyFramesUpdate(current => [...current, ...pending]);
    }, [applyFramesUpdate]);

    const appendFrames = useCallback((frames: LogsViewerFrame[]) => {
        if (frames.length === 0) {
            return;
        }

        bufferRef.current.push(...frames);
    }, []);

    const reset = useCallback(() => {
        bufferRef.current = [];
        applyFramesUpdate([]);
    }, [applyFramesUpdate]);

    const replaceFrames = useCallback(
        (frames: LogsViewerFrame[]) => {
            bufferRef.current = [];
            applyFramesUpdate(capFrames(frames, maxFrames));
        },
        [applyFramesUpdate, maxFrames],
    );

    useEffect(() => {
        intervalRef.current = window.setInterval(flush, flushIntervalMs);

        return () => {
            if (intervalRef.current !== null) {
                window.clearInterval(intervalRef.current);
                intervalRef.current = null;
            }

            bufferRef.current = [];
        };
    }, [flush, flushIntervalMs]);

    const frames = usesExternalState ? (options.frames ?? []) : internalFrames;

    return {
        frames,
        appendFrames,
        reset,
        replaceFrames,
        flush,
    };
}
