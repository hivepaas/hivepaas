import type { Key, ReactNode } from "react";

export const ELogsViewerFrameType = {
    In: "in",
    Out: "out",
    Err: "err",
    Warn: "warn",
    Debug: "debug",
} as const;

export type ELogsViewerFrameType = (typeof ELogsViewerFrameType)[keyof typeof ELogsViewerFrameType];

export interface LogsViewerFrame {
    type: ELogsViewerFrameType;
    data: string;
    ts: Date | null;
}

export interface LogsViewerProps {
    frames: LogsViewerFrame[];
    logViewerKey?: Key;
    isStreaming?: boolean;
    isRefreshPending?: boolean;
    hasLineNumbers?: boolean;
    useAnsiClasses?: boolean;
    height?: number | string;
    fullscreenHeight?: number | string;
    isFullView?: boolean;
    isFullHeight?: boolean;
    fontSize?: number;
    themeId?: string;
    downloadFileName?: string;
    defaultShowDebugLogs?: boolean;
    defaultShowTimestamps?: boolean;
    defaultTextWrapped?: boolean;
    toolbarStart?: ReactNode;
    toolbarFilters?: ReactNode;
    className?: string;
    onRefresh?: () => void;
}

export interface LogsViewerSearchResult {
    resultIndex: number;
    resultCount: number;
}

export interface LogsViewerToolbarProps {
    isStreaming: boolean;
    isRefreshPending: boolean;
    displayedPlainLines: string[];
    downloadFileName: string;
    isTextWrapped: boolean;
    showTimestamps: boolean;
    showDebugLogs: boolean;
    followLogs: boolean;
    isFullscreen?: boolean;
    searchTerm: string;
    searchResult?: LogsViewerSearchResult | null;
    toolbarStart?: ReactNode;
    toolbarFilters?: ReactNode;
    onSearchTermChange: (term: string) => void;
    onFindNext: () => void;
    onFindPrevious: () => void;
    onToggleTextWrap: () => void;
    onToggleTimestamps: () => void;
    onToggleDebugLogs: () => void;
    onToggleFollowLogs: () => void;
    onToggleFullscreen?: () => void;
    onRefresh?: () => void;
}

export interface LogsViewerToolbarIconButtonProps {
    label: string;
    isActive?: boolean;
    children: ReactNode;
    onClick: () => void;
}
