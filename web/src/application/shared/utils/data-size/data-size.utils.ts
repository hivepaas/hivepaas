/**
 * Formats a byte size into a human-readable data size string.
 */
export function getFriendlyDataSize(numBytes: number | null | undefined): string {
    if (!numBytes) return "";

    const KB = 1024;
    const MB = KB * KB;
    const GB = MB * KB;
    const TB = GB * KB;

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
