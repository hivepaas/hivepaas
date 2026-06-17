import type { EAcmeDnsProviderKind } from "@application/shared/enums";

import type { SettingsBaseEntity } from "./settings.base.entity";

export interface SettingAcmeDnsProviderAcmeDNS {
    apiBase: string;
    allowList: string[];
    storagePath: string;
    storageBaseUrl: string;
}

export interface SettingAcmeDnsProviderAzure {
    clientId: string;
    clientSecret: string;
    subscriptionId: string;
    tenantId: string;
    resourceGroupName: string;
}

export interface SettingAcmeDnsProviderBaiduCloud {
    accessKey: string;
    secretKey: string;
}

export interface SettingAcmeDnsProviderCloudflare {
    authToken: string;
}

export interface SettingAcmeDnsProviderDigitalOcean {
    authToken: string;
}

export interface SettingAcmeDnsProviderGCloud {
    projectId: string;
    serviceAccount: string;
}

export interface SettingAcmeDnsProviderGoDaddy {
    apiKey: string;
    apiSecret: string;
}

export interface SettingAcmeDnsProviderHetzner {
    apiToken: string;
}

export interface SettingAcmeDnsProviderHuaweiCloud {
    accessKey: string;
    secretKey: string;
    region: string;
}

export interface SettingAcmeDnsProviderNamecheap {
    apiUser: string;
    apiKey: string;
}

export interface SettingAcmeDnsProviderRFC2136 {
    nameserver: string;
    tsigKeyName: string;
    tsigSecret: string;
    tsigAlgorithm: string;
}

export interface SettingAcmeDnsProviderRoute53 {
    accessKeyId: string;
    secretAccessKey: string;
    hostedZoneId: string;
    region: string;
}

export interface SettingAcmeDnsProviderTencentCloud {
    secretId: string;
    secretKey: string;
    region: string;
}

export interface SettingAcmeDnsProvider extends SettingsBaseEntity {
    kind: EAcmeDnsProviderKind;
    acmeDns?: SettingAcmeDnsProviderAcmeDNS | null;
    azure?: SettingAcmeDnsProviderAzure | null;
    baiduCloud?: SettingAcmeDnsProviderBaiduCloud | null;
    cloudflare?: SettingAcmeDnsProviderCloudflare | null;
    digitalOcean?: SettingAcmeDnsProviderDigitalOcean | null;
    gCloud?: SettingAcmeDnsProviderGCloud | null;
    goDaddy?: SettingAcmeDnsProviderGoDaddy | null;
    hetzner?: SettingAcmeDnsProviderHetzner | null;
    huaweiCloud?: SettingAcmeDnsProviderHuaweiCloud | null;
    namecheap?: SettingAcmeDnsProviderNamecheap | null;
    rfc2136?: SettingAcmeDnsProviderRFC2136 | null;
    route53?: SettingAcmeDnsProviderRoute53 | null;
    tencentCloud?: SettingAcmeDnsProviderTencentCloud | null;
    secretMasked?: boolean;
    inherited?: boolean;
}
