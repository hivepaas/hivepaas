import { z } from "zod";

import { EAcmeDnsProviderKind } from "@application/shared/enums";

const MAX_LEN_200 = 200;
const MAX_LEN_1000 = 1000;
const MAX_LEN_10000 = 10000;

function trimmedString(maxLength: number, label: string) {
    return z.string().trim().max(maxLength, `${label} must be at most ${maxLength} characters`);
}

function optionalTrimmedString(maxLength: number, label: string) {
    return z
        .string()
        .optional()
        .transform(value => value ?? "")
        .pipe(trimmedString(maxLength, label));
}

function requireField(
    ctx: z.RefinementCtx,
    value: string,
    path: keyof CreateOrEditAcmeDnsProviderFormInput,
    label: string,
) {
    if (value.trim()) {
        return;
    }

    ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: [path],
        message: `${label} is required`,
    });
}

export const CreateOrEditAcmeDnsProviderFormSchema = z
    .object({
        name: trimmedString(MAX_LEN_200, "Name").min(1, "Name is required"),
        kind: z.nativeEnum(EAcmeDnsProviderKind),
        acmeDnsApiBase: optionalTrimmedString(MAX_LEN_1000, "API Base"),
        acmeDnsAllowList: optionalTrimmedString(MAX_LEN_1000, "Allow List"),
        acmeDnsStoragePath: optionalTrimmedString(MAX_LEN_1000, "Storage Path"),
        acmeDnsStorageBaseUrl: optionalTrimmedString(MAX_LEN_1000, "Storage Base URL"),
        azureClientId: optionalTrimmedString(MAX_LEN_200, "Client ID"),
        azureClientSecret: optionalTrimmedString(MAX_LEN_1000, "Client Secret"),
        azureSubscriptionId: optionalTrimmedString(MAX_LEN_200, "Subscription ID"),
        azureTenantId: optionalTrimmedString(MAX_LEN_200, "Tenant ID"),
        azureResourceGroupName: optionalTrimmedString(MAX_LEN_200, "Resource Group"),
        baiduCloudAccessKey: optionalTrimmedString(MAX_LEN_200, "Access Key"),
        baiduCloudSecretKey: optionalTrimmedString(MAX_LEN_1000, "Secret Key"),
        cloudflareAuthToken: optionalTrimmedString(MAX_LEN_1000, "Auth Token"),
        digitalOceanAuthToken: optionalTrimmedString(MAX_LEN_1000, "Auth Token"),
        gCloudServiceAccount: optionalTrimmedString(MAX_LEN_10000, "Service Account"),
        gCloudProjectId: optionalTrimmedString(MAX_LEN_200, "Project ID"),
        goDaddyApiKey: optionalTrimmedString(MAX_LEN_200, "API Key"),
        goDaddyApiSecret: optionalTrimmedString(MAX_LEN_1000, "API Secret"),
        hetznerApiToken: optionalTrimmedString(MAX_LEN_1000, "API Token"),
        huaweiCloudAccessKey: optionalTrimmedString(MAX_LEN_200, "Access Key"),
        huaweiCloudSecretKey: optionalTrimmedString(MAX_LEN_1000, "Secret Key"),
        huaweiCloudRegion: optionalTrimmedString(MAX_LEN_200, "Region"),
        namecheapApiUser: optionalTrimmedString(MAX_LEN_200, "API User"),
        namecheapApiKey: optionalTrimmedString(MAX_LEN_1000, "API Key"),
        rfc2136Nameserver: optionalTrimmedString(MAX_LEN_200, "Nameserver"),
        rfc2136TsigKeyName: optionalTrimmedString(MAX_LEN_200, "Tsig Key Name"),
        rfc2136TsigSecret: optionalTrimmedString(MAX_LEN_1000, "Tsig Secret"),
        rfc2136TsigAlgorithm: optionalTrimmedString(MAX_LEN_200, "Tsig Algorithm"),
        route53AccessKeyId: optionalTrimmedString(MAX_LEN_200, "Access Key ID"),
        route53SecretAccessKey: optionalTrimmedString(MAX_LEN_1000, "Secret Access Key"),
        route53HostedZoneId: optionalTrimmedString(MAX_LEN_200, "Hosted Zone ID"),
        route53Region: optionalTrimmedString(MAX_LEN_200, "Region"),
        tencentCloudSecretId: optionalTrimmedString(MAX_LEN_200, "Secret ID"),
        tencentCloudSecretKey: optionalTrimmedString(MAX_LEN_1000, "Secret Key"),
        tencentCloudRegion: optionalTrimmedString(MAX_LEN_200, "Region"),
        availableInProjects: z.boolean(),
        default: z.boolean(),
    })
    .superRefine((value, ctx) => {
        switch (value.kind) {
            case EAcmeDnsProviderKind.AcmeDNS:
                requireField(ctx, value.acmeDnsApiBase, "acmeDnsApiBase", "API Base");
                break;
            case EAcmeDnsProviderKind.Azure:
                requireField(ctx, value.azureClientId, "azureClientId", "Client ID");
                requireField(ctx, value.azureClientSecret, "azureClientSecret", "Client Secret");
                break;
            case EAcmeDnsProviderKind.BaiduCloud:
                requireField(ctx, value.baiduCloudAccessKey, "baiduCloudAccessKey", "Access Key");
                requireField(ctx, value.baiduCloudSecretKey, "baiduCloudSecretKey", "Secret Key");
                break;
            case EAcmeDnsProviderKind.Cloudflare:
                requireField(ctx, value.cloudflareAuthToken, "cloudflareAuthToken", "Auth Token");
                break;
            case EAcmeDnsProviderKind.DigitalOcean:
                requireField(ctx, value.digitalOceanAuthToken, "digitalOceanAuthToken", "Auth Token");
                break;
            case EAcmeDnsProviderKind.GCloud:
                requireField(ctx, value.gCloudServiceAccount, "gCloudServiceAccount", "Service Account");
                break;
            case EAcmeDnsProviderKind.GoDaddy:
                requireField(ctx, value.goDaddyApiKey, "goDaddyApiKey", "API Key");
                requireField(ctx, value.goDaddyApiSecret, "goDaddyApiSecret", "API Secret");
                break;
            case EAcmeDnsProviderKind.Hetzner:
                requireField(ctx, value.hetznerApiToken, "hetznerApiToken", "API Token");
                break;
            case EAcmeDnsProviderKind.HuaweiCloud:
                requireField(ctx, value.huaweiCloudAccessKey, "huaweiCloudAccessKey", "Access Key");
                requireField(ctx, value.huaweiCloudSecretKey, "huaweiCloudSecretKey", "Secret Key");
                break;
            case EAcmeDnsProviderKind.Namecheap:
                requireField(ctx, value.namecheapApiUser, "namecheapApiUser", "API User");
                requireField(ctx, value.namecheapApiKey, "namecheapApiKey", "API Key");
                break;
            case EAcmeDnsProviderKind.RFC2136:
                requireField(ctx, value.rfc2136Nameserver, "rfc2136Nameserver", "Nameserver");
                requireField(ctx, value.rfc2136TsigKeyName, "rfc2136TsigKeyName", "Tsig Key Name");
                requireField(ctx, value.rfc2136TsigSecret, "rfc2136TsigSecret", "Tsig Secret");
                break;
            case EAcmeDnsProviderKind.Route53:
                requireField(ctx, value.route53AccessKeyId, "route53AccessKeyId", "Access Key ID");
                requireField(ctx, value.route53SecretAccessKey, "route53SecretAccessKey", "Secret Access Key");
                break;
            case EAcmeDnsProviderKind.TencentCloud:
                requireField(ctx, value.tencentCloudSecretId, "tencentCloudSecretId", "Secret ID");
                requireField(ctx, value.tencentCloudSecretKey, "tencentCloudSecretKey", "Secret Key");
                break;
        }
    });

