import type { ReactNode } from "react";

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
    isStreaming?: boolean;
    isRefreshPending?: boolean;
    hasLineNumbers?: boolean;
    height?: number | string;
    fullscreenHeight?: number | string;
    fontSize?: string;
    downloadFileName?: string;
    defaultShowDebugLogs?: boolean;
    defaultShowTimestamps?: boolean;
    defaultTextWrapped?: boolean;
    toolbarStart?: ReactNode;
    toolbarFilters?: ReactNode;
    className?: string;
    onRefresh?: () => void;
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
    isFullscreen: boolean;
    toolbarStart?: ReactNode;
    toolbarFilters?: ReactNode;
    onToggleTextWrap: () => void;
    onToggleTimestamps: () => void;
    onToggleDebugLogs: () => void;
    onToggleFollowLogs: () => void;
    onToggleFullscreen: () => void;
    onRefresh?: () => void;
}

export interface LogsViewerToolbarIconButtonProps {
    label: string;
    isActive?: boolean;
    children: ReactNode;
    onClick: () => void;
}
