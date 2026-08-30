import {
    ArrowDownToLine,
    Bug,
    ChevronDown,
    ChevronUp,
    Clock,
    Copy,
    Download,
    LoaderCircle,
    Search,
    TextWrap,
    X,
} from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui";

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
    searchTerm,
    searchResult,
    toolbarStart,
    toolbarFilters,
    onSearchTermChange,
    onFindNext,
    onFindPrevious,
    onToggleTextWrap,
    onToggleTimestamps,
    onToggleDebugLogs,
    onToggleFollowLogs,
    onRefresh,
}: LogsViewerToolbarProps) {
    const textContent = displayedPlainLines.join("\n");

    return (
        <div className="flex flex-col gap-2 sm:gap-2.5 pb-2 sm:pb-2.5 min-w-0 w-full">
            <div className="flex flex-wrap items-center justify-between gap-2 sm:gap-2.5 min-w-0">
                {/* Left group: Stream/Stop */}
                <div className="flex flex-wrap items-center gap-2 sm:gap-2.5 min-w-0">
                    {toolbarStart ?? (
                        <div className="flex min-w-0 items-center gap-2 sm:gap-2.5">
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
                </div>

                {/* Right group: Filter inputs, Search input and Action icons */}
                <div className="flex flex-wrap items-center gap-2.5 sm:gap-3 min-w-0">
                    {toolbarFilters}
                    <div className="relative flex items-center min-w-0 w-full sm:w-56 max-w-full">
                        <Search className="absolute left-2.5 size-3.5 text-muted-foreground pointer-events-none" />
                        <input
                            type="text"
                            placeholder="Find in logs..."
                            value={searchTerm}
                            onChange={e => {
                                onSearchTermChange(e.target.value);
                            }}
                            onKeyDown={e => {
                                if (e.key === "Enter") {
                                    if (e.shiftKey) {
                                        onFindPrevious();
                                    } else {
                                        onFindNext();
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
                                    onClick={onFindPrevious}
                                    className="p-1 hover:text-foreground rounded hover:bg-muted"
                                >
                                    <ChevronUp className="size-3 sm:size-3.5" />
                                </button>
                                <button
                                    type="button"
                                    aria-label="Next match"
                                    onClick={onFindNext}
                                    className="p-1 hover:text-foreground rounded hover:bg-muted"
                                >
                                    <ChevronDown className="size-3 sm:size-3.5" />
                                </button>
                                <button
                                    type="button"
                                    aria-label="Clear search"
                                    onClick={() => {
                                        onSearchTermChange("");
                                    }}
                                    className="p-1 hover:text-foreground rounded hover:bg-muted"
                                >
                                    <X className="size-3 sm:size-3.5" />
                                </button>
                            </div>
                        )}
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
                    </div>
                </div>
            </div>
        </div>
    );
}
