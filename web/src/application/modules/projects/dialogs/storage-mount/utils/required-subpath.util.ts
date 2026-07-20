/**
 * Mirrors BE `StorageBindSettings.CaclRequiredSubpath` / volume / cluster
 * (`hivepaas_app/entity/setting_storage_settings.go`).
 */

export interface SubpathPrefixSettings {
    enabled?: boolean;
    baseSubpath?: string;
    appsMustUseSubPaths?: boolean;
}

function normalizeSegment(segment: string): string {
    return segment.replace(/\\/g, "/").replace(/^\/+|\/+$/g, "");
}

/** Join path segments with `/` like Go `filepath.Join` on Unix. */
export function joinPathSegments(...parts: string[]): string {
    return parts.map(normalizeSegment).filter(Boolean).join("/");
}

export function computeRequiredSubpath(
    s: SubpathPrefixSettings | undefined,
    projectKey: string | undefined,
    appKey: string | undefined,
): string {
    if (!s?.enabled) {
        return "";
    }
    const base = normalizeSegment(s.baseSubpath ?? "");
    if (!s.appsMustUseSubPaths) {
        return base;
    }
    return joinPathSegments(base, projectKey ?? "", appKey ?? "");
}
