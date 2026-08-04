export const PROXY_PROVIDER_UNSPECIFIED = "unspecified" as const;

export type HivePaaSKnownProxyProvider =
    | "cloudflare"
    | "fastly"
    | "aws-cloudfront"
    | "imperva"
    | typeof PROXY_PROVIDER_UNSPECIFIED;

export const PROXY_PROVIDER_OPTIONS: string[] = [
    "cloudflare",
    "fastly",
    "aws-cloudfront",
    "imperva",
    PROXY_PROVIDER_UNSPECIFIED,
];

export const PROXY_PROVIDER_IP_URLS: Partial<Record<HivePaaSKnownProxyProvider, string>> = {
    "cloudflare": "https://www.cloudflare.com/ips",
    "fastly": "https://api.fastly.com/public-ip-list",
    "aws-cloudfront": "https://ip-ranges.amazonaws.com/ip-ranges.json",
    "imperva": "https://my.incapsula.com/api/integration/v1/ips",
};

export function splitTrustedIPsText(text: string): string[] {
    return text
        .split(/[\s,]+/)
        .map(entry => entry.trim())
        .filter(Boolean);
}

export function joinTrustedIPsText(trustedIPs: string[]): string {
    return trustedIPs.join("\n");
}
