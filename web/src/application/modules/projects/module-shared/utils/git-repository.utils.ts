import { ESettingType } from "@application/shared/enums";
import type { ESettingType as SettingType } from "@application/shared/enums";

export const GIT_SELECTOR_PAGE_SIZE = 50;

export interface ParsedGitRepository {
    owner?: string;
    repo: string;
}

function trimGitSuffix(value: string): string {
    return value.replace(/\.git$/i, "");
}

export function parseGitRepository(repoURL?: string | null): ParsedGitRepository | null {
    const normalized = repoURL?.trim() ?? "";

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

export function truncateSha(sha: string): string {
    return sha.slice(0, 7);
}

export function canUseGitCredentialSelectors(type?: SettingType | null): boolean {
    return type === ESettingType.GithubApp || type === ESettingType.AccessToken;
}
