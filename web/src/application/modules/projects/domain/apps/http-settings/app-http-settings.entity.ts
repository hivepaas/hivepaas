import type { EHttpPathMode, ELBStrategy } from "~/projects/module-shared/enums";
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

export type AppHttpLBConfig = {
    strategy: ELBStrategy;
};

export type AppHttpBasicAuthConfig = {
    id: string;
    name: string;
    enabled: boolean;
};

export type AppHttpDomain = {
    enabled: boolean;
    domain: string;
    domainRedirect?: string;
    sslCert?: { id: string; name: string } | null;
    containerPort: number;
    forceHttps?: boolean;
    basicAuth?: AppHttpBasicAuthConfig | null;
    lbConfig?: AppHttpLBConfig | null;
    clientConfig?: AppHttpClientConfig | null;
    headerConfig?: AppHttpHeaderConfig | null;
    compressionConfig?: AppHttpCompressionConfig | null;
    rateLimitConfig?: AppHttpRateLimitConfig | null;
    pathRewriteConfig?: AppHttpPathRewriteConfig | null;
    circuitBreakerConfig?: AppHttpCircuitBreakerConfig | null;
    paths?: AppHttpPathConfig[];
};

export type AppHttpClientConfig = {
    enabled: boolean;
    maxRequestBody: string;
    memRequestBody: string;
    allowedIPs: string[];
};

export type AppHttpHeaderConfig = {
    enabled: boolean;
    autoContentType: boolean;
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

export type AppHttpPathRewriteConfig = {
    enabled: boolean;
    prefixAdd: string;
    prefixStrip: string;
    prefixStripIsRegex: boolean;
    pathReplace: string;
    pathReplaceIsRegex: boolean;
    pathReplaceWith: string;
};

export type AppHttpCircuitBreakerConfig = {
    enabled: boolean;
    expression: string;
    checkPeriod: string;
    fallbackDuration: string;
    recoveryDuration: string;
    responseCode: number;
};

export type AppHttpPathConfig = {
    enabled: boolean;
    path: string;
    mode: EHttpPathMode;
    basicAuth?: AppHttpBasicAuthConfig | null;
    clientConfig?: AppHttpClientConfig | null;
    headerConfig?: AppHttpHeaderConfig | null;
    compressionConfig?: AppHttpCompressionConfig | null;
    rateLimitConfig?: AppHttpRateLimitConfig | null;
    pathRewriteConfig?: AppHttpPathRewriteConfig | null;
    circuitBreakerConfig?: AppHttpCircuitBreakerConfig | null;
};

export type AppHttpSettingsObjectIdReq = {
    id: string;
};

export type AppHttpBasicAuthConfigReq = AppHttpSettingsObjectIdReq & {
    enabled: boolean;
};

export type AppHttpSettingsUpdateDomain = {
    enabled: boolean;
    domain: string;
    domainRedirect: string;
    sslCert: AppHttpSettingsObjectIdReq;
    containerPort: number;
    forceHttps: boolean;
    basicAuth: AppHttpBasicAuthConfigReq;
    lbConfig?: AppHttpLBConfig | null;
    clientConfig?: AppHttpClientConfig | null;
    headerConfig?: AppHttpHeaderConfig | null;
    compressionConfig?: AppHttpCompressionConfig | null;
    rateLimitConfig?: AppHttpRateLimitConfig | null;
    pathRewriteConfig?: AppHttpPathRewriteConfig | null;
    circuitBreakerConfig?: AppHttpCircuitBreakerConfig | null;
    paths: AppHttpSettingsUpdatePath[] | null;
};

export type AppHttpSettingsUpdatePath = {
    enabled: boolean;
    path: string;
    mode: EHttpPathMode;
    basicAuth: AppHttpBasicAuthConfigReq;
    clientConfig?: AppHttpClientConfig | null;
    headerConfig?: AppHttpHeaderConfig | null;
    compressionConfig?: AppHttpCompressionConfig | null;
    rateLimitConfig?: AppHttpRateLimitConfig | null;
    pathRewriteConfig?: AppHttpPathRewriteConfig | null;
    circuitBreakerConfig?: AppHttpCircuitBreakerConfig | null;
};

export type AppHttpSettingsUpdatePayload = {
    exposePublicly: boolean;
    domains: AppHttpSettingsUpdateDomain[];
    updateVer: number;
};
