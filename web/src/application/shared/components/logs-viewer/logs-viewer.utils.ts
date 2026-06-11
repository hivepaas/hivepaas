import { ELogsViewerFrameType, type LogsViewerFrame } from "./logs-viewer.types";

const BACKSPACE_CHARACTER = String.fromCharCode(8);
const ANSI_RESET = "\u001B[0m";
const ANSI_TIMESTAMP_COLOR = "\u001B[38;2;121;170;255m";
const ANSI_LOG_COLORS: Record<ELogsViewerFrameType, string> = {
    [ELogsViewerFrameType.In]: "\u001B[36m",
    [ELogsViewerFrameType.Out]: "",
    [ELogsViewerFrameType.Err]: "\u001B[38;2;238;31;31m",
    [ELogsViewerFrameType.Warn]: "\u001B[38;2;252;212;82m",
    [ELogsViewerFrameType.Debug]: "\u001B[38;2;255;156;253m",
};

export function parseLogsViewerFrames(rawMessage: string): LogsViewerFrame[] {
    if (!rawMessage.trim()) {
        return [];
    }

    const parsed = JSON.parse(rawMessage) as unknown;
    const frames = Array.isArray(parsed) ? parsed : [parsed];

    return frames
        .map(frame => normalizeLogsViewerFrame(frame))
        .filter((frame): frame is LogsViewerFrame => frame !== null);
}

export function normalizeLogsViewerFrame(frame: unknown): LogsViewerFrame | null {
    if (!frame || typeof frame !== "object") {
        return null;
    }

    const input = frame as Record<string, unknown>;
    const { type, data, ts } = input;

    if (!isLogsViewerFrameType(type) || typeof data !== "string") {
        return null;
    }

    return {
        type,
        data,
        ts: typeof ts === "string" && ts.trim() ? new Date(ts) : null,
    };
}

export function buildDisplayedLogFrames(frames: LogsViewerFrame[], showDebugLogs: boolean): LogsViewerFrame[] {
    if (showDebugLogs) {
        return frames;
    }

    return frames.filter(frame => frame.type !== ELogsViewerFrameType.Debug);
}

export function getPlainLogLines(frame: LogsViewerFrame, showTimestamps: boolean): string[] {
    const prefix = getTimestampPrefix(frame, showTimestamps);

    return normalizeLogTextForDisplay(frame.data).map(line => `${prefix}${line}`);
}

export function getAnsiLogLines(frame: LogsViewerFrame, showTimestamps: boolean): string[] {
    const prefix = getAnsiTimestampPrefix(frame, showTimestamps);
    const color = ANSI_LOG_COLORS[frame.type];

    return normalizeLogTextForDisplay(frame.data).map(line => {
        if (!color) {
            return `${prefix}${line}`;
        }

        return `${prefix}${color}${line}${ANSI_RESET}`;
    });
}

export function normalizeLogTextForDisplay(data: string): string[] {
    const lines = data.replace(/\r\n/g, "\n").replace(/\r/g, "\n").replaceAll(BACKSPACE_CHARACTER, "").split("\n");

    if (lines.length > 1 && lines[lines.length - 1] === "") {
        lines.pop();
    }

    return lines.length > 0 ? lines : [""];
}

export function formatLogTimestamp(timestamp: Date | null): string {
    if (!timestamp || Number.isNaN(timestamp.getTime())) {
        return "-";
    }

    const year = timestamp.getFullYear();
    const month = formatDatePart(timestamp.getMonth() + 1);
    const date = formatDatePart(timestamp.getDate());
    const hours = formatDatePart(timestamp.getHours());
    const minutes = formatDatePart(timestamp.getMinutes());
    const seconds = formatDatePart(timestamp.getSeconds());
    const milliseconds = formatDatePart(timestamp.getMilliseconds(), 3);

    return `${year}-${month}-${date} ${hours}:${minutes}:${seconds}.${milliseconds}`;
}

function isLogsViewerFrameType(value: unknown): value is ELogsViewerFrameType {
    return Object.values(ELogsViewerFrameType).includes(value as ELogsViewerFrameType);
}

function getTimestampPrefix(frame: LogsViewerFrame, showTimestamps: boolean): string {
    if (!showTimestamps) {
        return "";
    }

    return `[${formatLogTimestamp(frame.ts)}] `;
}

function getAnsiTimestampPrefix(frame: LogsViewerFrame, showTimestamps: boolean): string {
    if (!showTimestamps) {
        return "";
    }

    return `${ANSI_TIMESTAMP_COLOR}[${formatLogTimestamp(frame.ts)}]${ANSI_RESET} `;
}

function formatDatePart(value: number, length: number = 2): string {
    return String(value).padStart(length, "0");
}
