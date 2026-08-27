import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import { cn } from "@/lib/utils";
import { FitAddon } from "@xterm/addon-fit";
import { SearchAddon } from "@xterm/addon-search";
import { WebLinksAddon } from "@xterm/addon-web-links";
import { WebglAddon } from "@xterm/addon-webgl";
import { Terminal } from "@xterm/xterm";
import "@xterm/xterm/css/xterm.css";

import { LogsViewerToolbar } from "./building-blocks";
import {
    DEFAULT_DOWNLOAD_FILE_NAME,
    DEFAULT_LOG_VIEWER_HEIGHT,
    LOG_FONT_SIZES,
    TERMINAL_SCROLLBACK,
} from "./logs-viewer.constants";
import styles from "./logs-viewer.module.scss";
import type { LogsViewerProps, LogsViewerSearchResult } from "./logs-viewer.types";
import { buildDisplayedLogFrames, formatFramesForXterm, getPlainLogLines } from "./logs-viewer.utils";

export function LogsViewer({
    frames,
    isStreaming = false,
    isRefreshPending = false,
    height = DEFAULT_LOG_VIEWER_HEIGHT,
    fontSize: controlledFontSize,
    downloadFileName = DEFAULT_DOWNLOAD_FILE_NAME,
    defaultShowDebugLogs = false,
    defaultShowTimestamps = false,
    defaultTextWrapped = true,
    toolbarStart,
    toolbarFilters,
    className,
    onRefresh,
}: LogsViewerProps) {
    const terminalElementRef = useRef<HTMLDivElement | null>(null);
    const terminalFrameRef = useRef<HTMLDivElement | null>(null);
    const terminalRef = useRef<Terminal | null>(null);
    const fitAddonRef = useRef<FitAddon | null>(null);
    const searchAddonRef = useRef<SearchAddon | null>(null);
    const renderedFramesCountRef = useRef(0);
    const isTerminalReadyRef = useRef(false);

    const [fontSizeIndex] = useState(0);
    const [isTextWrapped, setIsTextWrapped] = useState(defaultTextWrapped);
    const [showTimestamps, setShowTimestamps] = useState(defaultShowTimestamps);
    const [showDebugLogs, setShowDebugLogs] = useState(defaultShowDebugLogs);
    const [followLogs, setFollowLogs] = useState(true);
    const [isFullscreen, setIsFullscreen] = useState(false);
    const [searchTerm, setSearchTerm] = useState("");
    const [searchResult, setSearchResult] = useState<LogsViewerSearchResult | null>(null);

    const currentFontSize = controlledFontSize ?? LOG_FONT_SIZES[fontSizeIndex] ?? 14;

    const displayedFrames = buildDisplayedLogFrames(frames, showDebugLogs);
    const displayedPlainLines = displayedFrames.flatMap(frame => getPlainLogLines(frame, showTimestamps));

    const longestLineLength = useMemo(() => {
        let max = 0;
        for (const line of displayedPlainLines) {
            if (line.length > max) {
                max = line.length;
            }
        }
        return max;
    }, [displayedPlainLines]);

    const updateDimensions = useCallback(() => {
        const terminal = terminalRef.current;
        const fitAddon = fitAddonRef.current;
        if (!terminal || !fitAddon) {
            return;
        }

        if (isTextWrapped) {
            try {
                fitAddon.fit();
            } catch {
                // Ignore fit error if container is hidden
            }
        } else {
            try {
                const dims = fitAddon.proposeDimensions();
                const containerCols = dims?.cols ?? 80;
                const containerRows = dims?.rows ?? 24;
                const targetCols = Math.min(Math.max(longestLineLength + 4, containerCols), 2500);
                terminal.resize(targetCols, containerRows);
            } catch {
                // Ignore resize error
            }
        }
    }, [isTextWrapped, longestLineLength]);

    useEffect(() => {
        if (terminalRef.current) {
            terminalRef.current.options.fontSize = currentFontSize;
            updateDimensions();
        }
    }, [currentFontSize, updateDimensions]);

    // Initialize xterm
    useEffect(() => {
        const element = terminalElementRef.current;
        const frame = terminalFrameRef.current;

        if (!element || !frame) {
            return;
        }

        renderedFramesCountRef.current = 0;

        const terminal = new Terminal({
            convertEol: true,
            disableStdin: true,
            cursorBlink: false,
            cursorStyle: "bar",
            cursorInactiveStyle: "none",
            fontFamily: "Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace",
            fontSize: LOG_FONT_SIZES[0],
            scrollback: TERMINAL_SCROLLBACK,
            theme: {
                background: "#0f172a",
                foreground: "#e5e7eb",
                selectionBackground: "#334155",
                scrollbarSliderBackground: "#334155",
                scrollbarSliderHoverBackground: "#475569",
                scrollbarSliderActiveBackground: "#64748b",
            },
        });

        const fitAddon = new FitAddon();
        const webLinksAddon = new WebLinksAddon();
        const searchAddon = new SearchAddon();

        terminal.loadAddon(fitAddon);
        terminal.loadAddon(webLinksAddon);
        terminal.loadAddon(searchAddon);

        terminal.open(element);
        terminalRef.current = terminal;
        fitAddonRef.current = fitAddon;
        searchAddonRef.current = searchAddon;

        let webglAddon: WebglAddon | null = null;
        try {
            webglAddon = new WebglAddon();
            webglAddon.onContextLoss(() => {
                webglAddon?.dispose();
                webglAddon = null;
            });
            terminal.loadAddon(webglAddon);
        } catch {
            webglAddon = null;
        }

        const searchDisposable = searchAddon.onDidChangeResults(event => {
            setSearchResult({ resultIndex: event.resultIndex, resultCount: event.resultCount });
        });

        const resizeObserver = new ResizeObserver(updateDimensions);
        resizeObserver.observe(frame);
        window.addEventListener("resize", updateDimensions);
        updateDimensions();

        isTerminalReadyRef.current = true;

        // Initial render of frames
        const initialDisplayed = buildDisplayedLogFrames(frames, showDebugLogs);
        if (initialDisplayed.length > 0) {
            const content = formatFramesForXterm(initialDisplayed, showTimestamps);
            terminal.write(content);
            renderedFramesCountRef.current = frames.length;
            if (followLogs) {
                terminal.scrollToBottom();
            }
        }

        return () => {
            isTerminalReadyRef.current = false;
            resizeObserver.disconnect();
            window.removeEventListener("resize", updateDimensions);
            searchDisposable.dispose();
            webglAddon?.dispose();
            searchAddon.dispose();
            webLinksAddon.dispose();
            terminal.dispose();
            terminalRef.current = null;
            fitAddonRef.current = null;
            searchAddonRef.current = null;
        };
    }, []); // eslint-disable-line react-hooks/exhaustive-deps

    // Handle frame updates and stream appending
    useEffect(() => {
        const terminal = terminalRef.current;
        if (!terminal || !isTerminalReadyRef.current) {
            return;
        }

        const lastCount = renderedFramesCountRef.current;

        if (frames.length < lastCount || lastCount === 0) {
            // Full re-render when frames are reset/replaced
            terminal.reset();
            const displayed = buildDisplayedLogFrames(frames, showDebugLogs);
            if (displayed.length > 0) {
                terminal.write(formatFramesForXterm(displayed, showTimestamps));
            }
            renderedFramesCountRef.current = frames.length;
        } else if (frames.length > lastCount) {
            // Incremental append for new streamed frames
            const newRawFrames = frames.slice(lastCount);
            const newDisplayed = buildDisplayedLogFrames(newRawFrames, showDebugLogs);
            if (newDisplayed.length > 0) {
                terminal.write(formatFramesForXterm(newDisplayed, showTimestamps));
            }
            renderedFramesCountRef.current = frames.length;
        }

        if (followLogs) {
            terminal.scrollToBottom();
        }
    }, [frames, showDebugLogs, showTimestamps, followLogs]);

    // Handle full re-render when showDebugLogs or showTimestamps toggles
    const isFirstMountRef = useRef(true);
    useEffect(() => {
        if (isFirstMountRef.current) {
            isFirstMountRef.current = false;
            return;
        }

        const terminal = terminalRef.current;
        if (!terminal || !isTerminalReadyRef.current) {
            return;
        }

        terminal.reset();
        const displayed = buildDisplayedLogFrames(frames, showDebugLogs);
        if (displayed.length > 0) {
            terminal.write(formatFramesForXterm(displayed, showTimestamps));
        }
        renderedFramesCountRef.current = frames.length;

        if (followLogs) {
            terminal.scrollToBottom();
        }
    }, [showDebugLogs, showTimestamps]); // eslint-disable-line react-hooks/exhaustive-deps

    // Search effect
    useEffect(() => {
        const searchAddon = searchAddonRef.current;
        if (!searchAddon || !isTerminalReadyRef.current) {
            return;
        }

        if (!searchTerm.trim()) {
            try {
                searchAddon.clearDecorations();
            } catch {
                // Ignore clear error
            }
            setSearchResult(null);
            return;
        }

        try {
            searchAddon.findNext(searchTerm, {
                incremental: true,
                regex: false,
                caseSensitive: false,
                decorations: {
                    matchBackground: "#4338ca",
                    activeMatchBackground: "#eab308",
                    matchBorder: "#6366f1",
                    activeMatchBorder: "#fde047",
                    matchOverviewRuler: "transparent",
                    activeMatchColorOverviewRuler: "transparent",
                },
            });
        } catch {
            try {
                searchAddon.findNext(searchTerm, {
                    incremental: true,
                    regex: false,
                    caseSensitive: false,
                });
            } catch {
                setSearchResult(null);
            }
        }
    }, [searchTerm]);

    const handleFindNext = useCallback(() => {
        if (!searchTerm.trim()) {
            return;
        }
        try {
            searchAddonRef.current?.findNext(searchTerm, {
                incremental: false,
                regex: false,
                caseSensitive: false,
                decorations: {
                    matchBackground: "#4338ca",
                    activeMatchBackground: "#eab308",
                    matchBorder: "#6366f1",
                    activeMatchBorder: "#fde047",
                    matchOverviewRuler: "transparent",
                    activeMatchColorOverviewRuler: "transparent",
                },
            });
        } catch {
            try {
                searchAddonRef.current?.findNext(searchTerm, {
                    incremental: false,
                    regex: false,
                    caseSensitive: false,
                });
            } catch {
                // Ignore search error
            }
        }
    }, [searchTerm]);

    const handleFindPrevious = useCallback(() => {
        if (!searchTerm.trim()) {
            return;
        }
        try {
            searchAddonRef.current?.findPrevious(searchTerm, {
                regex: false,
                caseSensitive: false,
                decorations: {
                    matchBackground: "#4338ca",
                    activeMatchBackground: "#eab308",
                    matchBorder: "#6366f1",
                    activeMatchBorder: "#fde047",
                    matchOverviewRuler: "transparent",
                    activeMatchColorOverviewRuler: "transparent",
                },
            });
        } catch {
            try {
                searchAddonRef.current?.findPrevious(searchTerm, {
                    regex: false,
                    caseSensitive: false,
                });
            } catch {
                // Ignore search error
            }
        }
    }, [searchTerm]);

    // Handle ESC to exit fullscreen
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

    // Handle isTextWrapped toggle
    useEffect(() => {
        const terminal = terminalRef.current;
        if (!terminal || !isTerminalReadyRef.current) {
            return;
        }

        updateDimensions();
        terminal.reset();
        const displayed = buildDisplayedLogFrames(frames, showDebugLogs);
        if (displayed.length > 0) {
            terminal.write(formatFramesForXterm(displayed, showTimestamps));
        }
        renderedFramesCountRef.current = frames.length;

        if (followLogs) {
            terminal.scrollToBottom();
        }
    }, [isTextWrapped]); // eslint-disable-line react-hooks/exhaustive-deps

    // Refit when fullscreen toggles
    useEffect(() => {
        updateDimensions();
        const timer = setTimeout(updateDimensions, 50);
        return () => {
            clearTimeout(timer);
        };
    }, [isFullscreen, updateDimensions]);

    const frameHeight = typeof height === "number" ? `${height}px` : height;
    const isFlexibleHeight = isFullscreen || height === "100%";

    return (
        <div
            className={cn(
                styles["root"],
                isFlexibleHeight && "flex-1 min-h-0",
                className,
                isFullscreen && [styles["fullscreen"], "bg-background border border-border shadow-2xl"],
            )}
        >
            <LogsViewerToolbar
                isStreaming={isStreaming}
                isRefreshPending={isRefreshPending}
                displayedPlainLines={displayedPlainLines}
                downloadFileName={downloadFileName}
                isTextWrapped={isTextWrapped}
                showTimestamps={showTimestamps}
                showDebugLogs={showDebugLogs}
                followLogs={followLogs}
                searchTerm={searchTerm}
                searchResult={searchResult}
                toolbarStart={toolbarStart}
                toolbarFilters={toolbarFilters}
                onSearchTermChange={setSearchTerm}
                onFindNext={handleFindNext}
                onFindPrevious={handleFindPrevious}
                onToggleTextWrap={() => {
                    setIsTextWrapped(current => !current);
                }}
                onToggleTimestamps={() => {
                    setShowTimestamps(current => !current);
                }}
                onToggleDebugLogs={() => {
                    setShowDebugLogs(current => !current);
                }}
                onToggleFollowLogs={() => {
                    const next = !followLogs;
                    setFollowLogs(next);
                    if (next) {
                        terminalRef.current?.scrollToBottom();
                    }
                }}
                onRefresh={onRefresh}
            />

            <div
                className={cn("flex min-h-0", isFlexibleHeight ? "flex-1 w-full" : "w-full")}
                style={isFlexibleHeight ? undefined : { height: frameHeight }}
            >
                <div
                    ref={terminalFrameRef}
                    className={cn(
                        styles["terminalFrame"],
                        !isTextWrapped && styles["unwrapped"],
                        "border border-border/60 flex-1 min-h-0 h-full w-full",
                    )}
                >
                    <div
                        ref={terminalElementRef}
                        className={cn(styles["terminalHost"], !isTextWrapped && styles["unwrapped"])}
                    />
                </div>
            </div>
        </div>
    );
}
