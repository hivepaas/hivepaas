/** A call's kind as a person reads it. */
export const CALL_KIND_LABEL: Record<string, string> = {
    http: "HTTP",
    db: "Database",
    rpc: "RPC",
};

export function callKindLabel(kind: string): string {
    return CALL_KIND_LABEL[kind] ?? kind;
}

/** A share of a count, in percent, with the precision a small one needs. */
export function formatShare(part: number, whole: number): string {
    const share = whole > 0 ? (part / whole) * 100 : 0;
    return `${share.toFixed(share < 10 ? 2 : 1)}%`;
}

/** A duration in milliseconds; a dash without one. */
export function formatMs(value: number | null): string {
    return value === null ? "-" : `${value.toFixed(1)} ms`;
}