export const TestAcmeDnsProviderAccessFormSchema = z.object({
    testDomain: z.string().trim().min(1, "Domain is required").max(255, "Domain must be at most 255 characters"),
});

export type CreateOrEditAcmeDnsProviderFormOutput = z.output<typeof CreateOrEditAcmeDnsProviderFormSchema>;
export type CreateOrEditAcmeDnsProviderFormInput = CreateOrEditAcmeDnsProviderFormOutput;
export type TestAcmeDnsProviderAccessFormInput = z.input<typeof TestAcmeDnsProviderAccessFormSchema>;
export type TestAcmeDnsProviderAccessFormOutput = z.output<typeof TestAcmeDnsProviderAccessFormSchema>;

export const DEFAULT_ACME_DNS_PROVIDER_FORM_VALUES: CreateOrEditAcmeDnsProviderFormInput = {
    name: "",
    kind: EAcmeDnsProviderKind.AcmeDNS,
    acmeDnsApiBase: "",
    acmeDnsAllowList: "",
    acmeDnsStoragePath: "",
    acmeDnsStorageBaseUrl: "",
    azureClientId: "",
    azureClientSecret: "",
    azureSubscriptionId: "",
    azureTenantId: "",
    azureResourceGroupName: "",
    baiduCloudAccessKey: "",
    baiduCloudSecretKey: "",
    cloudflareAuthToken: "",
    digitalOceanAuthToken: "",
    gCloudServiceAccount: "",
    gCloudProjectId: "",
    goDaddyApiKey: "",
    goDaddyApiSecret: "",
    hetznerApiToken: "",
    huaweiCloudAccessKey: "",
    huaweiCloudSecretKey: "",
    huaweiCloudRegion: "",
    namecheapApiUser: "",
    namecheapApiKey: "",
    rfc2136Nameserver: "",
    rfc2136TsigKeyName: "",
    rfc2136TsigSecret: "",
    rfc2136TsigAlgorithm: "",
    route53AccessKeyId: "",
    route53SecretAccessKey: "",
    route53HostedZoneId: "",
    route53Region: "",
    tencentCloudSecretId: "",
    tencentCloudSecretKey: "",
    tencentCloudRegion: "",
    availableInProjects: false,
    default: false,
};
