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

export function formatFrameForXterm(frame: LogsViewerFrame, showTimestamps: boolean): string {
    const lines = getAnsiLogLines(frame, showTimestamps);
    return lines.join("\r\n") + "\r\n";
}

export function formatFramesForXterm(frames: LogsViewerFrame[], showTimestamps: boolean): string {
    return frames.map(frame => formatFrameForXterm(frame, showTimestamps)).join("");
}

/**
 * logsViewerFrameSignature identifies a frame by value.
 *
 * The frames array is rebuilt from scratch whenever its source changes, so the
 * objects in it are never the same objects twice and identity says nothing.
 */
export function logsViewerFrameSignature(frame: LogsViewerFrame | undefined): string {
    if (!frame) {
        return "";
    }
    return `${frame.ts?.getTime() ?? ""}|${frame.type}|${frame.data}`;
}

/**
 * isLogsViewerFramesAppend says whether frames only grew at the end since the
 * rendered state described by anchor.
 *
 * Everything that consumes frames does so incrementally, which is only sound
 * when new frames arrive at the end - a live stream. Stored logs page backwards
 * and put the new page at the FRONT, and an incremental consumer that assumes
 * otherwise writes the wrong slice: the newest lines a second time, while the
 * older ones it was asked for never appear. Whoever cannot prove an append has
 * to rebuild from the whole array.
 */
export function isLogsViewerFramesAppend(frames: LogsViewerFrame[], anchor: LogsViewerFramesAnchor): boolean {
    if (anchor.length === 0 || frames.length < anchor.length) {
        return false;
    }
    return (
        logsViewerFrameSignature(frames[0]) === anchor.first &&
        logsViewerFrameSignature(frames[anchor.length - 1]) === anchor.last
    );
}

/** anchorLogsViewerFrames records what a consumer has taken in, for the check above. */
export function anchorLogsViewerFrames(frames: LogsViewerFrame[]): LogsViewerFramesAnchor {
    return {
        length: frames.length,
        first: logsViewerFrameSignature(frames[0]),
        last: logsViewerFrameSignature(frames[frames.length - 1]),
    };
}

export interface LogsViewerFramesAnchor {
    length: number;
    first: string;
    last: string;
}

/**
 * isValidRegex says whether a pattern can be used as one yet.
 *
 * A pattern half-typed - "(err" - is not an error to report, it is a person
 * mid-keystroke. Both search paths check this so that neither reports "no
 * matches" for something that was never run.
 */
export function isValidRegex(pattern: string): boolean {
    if (!pattern) {
        return true;
    }
    try {
        new RegExp(pattern);
        return true;
    } catch {
        return false;
    }
}
