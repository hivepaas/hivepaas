import type { EHttpPathMode } from "~/projects/module-shared/enums";
import type { SettingsBaseEntity } from "~/settings/domain";

export type AppHttpSettings = {
    internalEndpoints: string[];
    domainSuggestion: string;
    exposePublicly: boolean;
    domains: AppHttpDomain[];
    updateVer: number;
};

export type AppHttpSslCert = SettingsBaseEntity & {
    certType?: string;
    domain?: string;
    certificate?: string;
    privateKey?: string;
    keyType?: string;
    validPeriod?: number;
    email?: string;
    autoRenew?: boolean;
    renewableFrom?: string | null;
    notifyFrom?: string | null;
    secretMasked?: boolean;
    notification?: unknown;
};

export type AppHttpDomain = {
    enabled: boolean;
    domain: string;
    domainRedirect?: string;
    sslCert?: { id: string; name: string } | null;
    containerPort: number;
    forceHttps?: boolean;
    basicAuth?: { id: string; name: string } | null;
    clientConfig?: AppHttpClientConfig | null;
    headerConfig?: AppHttpHeaderConfig | null;
    compressionConfig?: AppHttpCompressionConfig | null;
    rateLimitConfig?: AppHttpRateLimitConfig | null;
    paths?: AppHttpPathConfig[];
};

export type AppHttpClientConfig = {
    enabled: boolean;
    maxRequestBody: string;
    memRequestBody: string;
    allowedIPs: string[];
};

export type AppHttpHeaderConfig = {
    toAddToRequests: Record<string, string>;
    toRemoveFromRequests: string[];
    toAddToResponses: Record<string, string>;
    toRemoveFromResponses: string[];
};

export type AppHttpCompressionConfig = {
    enabled: boolean;
    excludedContentTypes: string[];
    includedContentTypes: string[];
    minResponseBody: string;
    defaultEncoding: string;
};

export type AppHttpRateLimitConfig = {
    enabled: boolean;
    average: number;
    period: string;
    burst: number;
    maxInFlightReq: number;
};

export type AppHttpPathConfig = {
    path: string;
    mode: EHttpPathMode;
    basicAuth?: { id: string; name: string } | null;
    clientConfig?: AppHttpClientConfig | null;
    rateLimitConfig?: AppHttpRateLimitConfig | null;
};

export type AppHttpSettingsObjectIdReq = {
    id: string;
};

export type AppHttpSettingsUpdateDomain = {
    enabled: boolean;
    domain: string;
    domainRedirect: string;
    sslCert: AppHttpSettingsObjectIdReq;
    containerPort: number;
    forceHttps: boolean;
    basicAuth: AppHttpSettingsObjectIdReq;
    clientConfig?: AppHttpClientConfig | null;
    headerConfig?: AppHttpHeaderConfig | null;
    compressionConfig?: AppHttpCompressionConfig | null;
    rateLimitConfig?: AppHttpRateLimitConfig | null;
    paths: AppHttpSettingsUpdatePath[] | null;
};

export type AppHttpSettingsUpdatePath = {
    path: string;
    mode: EHttpPathMode;
    basicAuth: AppHttpSettingsObjectIdReq;
    clientConfig?: AppHttpClientConfig | null;
    rateLimitConfig?: AppHttpRateLimitConfig | null;
};

export type AppHttpSettingsUpdatePayload = {
    exposePublicly: boolean;
    domains: AppHttpSettingsUpdateDomain[];
    updateVer: number;
};
