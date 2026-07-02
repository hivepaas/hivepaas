export { parseGitRepository, truncateSha } from "~/projects/module-shared/utils";
export type { ParsedGitRepository } from "~/projects/module-shared/utils";

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
