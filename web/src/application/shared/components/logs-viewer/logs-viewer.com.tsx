import { type CSSProperties, useEffect, useMemo, useState } from "react";

import { cn } from "@/lib/utils";
import "@patternfly/react-core/dist/styles/base-no-reset.css";
import { LogViewer } from "@patternfly/react-log-viewer";

import { LogsViewerToolbar } from "./building-blocks";
import styles from "./logs-viewer.module.scss";
import type { LogsViewerProps } from "./logs-viewer.types";
import { buildDisplayedLogFrames, getAnsiLogLines, getPlainLogLines } from "./logs-viewer.utils";

const DEFAULT_LOG_VIEWER_HEIGHT = 1_000;
const DEFAULT_FULLSCREEN_LOG_VIEWER_HEIGHT = "auto";
const DEFAULT_DOWNLOAD_FILE_NAME = "logs.txt";
const DEFAULT_LOG_VIEWER_FONT_SIZE = "0.875rem";

export function LogsViewer({
    frames,
    isStreaming = false,
    isRefreshPending = false,
    hasLineNumbers = true,
    height = DEFAULT_LOG_VIEWER_HEIGHT,
    fullscreenHeight = DEFAULT_FULLSCREEN_LOG_VIEWER_HEIGHT,
    fontSize = DEFAULT_LOG_VIEWER_FONT_SIZE,
    downloadFileName = DEFAULT_DOWNLOAD_FILE_NAME,
    defaultShowDebugLogs = false,
    defaultShowTimestamps = false,
    defaultTextWrapped = true,
    toolbarStart,
    toolbarFilters,
    className,
    onRefresh,
}: LogsViewerProps) {
    const [isTextWrapped, setIsTextWrapped] = useState(defaultTextWrapped);
    const [showTimestamps, setShowTimestamps] = useState(defaultShowTimestamps);
    const [showDebugLogs, setShowDebugLogs] = useState(defaultShowDebugLogs);
    const [followLogs, setFollowLogs] = useState(true);
    const [isFullscreen, setIsFullscreen] = useState(false);

    const displayedFrames = useMemo(() => buildDisplayedLogFrames(frames, showDebugLogs), [frames, showDebugLogs]);
    const displayedPlainLines = useMemo(
        () => displayedFrames.flatMap(frame => getPlainLogLines(frame, showTimestamps)),
        [displayedFrames, showTimestamps],
    );
    const displayedAnsiLines = useMemo(
        () => displayedFrames.flatMap(frame => getAnsiLogLines(frame, showTimestamps)),
        [displayedFrames, showTimestamps],
    );
    const scrollToRow = followLogs && displayedAnsiLines.length > 0 ? displayedAnsiLines.length - 1 : undefined;
    const rootStyle = {
        "--logs-viewer-font-size": fontSize,
    } as CSSProperties;

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
                styles["root"],
                "min-w-0",
                className,
                isFullscreen && [
                    styles["fullscreen"],
                    "fixed inset-4 z-50 flex min-h-0 min-w-0 flex-col overflow-hidden rounded-lg border bg-background p-4 shadow-2xl",
                ],
            )}
            style={rootStyle}
        >
            <LogViewer
                key={isFullscreen ? "fullscreen" : "inline"}
                data={displayedAnsiLines}
                hasLineNumbers={hasLineNumbers}
                theme="dark"
                height={isFullscreen ? fullscreenHeight : height}
                scrollToRow={scrollToRow}
                isTextWrapped={isTextWrapped}
                fastRowHeightEstimationLimit={0}
                toolbar={
                    <LogsViewerToolbar
                        isStreaming={isStreaming}
                        isRefreshPending={isRefreshPending}
                        displayedPlainLines={displayedPlainLines}
                        downloadFileName={downloadFileName}
                        isTextWrapped={isTextWrapped}
                        showTimestamps={showTimestamps}
                        showDebugLogs={showDebugLogs}
                        followLogs={followLogs}
                        isFullscreen={isFullscreen}
                        toolbarStart={toolbarStart}
                        toolbarFilters={toolbarFilters}
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
                            setFollowLogs(current => !current);
                        }}
                        onToggleFullscreen={() => {
                            setIsFullscreen(current => !current);
                        }}
                        onRefresh={onRefresh}
                    />
                }
            />
        </div>
    );
}
