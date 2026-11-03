const BYTE_UNITS = ["B", "KiB", "MiB", "GiB", "TiB"];

/** Bytes in binary units, as `docker stats` shows them. */
export function formatBytes(value: number | null): string {
    if (value === null) {
        return "-";
    }
    let unit = 0;
    let n = value;
    while (n >= 1024 && unit < BYTE_UNITS.length - 1) {
        n /= 1024;
        unit++;
    }
    return `${n.toFixed(unit === 0 ? 0 : 1)} ${BYTE_UNITS[unit]}`;
}

/** Cores, with the precision a fraction of one needs. */
export function formatCores(value: number | null): string {
    if (value === null) {
        return "-";
    }
    return value < 1 ? `${value.toFixed(3)} cores` : `${value.toFixed(2)} cores`;
}
