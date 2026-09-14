import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import { cn } from "@/lib/utils";
import { FitAddon } from "@xterm/addon-fit";
import { SearchAddon } from "@xterm/addon-search";
import { WebLinksAddon } from "@xterm/addon-web-links";
import { WebglAddon } from "@xterm/addon-webgl";
import { Terminal } from "@xterm/xterm";
import "@xterm/xterm/css/xterm.css";

import { DEFAULT_SEARCH_MODE, LogsViewerToolbar, type SearchMode } from "./building-blocks";
import {
    DEFAULT_DOWNLOAD_FILE_NAME,
    LOG_FONT_SIZES,
    LOG_SEARCH_DECORATIONS,
    TERMINAL_SCROLLBACK,
} from "./logs-viewer.constants";
import styles from "./logs-viewer.module.scss";
import { useTerminalTheme } from "./logs-viewer.themes";
import type { LogsViewerProps, LogsViewerSearchResult } from "./logs-viewer.types";
import {
    type LogsViewerFramesAnchor,
    anchorLogsViewerFrames,
    buildDisplayedLogFrames,
    formatFramesForXterm,
    getPlainLogLines,
    isLogsViewerFramesAppend,
    isValidRegex,
} from "./logs-viewer.utils";
import { useFullViewHeight } from "./use-full-view-height";

