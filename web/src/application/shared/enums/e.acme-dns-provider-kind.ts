export const EAcmeDnsProviderKind = {
    AcmeDNS: "acmedns",
    Azure: "azure",
    BaiduCloud: "baiducloud",
    Cloudflare: "cloudflare",
    DigitalOcean: "digitalocean",
    GCloud: "gcloud",
    GoDaddy: "godaddy",
    Hetzner: "hetzner",
    HuaweiCloud: "huaweicloud",
    Namecheap: "namecheap",
    RFC2136: "rfc2136",
    Route53: "route53",
    TencentCloud: "tencentcloud",
} as const;

export type EAcmeDnsProviderKind = (typeof EAcmeDnsProviderKind)[keyof typeof EAcmeDnsProviderKind];
