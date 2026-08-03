export const PROXY_PROVIDER_UNSPECIFIED = "" as const;

export type HivePaaSProxyProviderValue =
    | typeof PROXY_PROVIDER_UNSPECIFIED
    | "cloudflare"
    | "fastly"
    | "aws-cloudfront"
    | "imperva";

/** Radix Select does not support empty string values. */
export const PROXY_PROVIDER_SELECT_UNSPECIFIED = "__unspecified__" as const;

export const PROXY_PROVIDER_OPTIONS: { label: string; value: HivePaaSProxyProviderValue }[] = [
    { label: "Cloudflare", value: "cloudflare" },
    { label: "Fastly", value: "fastly" },
    { label: "AWS CloudFront", value: "aws-cloudfront" },
    { label: "Imperva", value: "imperva" },
    { label: "unspecified", value: PROXY_PROVIDER_UNSPECIFIED },
];

export const PROXY_PROVIDER_IP_URLS: Partial<Record<Exclude<HivePaaSProxyProviderValue, "">, string>> = {
    "cloudflare": "https://www.cloudflare.com/ips",
    "fastly": "https://api.fastly.com/public-ip-list",
    "aws-cloudfront": "https://ip-ranges.amazonaws.com/ip-ranges.json",
    "imperva": "https://my.incapsula.com/api/integration/v1/ips",
};

export function proxyProviderToSelectValue(provider: string): string {
    return provider === PROXY_PROVIDER_UNSPECIFIED ? PROXY_PROVIDER_SELECT_UNSPECIFIED : provider;
}

export function selectValueToProxyProvider(value: string): HivePaaSProxyProviderValue {
    return value === PROXY_PROVIDER_SELECT_UNSPECIFIED
        ? PROXY_PROVIDER_UNSPECIFIED
        : (value as HivePaaSProxyProviderValue);
}

const KNOWN_PROXY_PROVIDERS = new Set<HivePaaSProxyProviderValue>([
    PROXY_PROVIDER_UNSPECIFIED,
    "cloudflare",
    "fastly",
    "aws-cloudfront",
    "imperva",
]);

export function normalizeProxyProvider(provider: string): HivePaaSProxyProviderValue {
    return KNOWN_PROXY_PROVIDERS.has(provider as HivePaaSProxyProviderValue)
        ? (provider as HivePaaSProxyProviderValue)
        : PROXY_PROVIDER_UNSPECIFIED;
}

export function splitTrustedIPsText(text: string): string[] {
    return text
        .split(/[\s,]+/)
        .map(entry => entry.trim())
        .filter(Boolean);
}

export function joinTrustedIPsText(trustedIPs: string[]): string {
    return trustedIPs.join("\n");
}
