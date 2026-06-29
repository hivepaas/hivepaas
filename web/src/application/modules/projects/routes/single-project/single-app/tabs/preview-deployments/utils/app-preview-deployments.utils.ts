export interface ParsedGitRepository {
    owner?: string;
    repo: string;
}

export function getPreviewSubdomain(accessLinks: readonly string[]): string | null {
    const accessLink = accessLinks[0];

    if (!accessLink) {
        return null;
    }

    const [subdomain] = accessLink
        .replace(/^https?:\/\//i, "")
        .split("/")
        .filter(Boolean);

    return subdomain ?? null;
}

export function getExternalPreviewUrl(accessLink: string): string {
    if (/^https?:\/\//i.test(accessLink)) {
        return accessLink;
    }

    return `https://${accessLink}`;
}

export function truncateSha(sha: string): string {
    return sha.slice(0, 7);
}

function trimGitSuffix(value: string): string {
    return value.replace(/\.git$/i, "");
}

export function parseGitRepository(repoURL: string): ParsedGitRepository | null {
    const normalized = repoURL.trim();

    if (!normalized) {
        return null;
    }

    const scpLikeMatch = /^[^@]+@[^:]+:(?<path>.+)$/.exec(normalized);
    const pathValue =
        scpLikeMatch?.groups?.["path"] ??
        (() => {
            try {
                const url = new URL(normalized);
                return url.pathname;
            } catch {
                return normalized.replace(/^https?:\/\//i, "");
            }
        })();

    const segments = pathValue
        .replace(/^\/+/, "")
        .split("/")
        .map(segment => segment.trim())
        .filter(Boolean);

    if (segments.length === 0) {
        return null;
    }

    const repoSegment = segments.at(-1);

    if (!repoSegment) {
        return null;
    }

    const repo = trimGitSuffix(repoSegment);

    if (!repo) {
        return null;
    }

    const ownerSegments = segments.slice(0, -1);

    return {
        repo,
        ...(ownerSegments.length > 0 ? { owner: ownerSegments.join("/") } : {}),
    };
}