export function LogsViewer({
    frames,
    isStreaming = false,
    isRefreshPending = false,
    height,
    isFullView = false,
    isFullHeight = false,
    fontSize: controlledFontSize,
    themeId,
    downloadFileName = DEFAULT_DOWNLOAD_FILE_NAME,
    defaultShowDebugLogs = false,
    defaultShowTimestamps = false,
    defaultTextWrapped = true,
    toolbarStart,
    toolbarFilters,
    toolbarSearch,
    status,
    className,
    onRefresh,
}: LogsViewerProps) {
    const { currentTheme } = useTerminalTheme(themeId);
    const terminalElementRef = useRef<HTMLDivElement | null>(null);
    const terminalFrameRef = useRef<HTMLDivElement | null>(null);
    const terminalRef = useRef<Terminal | null>(null);
    const fitAddonRef = useRef<FitAddon | null>(null);
    const searchAddonRef = useRef<SearchAddon | null>(null);
    // What the terminal currently holds, so an update can prove it only has to
    // append rather than rewrite. See isLogsViewerFramesAppend.
    const renderedFramesAnchorRef = useRef<LogsViewerFramesAnchor>({ length: 0, first: "", last: "" });
    const isTerminalReadyRef = useRef(false);

    const [fontSizeIndex] = useState(0);
    const [isTextWrapped, setIsTextWrapped] = useState(defaultTextWrapped);
    const [showTimestamps, setShowTimestamps] = useState(defaultShowTimestamps);
    const [showDebugLogs, setShowDebugLogs] = useState(defaultShowDebugLogs);
    const [followLogs, setFollowLogs] = useState(true);
    const [isFullscreen, setIsFullscreen] = useState(false);
    const [searchTerm, setSearchTerm] = useState("");
    const [searchMode, setSearchMode] = useState<SearchMode>(DEFAULT_SEARCH_MODE);
    // xterm throws on a malformed pattern and reports no match, which reads as
    // "nothing found" rather than "this is not a regular expression yet".
    const isSearchTermInvalid = searchMode.isRegex && !isValidRegex(searchTerm);
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
        if (!terminalRef.current) {
            return undefined;
        }

        terminalRef.current.options.fontSize = currentFontSize;
        updateDimensions();
        const timer = setTimeout(updateDimensions, 50);
        return () => {
            clearTimeout(timer);
        };
    }, [currentFontSize, updateDimensions]);

    // Update theme when currentTheme changes
    useEffect(() => {
        if (terminalRef.current) {
            terminalRef.current.options.theme = currentTheme.theme;
        }
    }, [currentTheme]);

    // Initialize xterm
    useEffect(() => {
        const element = terminalElementRef.current;
        const frame = terminalFrameRef.current;

        if (!element || !frame) {
            return;
        }

        renderedFramesAnchorRef.current = { length: 0, first: "", last: "" };

        const terminal = new Terminal({
            allowTransparency: true,
            convertEol: true,
            disableStdin: true,
            cursorBlink: false,
            cursorStyle: "bar",
            cursorInactiveStyle: "none",
            fontFamily: "Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace",
            fontSize: LOG_FONT_SIZES[0],
            scrollback: TERMINAL_SCROLLBACK,
            theme: currentTheme.theme,
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
            renderedFramesAnchorRef.current = anchorLogsViewerFrames(frames);
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

        const anchor = renderedFramesAnchorRef.current;
        const isAppend = isLogsViewerFramesAppend(frames, anchor);

        if (!isAppend) {
            // Frames were replaced, or grew somewhere other than the end - a
            // page of older stored logs arrives at the front. xterm cannot
            // prepend to its scrollback, so the whole buffer is written again.
            terminal.reset();
            const displayed = buildDisplayedLogFrames(frames, showDebugLogs);
            if (displayed.length > 0) {
                terminal.write(formatFramesForXterm(displayed, showTimestamps));
            }
        } else if (frames.length > anchor.length) {
            // Incremental append for new streamed frames
            const newRawFrames = frames.slice(anchor.length);
            const newDisplayed = buildDisplayedLogFrames(newRawFrames, showDebugLogs);
            if (newDisplayed.length > 0) {
                terminal.write(formatFramesForXterm(newDisplayed, showTimestamps));
            }
        }

        const grewAtTheFront = !isAppend && anchor.length > 0 && frames.length > anchor.length;

        renderedFramesAnchorRef.current = anchorLogsViewerFrames(frames);

        if (followLogs) {
            terminal.scrollToBottom();
        } else if (grewAtTheFront) {
            // Writing leaves the viewport at the bottom, which is the opposite
            // of what someone who just asked for older lines wants to see.
            terminal.scrollToTop();
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
        renderedFramesAnchorRef.current = anchorLogsViewerFrames(frames);

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

        if (!searchTerm.trim() || isSearchTermInvalid) {
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
                regex: searchMode.isRegex,
                caseSensitive: searchMode.isCaseSensitive,
                decorations: LOG_SEARCH_DECORATIONS,
            });
        } catch {
            try {
                searchAddon.findNext(searchTerm, {
                    incremental: true,
                    regex: searchMode.isRegex,
                    caseSensitive: searchMode.isCaseSensitive,
                });
            } catch {
                setSearchResult(null);
            }
        }
    }, [searchTerm, searchMode, isSearchTermInvalid]);

    const handleFindNext = useCallback(() => {
        if (!searchTerm.trim() || isSearchTermInvalid) {
            return;
        }
        try {
            searchAddonRef.current?.findNext(searchTerm, {
                incremental: false,
                regex: searchMode.isRegex,
                caseSensitive: searchMode.isCaseSensitive,
                decorations: LOG_SEARCH_DECORATIONS,
            });
        } catch {
            try {
                searchAddonRef.current?.findNext(searchTerm, {
                    incremental: false,
                    regex: searchMode.isRegex,
                    caseSensitive: searchMode.isCaseSensitive,
                });
            } catch {
                // Ignore search error
            }
        }
    }, [searchTerm, searchMode, isSearchTermInvalid]);

    const handleFindPrevious = useCallback(() => {
        if (!searchTerm.trim() || isSearchTermInvalid) {
            return;
        }
        try {
            searchAddonRef.current?.findPrevious(searchTerm, {
                regex: searchMode.isRegex,
                caseSensitive: searchMode.isCaseSensitive,
                decorations: LOG_SEARCH_DECORATIONS,
            });
        } catch {
            try {
                searchAddonRef.current?.findPrevious(searchTerm, {
                    regex: searchMode.isRegex,
                    caseSensitive: searchMode.isCaseSensitive,
                });
            } catch {
                // Ignore search error
            }
        }
    }, [searchTerm, searchMode, isSearchTermInvalid]);

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
        renderedFramesAnchorRef.current = anchorLogsViewerFrames(frames);

        if (followLogs) {
            terminal.scrollToBottom();
        }
    }, [isTextWrapped]); // eslint-disable-line react-hooks/exhaustive-deps

    const { containerRef: frameContainerRef, fullViewHeight } = useFullViewHeight({
        enabled: !isFullscreen,
        minHeight: 250,
    });

    // Refit when fullscreen, fullViewHeight, or isFullHeight toggles
    useEffect(() => {
        updateDimensions();
        const timer = setTimeout(updateDimensions, 50);
        return () => {
            clearTimeout(timer);
        };
    }, [fullViewHeight, isFullscreen, isFullHeight, updateDimensions]);

    const isFlexibleHeight = isFullscreen || height === "100%";
    const frameHeight = isFlexibleHeight
        ? undefined
        : isFullView && fullViewHeight !== null
          ? `${fullViewHeight}px`
          : height !== undefined
            ? typeof height === "number"
                ? `${height}px`
                : height
            : fullViewHeight !== null
              ? `${fullViewHeight}px`
              : undefined;

    return (
        <div
            className={cn(
                styles["root"],
                isFlexibleHeight && "flex-1 min-h-0",
                className,
                isFullscreen && [styles["fullscreen"], "bg-background border border-border shadow-2xl"],
            )}
        >
            {!isFullHeight && (
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
                    toolbarSearch={toolbarSearch}
                    searchMode={searchMode}
                    isSearchTermInvalid={isSearchTermInvalid}
                    onSearchTermChange={setSearchTerm}
                    onSearchModeChange={setSearchMode}
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
            )}

            {status !== undefined && (
                <div className="pb-2 text-[11px] text-muted-foreground sm:pb-2.5 sm:text-xs">{status}</div>
            )}

            <div
                ref={frameContainerRef}
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
                    style={{ background: currentTheme.background }}
                >
                    <div
                        ref={terminalElementRef}
                        className={cn(styles["terminalHost"], !isTextWrapped && styles["unwrapped"])}
                        style={{ background: currentTheme.background }}
                    />
                </div>
            </div>
        </div>
    );
}
