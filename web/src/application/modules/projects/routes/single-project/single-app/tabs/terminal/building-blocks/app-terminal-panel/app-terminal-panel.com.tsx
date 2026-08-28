import { useCallback, useEffect, useRef, useState } from "react";

import { Button, Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui";
import { Tooltip, TooltipContent, TooltipTrigger } from "@components/ui/tooltip";
import type { WebSocketReadyState, WebSocketSubscription } from "@infrastructure/websocket";
import { cn } from "@lib/utils";
import { ClipboardAddon } from "@xterm/addon-clipboard";
import { FitAddon } from "@xterm/addon-fit";
import { SearchAddon } from "@xterm/addon-search";
import { WebLinksAddon } from "@xterm/addon-web-links";
import { WebglAddon } from "@xterm/addon-webgl";
import { Terminal } from "@xterm/xterm";
import "@xterm/xterm/css/xterm.css";
import {
    ChevronDown,
    ChevronUp,
    Download,
    Maximize2,
    Minimize2,
    Search,
    TextCursorInputIcon,
    Upload,
    X,
} from "lucide-react";
import { useAppTerminalWsApi } from "~/projects/api";
import { type AppTerminalInitMessage, buildAppTerminalResizeMessage } from "~/projects/api/services";
import { useExportContainerFilesDialog } from "~/projects/dialogs/export-container-files";
import { useImportFilesToContainerDialog } from "~/projects/dialogs/import-files-to-container";

import { FullViewIcon, TextZoomIcon } from "@assets/icons";

import { LOG_SEARCH_DECORATIONS, useFullViewHeight } from "@application/shared/components/logs-viewer";

import { AppTerminalCommandTemplatePanel } from "./app-terminal-command-template-panel.com";
import styles from "./app-terminal-panel.module.scss";

const TERMINAL_HEIGHT = "clamp(350px, calc(100vh - 400px), 2000px)";
const TERMINAL_FONT_SIZES = [14, 16, 18, 20] as const;
const TERMINAL_SCROLLBACK = 10_000;

export function AppTerminalPanel({
    projectID,
    env,
    appID,
    supportedShells,
    selectedShell,
    isFullView: controlledFullView,
    onToggleFullView,
    onSelectedShellChange,
}: AppTerminalPanelProps) {
    const terminalElementRef = useRef<HTMLDivElement | null>(null);
    const terminalFrameRef = useRef<HTMLDivElement | null>(null);
    const terminalRef = useRef<Terminal | null>(null);
    const fitAddonRef = useRef<FitAddon | null>(null);
    const searchAddonRef = useRef<SearchAddon | null>(null);
    const [searchTerm, setSearchTerm] = useState("");
    const [searchResult, setSearchResult] = useState<{ resultIndex: number; resultCount: number } | null>(null);
    const subscriptionRef = useRef<WebSocketSubscription | null>(null);
    const abortControllerRef = useRef<AbortController | null>(null);
    const resizeFrameRef = useRef<number | null>(null);
    const lastObservedFrameSizeRef = useRef({ clientWidth: 0, clientHeight: 0 });
    const lastSentTerminalSizeRef = useRef({ cols: 0, rows: 0 });
    const inputEncoderRef = useRef(new TextEncoder());
    const [fontSizeIndex, setFontSizeIndex] = useState(0);
    const [internalFullView, setInternalFullView] = useState(false);
    const isFullView = controlledFullView ?? internalFullView;
    const handleToggleFullView =
        onToggleFullView ??
        (() => {
            setInternalFullView(current => !current);
        });
    const [webSocketReadyState, setWebSocketReadyState] = useState<WebSocketReadyState>(WebSocket.CLOSED);
    const [isFullscreen, setIsFullscreen] = useState(false);
    const { containerRef: terminalContainerRef, fullViewHeight } = useFullViewHeight({
        enabled: !isFullscreen,
        minHeight: 250,
    });
    const [isCommandTemplatePanelOpen, setIsCommandTemplatePanelOpen] = useState(false);
    const [connectedInfo, setConnectedInfo] = useState<AppTerminalInitMessage | null>(null);
    const { streams } = useAppTerminalWsApi();
    const exportContainerFilesDialog = useExportContainerFilesDialog();
    const importFilesToContainerDialog = useImportFilesToContainerDialog();

    const currentFontSize = TERMINAL_FONT_SIZES[fontSizeIndex] ?? 14;

    const cycleFontSize = useCallback(() => {
        setFontSizeIndex(current => (current + 1) % TERMINAL_FONT_SIZES.length);
    }, []);

    const isConnectionActive = webSocketReadyState === WebSocket.CONNECTING || webSocketReadyState === WebSocket.OPEN;
    const canConnect = selectedShell !== "" && !isConnectionActive;
    const status = getTerminalStatus(webSocketReadyState);
    const nodeId = connectedInfo?.nodeId ?? "";
    const containerId = connectedInfo?.containerId ?? "";

    function openImportDialog() {
        importFilesToContainerDialog.actions.open(projectID, env, appID, nodeId, containerId);
    }

    function openExportDialog() {
        exportContainerFilesDialog.actions.open(projectID, env, appID, nodeId, containerId);
    }

    function insertCommand(command: string) {
        terminalRef.current?.paste(command);
        terminalRef.current?.focus();
    }

    const sendResizeToSocket = useCallback((socket: WebSocket | undefined, width: number, height: number) => {
        if (!socket || socket.readyState !== WebSocket.OPEN || width <= 0 || height <= 0) {
            return;
        }

        if (lastSentTerminalSizeRef.current.cols === width && lastSentTerminalSizeRef.current.rows === height) {
            return;
        }

        lastSentTerminalSizeRef.current = { cols: width, rows: height };
        socket.send(JSON.stringify(buildAppTerminalResizeMessage(width, height)));
    }, []);

    const sendCurrentResize = useCallback((socket?: WebSocket) => {
        const targetSocket = socket ?? subscriptionRef.current?.socket;
        const terminal = terminalRef.current;

        if (!targetSocket || targetSocket.readyState !== WebSocket.OPEN || !terminal) {
            return;
        }

        const { cols, rows } = terminal;

        if (cols <= 0 || rows <= 0) {
            return;
        }

        const lastSent = lastSentTerminalSizeRef.current;
        if (lastSent.cols === cols && lastSent.rows === rows) {
            return;
        }

        lastSentTerminalSizeRef.current = { cols, rows };
        targetSocket.send(JSON.stringify(buildAppTerminalResizeMessage(cols, rows)));
    }, []);

    const fitTerminal = useCallback((force = false) => {
        const frame = terminalFrameRef.current;
        const fitAddon = fitAddonRef.current;
        const terminal = terminalRef.current;

        if (!frame || !fitAddon || !terminal) {
            return;
        }

        const { clientWidth, clientHeight } = frame;

        if (clientWidth <= 0 || clientHeight <= 0) {
            return;
        }

        const lastObserved = lastObservedFrameSizeRef.current;
        if (!force && lastObserved.clientWidth === clientWidth && lastObserved.clientHeight === clientHeight) {
            return;
        }

        lastObservedFrameSizeRef.current = { clientWidth, clientHeight };

        try {
            fitAddon.fit();
        } catch {
            // Ignore fit error if dimensions are not yet measurable
        }
    }, []);

    const fitAndSendResize = useCallback(
        (force = false) => {
            const terminal = terminalRef.current;
            const previousCols = terminal?.cols ?? 0;
            const previousRows = terminal?.rows ?? 0;

            fitTerminal(force);

            const nextCols = terminal?.cols ?? 0;
            const nextRows = terminal?.rows ?? 0;

            if (previousCols === nextCols && previousRows === nextRows && !force) {
                return;
            }

            sendCurrentResize();
        },
        [fitTerminal, sendCurrentResize],
    );

    const scheduleFitAndResize = useCallback(
        (force = false) => {
            if (resizeFrameRef.current !== null) {
                window.cancelAnimationFrame(resizeFrameRef.current);
            }

            resizeFrameRef.current = window.requestAnimationFrame(() => {
                resizeFrameRef.current = null;
                fitAndSendResize(force);
            });
        },
        [fitAndSendResize],
    );

    useEffect(() => {
        if (!terminalRef.current) {
            return undefined;
        }

        terminalRef.current.options.fontSize = currentFontSize;
        lastObservedFrameSizeRef.current = { clientWidth: 0, clientHeight: 0 };
        scheduleFitAndResize(true);
        const timer = setTimeout(() => {
            fitAndSendResize(true);
        }, 50);
        return () => {
            clearTimeout(timer);
        };
    }, [currentFontSize, fitAndSendResize, scheduleFitAndResize]);

    const closeConnection = useCallback(() => {
        abortControllerRef.current?.abort();
        abortControllerRef.current = null;

        const subscription = subscriptionRef.current;
        subscriptionRef.current = null;
        subscription?.close();
        setWebSocketReadyState(WebSocket.CLOSED);
        setConnectedInfo(null);
    }, []);

    const handleConnect = useCallback(() => {
        if (!canConnect) {
            return;
        }

        closeConnection();
        terminalRef.current?.clear();
        terminalRef.current?.focus();
        fitTerminal();

        const terminal = terminalRef.current;
        const width = terminal && terminal.cols > 0 ? terminal.cols : undefined;
        const height = terminal && terminal.rows > 0 ? terminal.rows : undefined;
        const abortController = new AbortController();
        abortControllerRef.current = abortController;
        setWebSocketReadyState(WebSocket.CONNECTING);

        void streams
            .open(
                {
                    projectID,
                    env,
                    appID,
                    shell: selectedShell,
                    width,
                    height,
                },
                {
                    onOpen: (_event, socket) => {
                        setWebSocketReadyState(socket.readyState);
                        terminalRef.current?.focus();
                        sendCurrentResize(socket);
                    },
                    onMessage: (message, event) => {
                        if (typeof event.data === "string") {
                            try {
                                const parsed: unknown = JSON.parse(event.data);
                                if (isAppTerminalInitMessage(parsed)) {
                                    setConnectedInfo(parsed);
                                    return;
                                }
                            } catch {
                                // Ignore non-JSON text message and write to terminal
                            }
                        }

                        terminalRef.current?.write(message);
                    },
                    onMessageError: error => {
                        console.error("Failed to read app terminal output", error);
                    },
                    onError: () => {
                        setWebSocketReadyState(WebSocket.CLOSING);
                    },
                    onClose: (_event, socket) => {
                        if (subscriptionRef.current?.socket === socket) {
                            subscriptionRef.current = null;
                        }

                        if (abortControllerRef.current === abortController) {
                            abortControllerRef.current = null;
                        }

                        setWebSocketReadyState(WebSocket.CLOSED);
                        setConnectedInfo(null);
                    },
                    onReadyStateChange: readyState => {
                        setWebSocketReadyState(readyState);
                    },
                },
                abortController.signal,
            )

            .then(subscription => {
                if (abortController.signal.aborted) {
                    subscription.close();
                    return;
                }

                subscriptionRef.current = subscription;
                setWebSocketReadyState(subscription.getReadyState());
                sendCurrentResize(subscription.socket);
            })
            .catch((error: unknown) => {
                if (!abortController.signal.aborted) {
                    console.error("Failed to connect app terminal", error);
                    setWebSocketReadyState(WebSocket.CLOSED);
                }
            });
    }, [appID, canConnect, closeConnection, env, fitTerminal, projectID, selectedShell, sendCurrentResize, streams]);

    useEffect(() => {
        const frameElement = terminalFrameRef.current;
        const terminalElement = terminalElementRef.current;

        if (!frameElement || !terminalElement) {
            return;
        }

        const terminal = new Terminal({
            cursorBlink: true,
            cursorStyle: "bar",
            fontSize: currentFontSize,
            fontFamily: "ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace",
            scrollback: TERMINAL_SCROLLBACK,
            theme: {
                background: "#0f172a",
                foreground: "#e2e8f0",
                cursor: "#38bdf8",
                selectionBackground: "#f59e0b",
                selectionForeground: "#000000",
                selectionInactiveBackground: "#d97706",
            },
        });

        const fitAddon = new FitAddon();
        const webLinksAddon = new WebLinksAddon();
        const clipboardAddon = new ClipboardAddon();
        const searchAddon = new SearchAddon();

        terminal.loadAddon(fitAddon);
        terminal.loadAddon(webLinksAddon);
        terminal.loadAddon(clipboardAddon);
        terminal.loadAddon(searchAddon);

        let webglAddon: WebglAddon | null = null;

        try {
            webglAddon = new WebglAddon();
            terminal.loadAddon(webglAddon);
        } catch {
            webglAddon = null;
        }

        terminal.open(terminalElement);
        terminalRef.current = terminal;
        fitAddonRef.current = fitAddon;
        searchAddonRef.current = searchAddon;

        const searchDisposable = searchAddon.onDidChangeResults(event => {
            setSearchResult({ resultIndex: event.resultIndex, resultCount: event.resultCount });
        });

        const dataDisposable = terminal.onData(data => {
            const socket = subscriptionRef.current?.socket;

            if (socket?.readyState === WebSocket.OPEN) {
                socket.send(inputEncoderRef.current.encode(data));
            }
        });

        const resizeDisposable = terminal.onResize(({ cols, rows }) => {
            const socket = subscriptionRef.current?.socket;

            if (socket) {
                sendResizeToSocket(socket, cols, rows);
            }
        });

        function handleWindowResize() {
            scheduleFitAndResize();
        }

        const resizeObserver = new ResizeObserver(() => {
            scheduleFitAndResize();
        });

        resizeObserver.observe(frameElement);
        window.addEventListener("resize", handleWindowResize);

        scheduleFitAndResize();

        return () => {
            closeConnection();
            searchDisposable.dispose();
            searchAddon.dispose();
            searchAddonRef.current = null;
            resizeObserver.disconnect();
            window.removeEventListener("resize", handleWindowResize);
            dataDisposable.dispose();
            resizeDisposable.dispose();
            webglAddon?.dispose();
            webLinksAddon.dispose();
            clipboardAddon.dispose();
            terminal.dispose();
            terminalRef.current = null;
            fitAddonRef.current = null;

            if (resizeFrameRef.current !== null) {
                window.cancelAnimationFrame(resizeFrameRef.current);
                resizeFrameRef.current = null;
            }
        };
    }, [closeConnection, scheduleFitAndResize, sendResizeToSocket]); // eslint-disable-line react-hooks/exhaustive-deps

    const handleFindNext = useCallback(() => {
        if (!searchTerm.trim()) {
            return;
        }
        try {
            searchAddonRef.current?.findNext(searchTerm, {
                incremental: false,
                regex: false,
                caseSensitive: false,
                decorations: LOG_SEARCH_DECORATIONS,
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
                decorations: LOG_SEARCH_DECORATIONS,
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

    useEffect(() => {
        const searchAddon = searchAddonRef.current;
        if (!searchAddon) {
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
                decorations: LOG_SEARCH_DECORATIONS,
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

    useEffect(() => {
        lastObservedFrameSizeRef.current = { clientWidth: 0, clientHeight: 0 };
        scheduleFitAndResize(true);
        const timer = setTimeout(() => {
            fitAndSendResize(true);
        }, 50);
        return () => {
            clearTimeout(timer);
        };
    }, [fullViewHeight, isCommandTemplatePanelOpen, isFullscreen, isFullView, fitAndSendResize, scheduleFitAndResize]);

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

    return (
        <div
            className={cn(
                "flex min-w-0 flex-col gap-3",
                isFullscreen && "fixed inset-4 z-50 min-h-0 rounded-lg border bg-background p-4 shadow-2xl",
            )}
        >
            <div className="flex flex-wrap items-center justify-between gap-2 sm:gap-3">
                <div className="flex flex-wrap items-center gap-2 sm:gap-3">
                    <div className="flex items-center gap-2">
                        <span className="text-sm sm:text-md text-foreground">Shell</span>
                        <Select
                            value={selectedShell}
                            disabled={isConnectionActive || supportedShells.length === 0}
                            onValueChange={onSelectedShellChange}
                        >
                            <SelectTrigger className="w-24 sm:w-28 h-8 sm:h-9 text-xs sm:text-sm">
                                <SelectValue placeholder="Shell" />
                            </SelectTrigger>
                            <SelectContent>
                                {supportedShells.map(shell => (
                                    <SelectItem
                                        key={shell}
                                        value={shell}
                                    >
                                        {shell}
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                    </div>
                    <Button
                        type="button"
                        variant="link"
                        className="h-auto px-0 text-sm sm:text-base"
                        disabled={!canConnect}
                        onClick={handleConnect}
                    >
                        Connect
                    </Button>
                </div>

                <div className="flex items-center gap-1">
                    <Tooltip>
                        <TooltipTrigger asChild>
                            <Button
                                type="button"
                                variant="ghost"
                                size="icon-sm"
                                aria-label={`Text size: ${currentFontSize}px`}
                                onClick={cycleFontSize}
                            >
                                <TextZoomIcon className="size-4" />
                            </Button>
                        </TooltipTrigger>
                        <TooltipContent>{`Text size: ${currentFontSize}px`}</TooltipContent>
                    </Tooltip>

                    <Tooltip>
                        <TooltipTrigger asChild>
                            <Button
                                type="button"
                                variant="ghost"
                                size="icon-sm"
                                aria-label={isFullView ? "Exit full view" : "Full view"}
                                className={cn(isFullView && "text-primary bg-accent")}
                                onClick={handleToggleFullView}
                            >
                                <FullViewIcon className="size-4" />
                            </Button>
                        </TooltipTrigger>
                        <TooltipContent>{isFullView ? "Exit full view" : "Full view"}</TooltipContent>
                    </Tooltip>

                    <Tooltip>
                        <TooltipTrigger asChild>
                            <Button
                                type="button"
                                variant="ghost"
                                size="icon-sm"
                                aria-label={isFullscreen ? "Exit fullscreen terminal" : "Fullscreen terminal"}
                                onClick={() => {
                                    setIsFullscreen(current => !current);
                                }}
                            >
                                {isFullscreen ? <Minimize2 className="size-4" /> : <Maximize2 className="size-4" />}
                            </Button>
                        </TooltipTrigger>
                        <TooltipContent>{isFullscreen ? "Exit fullscreen" : "Fullscreen terminal"}</TooltipContent>
                    </Tooltip>
                </div>
            </div>

            <div className="flex flex-wrap items-center justify-between gap-2 sm:gap-3">
                <div className="flex items-center gap-2 sm:gap-3">
                    <span
                        aria-label={status.label}
                        className={cn("size-3.5 sm:size-4 rounded-full border-2", status.indicatorClassName)}
                    />
                    <span className={cn("text-xs sm:text-sm", status.textClassName)}>{status.label}</span>
                    {isConnectionActive && (
                        <Button
                            type="button"
                            variant="link"
                            className="h-auto px-0 py-0 text-xs sm:text-sm"
                            onClick={closeConnection}
                        >
                            Disconnect
                        </Button>
                    )}
                </div>

                <div className="flex flex-wrap items-center gap-2 sm:gap-2.5 min-w-0">
                    <div className="relative flex items-center min-w-0 w-full sm:w-56 max-w-full">
                        <Search className="absolute left-2.5 size-3.5 text-muted-foreground pointer-events-none" />
                        <input
                            type="text"
                            placeholder="Find in terminal..."
                            value={searchTerm}
                            onChange={e => {
                                setSearchTerm(e.target.value);
                            }}
                            onKeyDown={e => {
                                if (e.key === "Enter") {
                                    if (e.shiftKey) {
                                        handleFindPrevious();
                                    } else {
                                        handleFindNext();
                                    }
                                }
                            }}
                            className="w-full h-8 sm:h-9 pl-8 pr-16 text-xs sm:text-sm rounded-md border border-input bg-background/50 focus:bg-background px-3 py-1 text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring"
                        />
                        {searchTerm && (
                            <div className="absolute right-1.5 flex items-center gap-0.5 text-muted-foreground">
                                {searchResult && (
                                    <span className="text-[10px] sm:text-xs font-mono mr-1 text-muted-foreground">
                                        {searchResult.resultCount > 0
                                            ? `${searchResult.resultIndex + 1}/${searchResult.resultCount}`
                                            : "0/0"}
                                    </span>
                                )}
                                <button
                                    type="button"
                                    aria-label="Previous match"
                                    onClick={handleFindPrevious}
                                    className="p-1 hover:text-foreground rounded hover:bg-muted"
                                >
                                    <ChevronUp className="size-3 sm:size-3.5" />
                                </button>
                                <button
                                    type="button"
                                    aria-label="Next match"
                                    onClick={handleFindNext}
                                    className="p-1 hover:text-foreground rounded hover:bg-muted"
                                >
                                    <ChevronDown className="size-3 sm:size-3.5" />
                                </button>
                                <button
                                    type="button"
                                    aria-label="Clear search"
                                    onClick={() => {
                                        setSearchTerm("");
                                    }}
                                    className="p-1 hover:text-foreground rounded hover:bg-muted"
                                >
                                    <X className="size-3 sm:size-3.5" />
                                </button>
                            </div>
                        )}
                    </div>

                    <div className="flex flex-wrap items-center gap-1 sm:gap-1.5">
                        <Button
                            type="button"
                            variant="link"
                            className="h-auto py-0 text-xs sm:text-sm gap-1"
                            aria-pressed={isCommandTemplatePanelOpen}
                            onClick={() => {
                                setIsCommandTemplatePanelOpen(current => !current);
                            }}
                        >
                            <TextCursorInputIcon className="size-3.5 sm:size-4" />
                            Insert Command
                        </Button>
                        <Button
                            type="button"
                            variant="link"
                            disabled={!isConnectionActive}
                            className="h-auto py-0 text-xs sm:text-sm gap-1"
                            onClick={openImportDialog}
                        >
                            <Upload className="size-3.5 sm:size-4" />
                            Upload
                        </Button>
                        <Button
                            type="button"
                            variant="link"
                            disabled={!isConnectionActive}
                            className="h-auto py-0 text-xs sm:text-sm gap-1"
                            onClick={openExportDialog}
                        >
                            <Download className="size-3.5 sm:size-4" />
                            Download
                        </Button>
                    </div>
                </div>
            </div>

            <div
                ref={terminalContainerRef}
                className={cn("flex min-h-0 gap-4", isFullscreen ? "flex-1" : "w-full")}
                style={
                    isFullscreen
                        ? undefined
                        : fullViewHeight !== null
                          ? { height: `${fullViewHeight}px` }
                          : { height: TERMINAL_HEIGHT }
                }
            >
                <div
                    ref={terminalFrameRef}
                    className={cn(styles["terminalFrame"], "min-h-0 flex-1 overflow-hidden border border-slate-800")}
                >
                    <div
                        ref={terminalElementRef}
                        className={styles["terminalHost"]}
                    />
                </div>
                {isCommandTemplatePanelOpen && (
                    <AppTerminalCommandTemplatePanel
                        projectID={projectID}
                        env={env}
                        appID={appID}
                        isConnected={webSocketReadyState === WebSocket.OPEN}
                        onInsertCommand={insertCommand}
                    />
                )}
            </div>
        </div>
    );
}

function getTerminalStatus(readyState: WebSocketReadyState): TerminalStatus {
    if (readyState === WebSocket.OPEN) {
        return {
            label: "connected",
            indicatorClassName: "border-emerald-500 bg-emerald-500/20",
            textClassName: "text-emerald-600",
        };
    }

    if (readyState === WebSocket.CONNECTING) {
        return {
            label: "connecting",
            indicatorClassName: "border-blue-500 bg-blue-500/20",
            textClassName: "text-blue-600",
        };
    }

    return {
        label: "disconnected",
        indicatorClassName: "border-rose-500 bg-transparent",
        textClassName: "text-rose-500",
    };
}

interface TerminalStatus {
    label: string;
    indicatorClassName: string;
    textClassName: string;
}

interface AppTerminalPanelProps {
    projectID: string;
    env: string;
    appID: string;
    supportedShells: string[];
    selectedShell: string;
    isFullView?: boolean;
    onToggleFullView?: () => void;
    onSelectedShellChange: (shell: string) => void;
}

function isAppTerminalInitMessage(data: unknown): data is AppTerminalInitMessage {
    if (typeof data !== "object" || data === null) {
        return false;
    }

    const candidate = data as Record<string, unknown>;
    return (
        candidate["type"] === "init" &&
        typeof candidate["containerId"] === "string" &&
        typeof candidate["nodeId"] === "string"
    );
}
