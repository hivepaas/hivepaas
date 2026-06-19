import { useCallback, useEffect, useRef, useState } from "react";

import { Button, Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui";
import { Tooltip, TooltipContent, TooltipTrigger } from "@components/ui/tooltip";
import type { WebSocketReadyState, WebSocketSubscription } from "@infrastructure/websocket";
import { cn } from "@lib/utils";
import { FitAddon } from "@xterm/addon-fit";
import { Terminal } from "@xterm/xterm";
import "@xterm/xterm/css/xterm.css";
import { Maximize2, Minimize2 } from "lucide-react";
import { useAppTerminalWsApi } from "~/projects/api";
import { buildAppTerminalResizeMessage } from "~/projects/api/services";

import styles from "./app-terminal-panel.module.scss";

const TERMINAL_HEIGHT = "clamp(700px, calc(100vh - 330px), 2000px)";
const TERMINAL_FONT_SIZE = 14;
const TERMINAL_SCROLLBACK = 10_000;

export function AppTerminalPanel({
    projectID,
    appID,
    supportedShells,
    selectedShell,
    onSelectedShellChange,
}: AppTerminalPanelProps) {
    const terminalElementRef = useRef<HTMLDivElement | null>(null);
    const terminalRef = useRef<Terminal | null>(null);
    const fitAddonRef = useRef<FitAddon | null>(null);
    const subscriptionRef = useRef<WebSocketSubscription | null>(null);
    const abortControllerRef = useRef<AbortController | null>(null);
    const resizeFrameRef = useRef<number | null>(null);
    const inputEncoderRef = useRef(new TextEncoder());
    const [webSocketReadyState, setWebSocketReadyState] = useState<WebSocketReadyState>(WebSocket.CLOSED);
    const [isFullscreen, setIsFullscreen] = useState(false);
    const { streams } = useAppTerminalWsApi();

    const isConnectionActive = webSocketReadyState === WebSocket.CONNECTING || webSocketReadyState === WebSocket.OPEN;
    const canConnect = selectedShell !== "" && !isConnectionActive;
    const status = getTerminalStatus(webSocketReadyState);

    const sendResizeToSocket = useCallback((socket: WebSocket | undefined, width: number, height: number) => {
        if (!socket || socket.readyState !== WebSocket.OPEN || width <= 0 || height <= 0) {
            return;
        }

        socket.send(JSON.stringify(buildAppTerminalResizeMessage(width, height)));
    }, []);

    const sendCurrentResize = useCallback(
        (socket = subscriptionRef.current?.socket) => {
            const terminal = terminalRef.current;

            if (!terminal) {
                return;
            }

            sendResizeToSocket(socket, terminal.cols, terminal.rows);
        },
        [sendResizeToSocket],
    );

    const fitTerminal = useCallback(() => {
        try {
            fitAddonRef.current?.fit();
        } catch (error) {
            console.error("Failed to fit app terminal", error);
        }
    }, []);

    const fitAndSendResize = useCallback(() => {
        fitTerminal();
        sendCurrentResize();
    }, [fitTerminal, sendCurrentResize]);

    const scheduleFitAndResize = useCallback(() => {
        if (resizeFrameRef.current !== null) {
            window.cancelAnimationFrame(resizeFrameRef.current);
        }

        resizeFrameRef.current = window.requestAnimationFrame(() => {
            resizeFrameRef.current = null;
            fitAndSendResize();
        });
    }, [fitAndSendResize]);

    const closeConnection = useCallback(() => {
        abortControllerRef.current?.abort();
        abortControllerRef.current = null;

        const subscription = subscriptionRef.current;
        subscriptionRef.current = null;
        subscription?.close();
        setWebSocketReadyState(WebSocket.CLOSED);
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
                    onMessage: message => {
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
    }, [appID, canConnect, closeConnection, fitTerminal, projectID, selectedShell, sendCurrentResize, streams]);

    useEffect(() => {
        const element = terminalElementRef.current;

        if (!element) {
            return;
        }

        const terminal = new Terminal({
            convertEol: true,
            cursorBlink: true,
            fontFamily: "Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace",
            fontSize: TERMINAL_FONT_SIZE,
            scrollback: TERMINAL_SCROLLBACK,
            theme: {
                background: "#0f172a",
                foreground: "#e5e7eb",
                cursor: "#f8fafc",
                selectionBackground: "#334155",
            },
        });
        const fitAddon = new FitAddon();
        terminal.loadAddon(fitAddon);
        terminal.open(element);
        terminalRef.current = terminal;
        fitAddonRef.current = fitAddon;

        const dataDisposable = terminal.onData(data => {
            const socket = subscriptionRef.current?.socket;

            if (socket?.readyState !== WebSocket.OPEN) {
                return;
            }

            socket.send(inputEncoderRef.current.encode(data));
        });
        const resizeDisposable = terminal.onResize(size => {
            sendResizeToSocket(subscriptionRef.current?.socket, size.cols, size.rows);
        });
        const resizeObserver = new ResizeObserver(scheduleFitAndResize);

        resizeObserver.observe(element);
        window.addEventListener("resize", scheduleFitAndResize);
        scheduleFitAndResize();

        return () => {
            closeConnection();
            resizeObserver.disconnect();
            window.removeEventListener("resize", scheduleFitAndResize);
            dataDisposable.dispose();
            resizeDisposable.dispose();
            terminal.dispose();
            terminalRef.current = null;
            fitAddonRef.current = null;

            if (resizeFrameRef.current !== null) {
                window.cancelAnimationFrame(resizeFrameRef.current);
                resizeFrameRef.current = null;
            }
        };
    }, [closeConnection, scheduleFitAndResize, sendResizeToSocket]);

    useEffect(() => {
        scheduleFitAndResize();
    }, [isFullscreen, scheduleFitAndResize]);

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
            <div className="flex flex-wrap items-center justify-between gap-3">
                <div className="flex flex-wrap items-center gap-3">
                    <div className="flex items-center gap-2">
                        <span className="text-md text-foreground">Shell</span>
                        <Select
                            value={selectedShell}
                            disabled={isConnectionActive || supportedShells.length === 0}
                            onValueChange={onSelectedShellChange}
                        >
                            <SelectTrigger className="w-28">
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
                        className="h-auto px-0 text-base"
                        disabled={!canConnect}
                        onClick={handleConnect}
                    >
                        Connect
                    </Button>
                </div>

                <Tooltip>
                    <TooltipTrigger asChild>
                        <Button
                            type="button"
                            variant="ghost"
                            size="icon"
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

            <div className="flex items-center gap-3">
                <span
                    aria-label={status.label}
                    className={cn("size-4 rounded-full border-2", status.indicatorClassName)}
                />
                <span className={cn("text-sm", status.textClassName)}>{status.label}</span>
                {isConnectionActive && (
                    <Button
                        type="button"
                        variant="link"
                        className="h-auto px-0 text-base"
                        onClick={closeConnection}
                    >
                        Stop
                    </Button>
                )}
            </div>

            <div
                className={cn(
                    styles["terminalFrame"],
                    "min-h-0 overflow-hidden border border-slate-800",
                    isFullscreen ? "flex-1" : "w-full",
                )}
                style={isFullscreen ? undefined : { height: TERMINAL_HEIGHT }}
            >
                <div
                    ref={terminalElementRef}
                    className={styles["terminalHost"]}
                />
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
    appID: string;
    supportedShells: string[];
    selectedShell: string;
    onSelectedShellChange: (shell: string) => void;
}
