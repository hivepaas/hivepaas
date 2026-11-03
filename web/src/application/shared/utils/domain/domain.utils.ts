const RFC_DOMAIN_REGEX = /^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+(?:[a-z]{2,}|xn--[a-z0-9]+)$/i;

// Reject characters that would cause URL() to reparse the input (path, fragment, credentials, whitespace)
const INVALID_CHARS_REGEX = /[\s/\\?#@]/;

export interface IsValidDomainOptions {
    allowWildcard?: boolean;
    /** Mirror of backend base.DomainNameMaxLen. Defaults to 253 (RFC 5890). */
    maxLength?: number;
}

function toAsciiDomain(domain: string): string | null {
    if (INVALID_CHARS_REGEX.test(domain)) {
        return null;
    }
    // ASCII-only fast path — skip URL allocation
    // eslint-disable-next-line no-control-regex
    if (/^[\x00-\x7F]*$/.test(domain)) {
        return domain;
    }
    // Convert IDN/Unicode labels to Punycode via browser URL API (RFC 3492)
    try {
        const { hostname } = new URL(`http://${domain}`);
        return hostname || null;
    } catch {
        return null;
    }
}

/**
 * Validate a domain name according to RFC 1123 / RFC 5890 (FQDN + IDN).
 *
 * Rules (mirrors backend `basedto.ValidateDomain`):
 *  - Each label: 1-63 chars, alphanumeric + hyphens, cannot start/end with hyphen
 *  - TLD: at least 2 alpha chars, or xn-- Punycode prefix
 *  - Total ASCII length ≤ maxLength (default 253; set 100 to match backend DomainNameMaxLen)
 *  - Wildcards (*.example.com) only when allowWildcard = true
 */
export function isValidDomain(domain: string, options: IsValidDomainOptions = {}): boolean {
    const { allowWildcard = false, maxLength = 253 } = options;
    if (typeof domain !== "string") {
        return false;
    }

    let target = domain;
    if (allowWildcard && target.startsWith("*.")) {
        target = target.slice(2);
    }
    if (target.includes("*")) {
        return false;
    }

    const ascii = toAsciiDomain(target);
    if (!ascii || ascii.length > maxLength) {
        return false;
    }

    return RFC_DOMAIN_REGEX.test(ascii);
}

/**
 * Computes default domain based on current window domain.
 * - Current domain `abc.xyz.tuv` -> `<prefix>.xyz.tuv`
 * - Current domain `tuv` -> `<prefix>.tuv`
 */
export function getDefaultDomain(prefix: string = "app"): string {
    if (typeof window === "undefined" || !window.location.hostname) {
        return "";
    }
    let hostname = window.location.hostname.trim().toLowerCase();
    if (hostname.includes(":")) {
        hostname = hostname.split(":")[0] ?? "";
    }
    if (!hostname) {
        return "";
    }

    const tName = prefix.trim().toLowerCase() || "app";
    const segments = hostname.split(".");
    if (segments.length > 1) {
        return `${tName}.${segments.slice(1).join(".")}`;
    }
    return `${tName}.${hostname}`;
}
