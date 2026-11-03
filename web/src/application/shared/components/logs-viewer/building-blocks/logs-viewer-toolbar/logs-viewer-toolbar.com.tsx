import {
    ArrowDownToLine,
    Bug,
    ChevronDown,
    ChevronUp,
    Clock,
    Copy,
    Download,
    LoaderCircle,
    TextWrap,
} from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui";

import type { LogsViewerToolbarProps } from "../../logs-viewer.types";
import { CollapsibleSearchInput } from "../collapsible-search-input";
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
    searchMode,
    isSearchTermInvalid,
    searchResult,
    toolbarStart,
    toolbarFilters,
    toolbarSearch,
    onSearchTermChange,
    onSearchModeChange,
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
                    {/* Search sits in the icon group so that, collapsed, it is
                        spaced like the icon it looks like. */}
                    <div className="flex min-w-0 flex-wrap items-center gap-0.5 sm:gap-1">
                        {toolbarSearch}
                        <CollapsibleSearchInput
                            value={searchTerm}
                            placeholder="Find in current view..."
                            label="Find in current view..."
                            expandedClassName="w-48 sm:w-72 mr-1.5 sm:mr-2"
                            mode={searchMode}
                            isInvalid={isSearchTermInvalid}
                            onValueChange={onSearchTermChange}
                            onModeChange={onSearchModeChange}
                            onKeyDown={e => {
                                if (e.key === "Enter") {
                                    if (e.shiftKey) {
                                        onFindPrevious();
                                    } else {
                                        onFindNext();
                                    }
                                }
                            }}
                            onClear={() => {
                                onSearchTermChange("");
                            }}
                            trailing={
                                searchTerm ? (
                                    <>
                                        {searchResult && (
                                            <span className="mr-1 font-mono text-[10px] text-muted-foreground sm:text-xs">
                                                {searchResult.resultCount > 0
                                                    ? `${searchResult.resultIndex + 1}/${searchResult.resultCount}`
                                                    : "0/0"}
                                            </span>
                                        )}
                                        <button
                                            type="button"
                                            aria-label="Previous match"
                                            className="rounded p-1 hover:bg-muted hover:text-foreground"
                                            onClick={onFindPrevious}
                                        >
                                            <ChevronUp className="size-3 sm:size-3.5" />
                                        </button>
                                        <button
                                            type="button"
                                            aria-label="Next match"
                                            className="rounded p-1 hover:bg-muted hover:text-foreground"
                                            onClick={onFindNext}
                                        >
                                            <ChevronDown className="size-3 sm:size-3.5" />
                                        </button>
                                    </>
                                ) : undefined
                            }
                        />
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
