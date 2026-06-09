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
    hasLineNumbers?: boolean;
    height?: number | string;
    fullscreenHeight?: number | string;
    fontSize?: string;
    downloadFileName?: string;
    defaultShowDebugLogs?: boolean;
    defaultShowTimestamps?: boolean;
    defaultTextWrapped?: boolean;
    className?: string;
}

export interface LogsViewerToolbarProps {
    isStreaming: boolean;
    displayedPlainLines: string[];
    downloadFileName: string;
    isTextWrapped: boolean;
    showTimestamps: boolean;
    showDebugLogs: boolean;
    followLogs: boolean;
    isFullscreen: boolean;
    onToggleTextWrap: () => void;
    onToggleTimestamps: () => void;
    onToggleDebugLogs: () => void;
    onToggleFollowLogs: () => void;
    onToggleFullscreen: () => void;
}

export interface LogsViewerToolbarIconButtonProps {
    label: string;
    isActive?: boolean;
    children: ReactNode;
    onClick: () => void;
}
