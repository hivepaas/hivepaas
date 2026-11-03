import { EAcmeDnsProviderKind } from "@application/shared/enums";

export const ACME_DNS_PROVIDER_OPTIONS = [
    {
        value: EAcmeDnsProviderKind.AcmeDNS,
        label: "ACME DNS",
    },
    {
        value: EAcmeDnsProviderKind.Azure,
        label: "Azure",
    },
    {
        value: EAcmeDnsProviderKind.BaiduCloud,
        label: "Baidu Cloud",
    },
    {
        value: EAcmeDnsProviderKind.Cloudflare,
        label: "Cloudflare",
    },
    {
        value: EAcmeDnsProviderKind.DigitalOcean,
        label: "Digital Ocean",
    },
    {
        value: EAcmeDnsProviderKind.GCloud,
        label: "Google Cloud",
    },
    {
        value: EAcmeDnsProviderKind.GoDaddy,
        label: "Go Daddy",
    },
    {
        value: EAcmeDnsProviderKind.Hetzner,
        label: "Hetzner",
    },
    {
        value: EAcmeDnsProviderKind.HuaweiCloud,
        label: "Huawei Cloud",
    },
    {
        value: EAcmeDnsProviderKind.Namecheap,
        label: "Namecheap",
    },
    {
        value: EAcmeDnsProviderKind.RFC2136,
        label: "RFC2136",
    },
    {
        value: EAcmeDnsProviderKind.Route53,
        label: "Route 53",
    },
    {
        value: EAcmeDnsProviderKind.TencentCloud,
        label: "Tencent Cloud",
    },
] as const;

export function formatAcmeDnsProviderKind(kind: string): string {
    return ACME_DNS_PROVIDER_OPTIONS.find(option => option.value === kind)?.label ?? kind;
}
