export const DATA_SIZE_CONSTANTS = {
    B: 1,
    KB: 1024,
    MB: 1024 * 1024,
    GB: 1024 * 1024 * 1024,
    TB: 1024 * 1024 * 1024 * 1024,
    PB: 1024 * 1024 * 1024 * 1024 * 1024,
} as const;

/**
 * Formats a byte size into a human-readable data size string with spaces (e.g. "512 MB", "1 GB").
 */
export function getFriendlyDataSize(numBytes: number | null | undefined): string {
    if (!numBytes) return "";

    const { KB, MB, GB, TB } = DATA_SIZE_CONSTANTS;

    if (numBytes >= TB) {
        let str = (numBytes / TB).toFixed(2);
        if (str.endsWith(".00")) {
            str = str.slice(0, -3);
        }
        return `${str} TB`;
    }

    if (numBytes >= GB) {
        let str = (numBytes / GB).toFixed(2);
        if (str.endsWith(".00")) {
            str = str.slice(0, -3);
        }
        return `${str} GB`;
    }

    if (numBytes >= MB) {
        let str = (numBytes / MB).toFixed(1);
        if (str.endsWith(".0")) {
            str = str.slice(0, -2);
        }
        return `${str} MB`;
    }

    if (numBytes >= KB) {
        return `${Math.ceil(numBytes / KB)} KB`;
    }

    return `${numBytes} B`;
}

/**
 * Converts a data size representation (e.g. "512MB", "1GB", "128mb", "1024b", "2tb", 1024) into bytes number.
 * Returns null if the value cannot be parsed.
 */
export function parseDataSizeToBytes(value: unknown): number | null {
    if (value === null || value === undefined) {
        return null;
    }

    if (typeof value === "number") {
        return !isNaN(value) && value >= 0 ? Math.floor(value) : null;
    }

    if (typeof value !== "string") {
        return null;
    }

    const trimmed = value.trim();
    if (trimmed === "") {
        return null;
    }

    const match = /^([0-9]+(?:\.[0-9]+)?)\s*([a-zA-Z]+)?$/.exec(trimmed);
    if (!match) {
        return null;
    }

    const numStr = match[1];
    const unitStr = (match[2] ?? "").toLowerCase();
    if (!numStr) {
        return null;
    }

    const num = parseFloat(numStr);
    if (isNaN(num) || num < 0) {
        return null;
    }

    const { B, KB, MB, GB, TB, PB } = DATA_SIZE_CONSTANTS;

    switch (unitStr) {
        case "":
        case "b":
        case "byte":
        case "bytes":
            return Math.round(num * B);

        case "k":
        case "kb":
        case "kib":
            return Math.round(num * KB);

        case "m":
        case "mb":
        case "mib":
            return Math.round(num * MB);

        case "g":
        case "gb":
        case "gib":
            return Math.round(num * GB);

        case "t":
        case "tb":
        case "tib":
            return Math.round(num * TB);

        case "p":
        case "pb":
        case "pib":
            return Math.round(num * PB);

        default:
            return null;
    }
}

/**
 * Formats a byte size into a compact unit string without spaces (e.g. "128MB", "1GB", "2GB", "4GB").
 */
export function formatDataSizeCompact(numBytes: number | null | undefined): string {
    if (numBytes === null || numBytes === undefined || isNaN(numBytes) || numBytes <= 0) {
        return "";
    }

    const { KB, MB, GB, TB } = DATA_SIZE_CONSTANTS;

    if (numBytes >= TB) {
        const val = numBytes / TB;
        const str = Number.isInteger(val) ? val.toString() : parseFloat(val.toFixed(2)).toString();
        return `${str}TB`;
    }

    if (numBytes >= GB) {
        const val = numBytes / GB;
        const str = Number.isInteger(val) ? val.toString() : parseFloat(val.toFixed(2)).toString();
        return `${str}GB`;
    }

    if (numBytes >= MB) {
        const val = numBytes / MB;
        const str = Number.isInteger(val) ? val.toString() : parseFloat(val.toFixed(2)).toString();
        return `${str}MB`;
    }

    if (numBytes >= KB) {
        const val = numBytes / KB;
        const str = Number.isInteger(val) ? val.toString() : parseFloat(val.toFixed(2)).toString();
        return `${str}KB`;
    }

    return `${numBytes}B`;
}

/**
 * Generates an array of size presets starting from min (or fallback), multiplying by 2 for each step.
 * By default returns 6 values (e.g. min 128MB -> ["128MB", "256MB", "512MB", "1GB", "2GB", "4GB"]).
 */
export function generateDataSizePresets(
    minInput?: unknown,
    defaultInput?: unknown,
    count: number = 6,
    maxInput?: unknown,
): string[] {
    const minBytes = parseDataSizeToBytes(minInput);
    let baseBytes: number;

    if (minBytes !== null && minBytes > 0) {
        baseBytes = minBytes;
    } else {
        const defBytes = parseDataSizeToBytes(defaultInput);
        if (defBytes !== null && defBytes > 0) {
            baseBytes = Math.min(defBytes, 128 * DATA_SIZE_CONSTANTS.MB);
        } else {
            baseBytes = 128 * DATA_SIZE_CONSTANTS.MB;
        }
    }

    const MAX_STEP_INCREMENT = 8 * DATA_SIZE_CONSTANTS.GB;
    const maxBytes = parseDataSizeToBytes(maxInput);

    const presets: string[] = [];
    let currentBytes = baseBytes;

    for (let i = 0; i < count; i++) {
        if (maxBytes !== null && maxBytes > 0 && currentBytes > maxBytes) {
            break;
        }
        const formatted = formatDataSizeCompact(currentBytes);
        if (formatted) {
            presets.push(formatted);
        }
        const stepIncrement = Math.min(currentBytes, MAX_STEP_INCREMENT);
        currentBytes += stepIncrement;
    }

    return presets;
}
