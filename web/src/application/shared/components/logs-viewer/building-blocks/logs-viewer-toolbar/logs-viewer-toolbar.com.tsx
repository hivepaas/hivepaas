import { LogViewerSearch } from "@patternfly/react-log-viewer";
import {
    ArrowDownToLine,
    Bug,
    Clock,
    Copy,
    Download,
    LoaderCircle,
    Maximize2,
    Minimize2,
    TextWrap,
} from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui";

import styles from "../../logs-viewer.module.scss";
import type { LogsViewerToolbarProps } from "../../logs-viewer.types";
import { LogsViewerToolbarIconButton } from "../logs-viewer-toolbar-icon-button";

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

export function LogsViewerToolbar({
    isStreaming,
    isRefreshPending,
    displayedPlainLines,
    downloadFileName,
    isTextWrapped,
    showTimestamps,
    showDebugLogs,
    followLogs,
    isFullscreen,
    toolbarStart,
    toolbarFilters,
    onToggleTextWrap,
    onToggleTimestamps,
    onToggleDebugLogs,
    onToggleFollowLogs,
    onToggleFullscreen,
    onRefresh,
}: LogsViewerToolbarProps) {
    const textContent = displayedPlainLines.join("\n");

    return (
        <div className="flex flex-col gap-2.5 pb-2.5 min-w-0 w-full">
            <div className="flex flex-wrap items-center justify-between gap-2.5 min-w-0">
                {/* Left group: Stream/Stop and Filter inputs */}
                <div className="flex flex-wrap items-center gap-2.5 min-w-0">
                    {toolbarStart ?? (
                        <div className="flex min-w-0 items-center gap-2.5">
                            <span className="text-xs sm:text-sm font-semibold text-foreground">Logs</span>
                            {isStreaming && (
                                <span className="flex items-center gap-1.5 sm:gap-2 text-xs sm:text-sm text-emerald-500">
                                    <LoaderCircle className="size-3.5 sm:size-4 animate-spin" />
                                    streaming
                                </span>
                            )}
                            {!isStreaming && onRefresh && (
                                <Button
                                    type="button"
                                    variant="link"
                                    className="h-auto p-0 text-xs sm:text-sm text-primary"
                                    isLoading={isRefreshPending}
                                    onClick={onRefresh}
                                >
                                    Refresh
                                </Button>
                            )}
                        </div>
                    )}
                    {toolbarFilters}
                </div>

                {/* Right group: Search input and Action icons */}
                <div className="flex flex-wrap items-center gap-2 sm:gap-3 min-w-0">
                    <div className="w-full sm:w-48 max-w-full">
                        <LogViewerSearch
                            placeholder="find in logs"
                            minSearchChars={1}
                            className={styles["search"]}
                        />
                    </div>

                    <div className="flex items-center gap-0.5 sm:gap-1">
                        <LogsViewerToolbarIconButton
                            label="Copy logs"
                            onClick={() => {
                                void navigator.clipboard.writeText(textContent);
                                toast.success("Logs copied");
                            }}
                        >
                            <Copy className="size-3.5 sm:size-4" />
                        </LogsViewerToolbarIconButton>
                        <LogsViewerToolbarIconButton
                            label="Download logs"
                            onClick={() => {
                                downloadTextFile(downloadFileName, textContent);
                            }}
                        >
                            <Download className="size-3.5 sm:size-4" />
                        </LogsViewerToolbarIconButton>
                        <LogsViewerToolbarIconButton
                            label={isTextWrapped ? "Disable text wrap" : "Enable text wrap"}
                            isActive={isTextWrapped}
                            onClick={onToggleTextWrap}
                        >
                            <TextWrap className="size-3.5 sm:size-4" />
                        </LogsViewerToolbarIconButton>
                        <LogsViewerToolbarIconButton
                            label={showTimestamps ? "Hide timestamps" : "Show timestamps"}
                            isActive={showTimestamps}
                            onClick={onToggleTimestamps}
                        >
                            <Clock className="size-3.5 sm:size-4" />
                        </LogsViewerToolbarIconButton>
                        <LogsViewerToolbarIconButton
                            label={showDebugLogs ? "Hide debug logs" : "Show debug logs"}
                            isActive={showDebugLogs}
                            onClick={onToggleDebugLogs}
                        >
                            <Bug className="size-3.5 sm:size-4" />
                        </LogsViewerToolbarIconButton>
                        <LogsViewerToolbarIconButton
                            label={followLogs ? "Pause follow logs" : "Follow logs"}
                            isActive={followLogs}
                            onClick={onToggleFollowLogs}
                        >
                            <ArrowDownToLine className="size-3.5 sm:size-4" />
                        </LogsViewerToolbarIconButton>
                        <LogsViewerToolbarIconButton
                            label={isFullscreen ? "Exit fullscreen" : "Fullscreen logs"}
                            isActive={isFullscreen}
                            onClick={onToggleFullscreen}
                        >
                            {isFullscreen ? (
                                <Minimize2 className="size-3.5 sm:size-4" />
                            ) : (
                                <Maximize2 className="size-3.5 sm:size-4" />
                            )}
                        </LogsViewerToolbarIconButton>
                    </div>
                </div>
            </div>
        </div>
    );
}
