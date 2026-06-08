import { type ReactNode, useEffect, useMemo, useRef, useState } from "react";

import { cn } from "@/lib/utils";
import { Toolbar, ToolbarContent, ToolbarItem } from "@patternfly/react-core";
import "@patternfly/react-core/dist/styles/base-no-reset.css";
import { LogViewer, LogViewerSearch } from "@patternfly/react-log-viewer";
import { ArrowDownToLine, Bug, Copy, Download, LoaderCircle, Maximize2, Minimize2 } from "lucide-react";
import { toast } from "sonner";
import { AppDeploymentsQueries } from "~/projects/data";
import type { AppDeploymentLogFrame } from "~/projects/domain";
import { EAppDeploymentLogType } from "~/projects/domain";

import { EnvConfig } from "@config";

import { Button, Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui";

import styles from "./deployment-logs-viewer.module.scss";

const DEFAULT_LOG_VIEWER_HEIGHT = 1_000;
const DEFAULT_LOG_TAIL = 100;

type WebSocketReadyState = WebSocket["readyState"];

const ANSI_RESET = "\u001B[0m";
const ANSI_LOG_COLORS: Record<EAppDeploymentLogType, string> = {
    [EAppDeploymentLogType.In]: "\u001B[36m",
    [EAppDeploymentLogType.Out]: "",
    [EAppDeploymentLogType.Err]: "\u001B[31m",
    [EAppDeploymentLogType.Warn]: "\u001B[33m",
    [EAppDeploymentLogType.Debug]: "\u001B[90m",
};

function buildDeploymentLogsWsUrl({ projectID, appID, deploymentID, token }: DeploymentLogsWsUrlParams): string {
    const baseUrl = EnvConfig.API_URL.endsWith("/") ? EnvConfig.API_URL : `${EnvConfig.API_URL}/`;
    const url = new URL(
        `projects/${encodeURIComponent(projectID)}/apps/${encodeURIComponent(appID)}/deployments/${encodeURIComponent(deploymentID)}/logs`,
        baseUrl,
    );

    url.protocol = url.protocol === "https:" ? "wss:" : "ws:";
    appendDeploymentLogsQuery(url, token);

    return url.toString();
}

function appendDeploymentLogsQuery(url: URL, token: string) {
    url.searchParams.set("follow", "true");
    url.searchParams.set("tail", String(DEFAULT_LOG_TAIL));
    url.searchParams.set("token", token);
}

async function readWebSocketMessageData(data: MessageEvent["data"]): Promise<string> {
    if (typeof data === "string") {
        return data;
    }

    if (data instanceof Blob) {
        return data.text();
    }

    if (data instanceof ArrayBuffer) {
        return new TextDecoder().decode(data);
    }

    return "";
}

function parseLogFrames(rawMessage: string): AppDeploymentLogFrame[] {
    if (!rawMessage.trim()) {
        return [];
    }

    const parsed = JSON.parse(rawMessage) as unknown;
    const frames = Array.isArray(parsed) ? parsed : [parsed];

    return frames
        .map(frame => normalizeLogFrame(frame))
        .filter((frame): frame is AppDeploymentLogFrame => frame !== null);
}

function normalizeLogFrame(frame: unknown): AppDeploymentLogFrame | null {
    if (!frame || typeof frame !== "object") {
        return null;
    }

    const input = frame as Record<string, unknown>;
    const { type, data, ts } = input;

    if (!isDeploymentLogType(type) || typeof data !== "string") {
        return null;
    }

    return {
        type,
        data,
        ts: typeof ts === "string" && ts.trim() ? new Date(ts) : null,
    };
}

function isDeploymentLogType(value: unknown): value is EAppDeploymentLogType {
    return Object.values(EAppDeploymentLogType).includes(value as EAppDeploymentLogType);
}

function getPlainLogLine(frame: AppDeploymentLogFrame): string {
    return frame.data;
}

function getAnsiLogLine(frame: AppDeploymentLogFrame): string {
    const color = ANSI_LOG_COLORS[frame.type];

    if (!color) {
        return frame.data;
    }

    return `${color}${frame.data}${ANSI_RESET}`;
}

function buildDisplayedLogs(logs: AppDeploymentLogFrame[], showDebugLogs: boolean): AppDeploymentLogFrame[] {
    if (showDebugLogs) {
        return logs;
    }

    return logs.filter(log => log.type !== EAppDeploymentLogType.Debug);
}

function downloadTextFile(fileName: string, content: string) {
    const url = URL.createObjectURL(new Blob([content], { type: "text/plain;charset=utf-8" }));
    const link = document.createElement("a");

    link.href = url;
    link.download = fileName;
    document.body.appendChild(link);
    link.click();
    link.remove();
    URL.revokeObjectURL(url);
}

function ToolbarIconButton({
    label,
    isActive = false,
    children,
    onClick,
}: {
    label: string;
    isActive?: boolean;
    children: ReactNode;
    onClick: () => void;
}) {
    return (
        <Tooltip>
            <TooltipTrigger asChild>
                <Button
                    type="button"
                    variant="ghost"
                    size="icon-sm"
                    className={cn("text-muted-foreground hover:text-foreground", isActive && "text-primary")}
                    aria-label={label}
                    title={label}
                    aria-pressed={isActive}
                    onClick={onClick}
                >
                    {children}
                </Button>
            </TooltipTrigger>
            <TooltipContent side="top">{label}</TooltipContent>
        </Tooltip>
    );
}

function DeploymentLogsToolbar({
    isStreaming,
    displayedPlainLines,
    showDebugLogs,
    followLogs,
    isFullscreen,
    onToggleDebugLogs,
    onToggleFollowLogs,
    onToggleFullscreen,
}: DeploymentLogsToolbarProps) {
    const textContent = displayedPlainLines.join("\n");

    return (
        <Toolbar
            hasNoPadding
            className={styles["toolbar"]}
        >
            <ToolbarContent alignItems="center">
                <ToolbarItem>
                    <div className="flex min-w-0 items-center gap-3">
                        <span className="text-sm font-semibold text-foreground">Logs</span>
                        {isStreaming && (
                            <span className="flex items-center gap-2 text-sm text-rose-500">
                                <LoaderCircle className="size-4 animate-spin" />
                                streaming
                            </span>
                        )}
                    </div>
                </ToolbarItem>

                <ToolbarItem
                    align={{ default: "alignEnd" }}
                    className={styles["searchItem"]}
                >
                    <LogViewerSearch
                        placeholder="find in logs"
                        minSearchChars={1}
                        className={styles["search"]}
                    />
                </ToolbarItem>

                <ToolbarItem className={styles["actions"]}>
                    <ToolbarIconButton
                        label="Copy logs"
                        onClick={() => {
                            void navigator.clipboard.writeText(textContent);
                            toast.success("Logs copied");
                        }}
                    >
                        <Copy className="size-4" />
                    </ToolbarIconButton>
                    <ToolbarIconButton
                        label="Download logs"
                        onClick={() => {
                            downloadTextFile("deployment-logs.txt", textContent);
                        }}
                    >
                        <Download className="size-4" />
                    </ToolbarIconButton>
                    <ToolbarIconButton
                        label={showDebugLogs ? "Hide debug logs" : "Show debug logs"}
                        isActive={showDebugLogs}
                        onClick={onToggleDebugLogs}
                    >
                        <Bug className="size-4" />
                    </ToolbarIconButton>
                    <ToolbarIconButton
                        label={followLogs ? "Pause follow logs" : "Follow logs"}
                        isActive={followLogs}
                        onClick={onToggleFollowLogs}
                    >
                        <ArrowDownToLine className="size-4" />
                    </ToolbarIconButton>
                    <ToolbarIconButton
                        label={isFullscreen ? "Exit fullscreen" : "Fullscreen logs"}
                        isActive={isFullscreen}
                        onClick={onToggleFullscreen}
                    >
                        {isFullscreen ? <Minimize2 className="size-4" /> : <Maximize2 className="size-4" />}
                    </ToolbarIconButton>
                </ToolbarItem>
            </ToolbarContent>
        </Toolbar>
    );
}

export function DeploymentLogsViewer({ projectID, appID, deploymentID }: DeploymentLogsViewerProps) {
    const [logs, setLogs] = useState<AppDeploymentLogFrame[]>([]);
    const [webSocketReadyState, setWebSocketReadyState] = useState<WebSocketReadyState>(WebSocket.CLOSED);
    const [showDebugLogs, setShowDebugLogs] = useState(false);
    const [followLogs, setFollowLogs] = useState(true);
    const [isFullscreen, setIsFullscreen] = useState(false);
    const webSocketRef = useRef<WebSocket | null>(null);
    const logsTokenRequest = useMemo(() => ({ projectID, appID, deploymentID }), [projectID, appID, deploymentID]);
    const { refetch: fetchLogsToken } = AppDeploymentsQueries.useGetLogsToken(logsTokenRequest, {
        enabled: false,
        staleTime: 0,
        gcTime: 0,
    });

    const displayedLogs = useMemo(() => buildDisplayedLogs(logs, showDebugLogs), [logs, showDebugLogs]);
    const displayedPlainLines = useMemo(() => displayedLogs.map(getPlainLogLine), [displayedLogs]);
    const displayedAnsiLines = useMemo(() => displayedLogs.map(getAnsiLogLine), [displayedLogs]);
    const scrollToRow = followLogs && displayedAnsiLines.length > 0 ? displayedAnsiLines.length - 1 : undefined;
    const isStreaming = webSocketReadyState === WebSocket.OPEN;

    useEffect(() => {
        let isDisposed = false;
        let ws: WebSocket | null = null;

        function updateReadyState(currentWebSocket: WebSocket) {
            if (isDisposed || webSocketRef.current !== currentWebSocket) {
                return;
            }

            setWebSocketReadyState(currentWebSocket.readyState);
        }

        async function connect() {
            try {
                const logsTokenResponse = await fetchLogsToken({ throwOnError: true });
                const token = logsTokenResponse.data?.data.token;

                if (isDisposed || !token) {
                    return;
                }

                const wsUrl = buildDeploymentLogsWsUrl({ projectID, appID, deploymentID, token });
                const currentWebSocket = new WebSocket(wsUrl);
                ws = currentWebSocket;
                webSocketRef.current = currentWebSocket;
                setWebSocketReadyState(currentWebSocket.readyState);

                currentWebSocket.addEventListener("open", () => {
                    updateReadyState(currentWebSocket);
                });

                currentWebSocket.addEventListener("message", event => {
                    void readWebSocketMessageData(event.data)
                        .then(parseLogFrames)
                        .then(frames => {
                            if (frames.length === 0) {
                                return;
                            }

                            setLogs(current => [...current, ...frames]);
                        })
                        .catch((error: unknown) => {
                            console.error("Failed to parse deployment log frame", error);
                        });
                });

                currentWebSocket.addEventListener("close", () => {
                    updateReadyState(currentWebSocket);
                    if (webSocketRef.current === currentWebSocket) {
                        webSocketRef.current = null;
                    }
                });

                currentWebSocket.addEventListener("error", () => {
                    if (isDisposed || webSocketRef.current !== currentWebSocket) {
                        return;
                    }

                    setWebSocketReadyState(WebSocket.CLOSING);
                    if (
                        currentWebSocket.readyState !== WebSocket.CLOSING &&
                        currentWebSocket.readyState !== WebSocket.CLOSED
                    ) {
                        currentWebSocket.close();
                    }
                });
            } catch (error) {
                if (!isDisposed) {
                    console.error("Failed to connect deployment logs", error);
                    setWebSocketReadyState(WebSocket.CLOSED);
                }
            }
        }

        void connect();

        return () => {
            isDisposed = true;
            setWebSocketReadyState(WebSocket.CLOSED);
            ws?.close();
            if (webSocketRef.current === ws) {
                webSocketRef.current = null;
            }
        };
    }, [appID, deploymentID, fetchLogsToken, projectID]);

    return (
        <div
            className={cn(
                styles["root"],
                "min-w-0",
                isFullscreen && "fixed inset-4 z-50 overflow-auto rounded-lg border bg-background p-4 shadow-2xl",
            )}
        >
            <LogViewer
                data={displayedAnsiLines}
                hasLineNumbers
                theme="dark"
                height={isFullscreen ? "calc(100vh - 9rem)" : DEFAULT_LOG_VIEWER_HEIGHT}
                scrollToRow={scrollToRow}
                isTextWrapped={false}
                toolbar={
                    <DeploymentLogsToolbar
                        isStreaming={isStreaming}
                        displayedPlainLines={displayedPlainLines}
                        showDebugLogs={showDebugLogs}
                        followLogs={followLogs}
                        isFullscreen={isFullscreen}
                        onToggleDebugLogs={() => {
                            setShowDebugLogs(current => !current);
                        }}
                        onToggleFollowLogs={() => {
                            setFollowLogs(current => !current);
                        }}
                        onToggleFullscreen={() => {
                            setIsFullscreen(current => !current);
                        }}
                    />
                }
            />
        </div>
    );
}

interface DeploymentLogsScope {
    projectID: string;
    appID: string;
    deploymentID: string;
}

type DeploymentLogsWsUrlParams = DeploymentLogsScope & {
    token: string;
};

type DeploymentLogsViewerProps = DeploymentLogsScope;

interface DeploymentLogsToolbarProps {
    isStreaming: boolean;
    displayedPlainLines: string[];
    showDebugLogs: boolean;
    followLogs: boolean;
    isFullscreen: boolean;
    onToggleDebugLogs: () => void;
    onToggleFollowLogs: () => void;
    onToggleFullscreen: () => void;
}
