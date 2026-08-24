import type { EHttpPathMode, ELBStrategy, ERoutingProtocol } from "~/projects/module-shared/enums";
import type { SettingsBaseEntity } from "~/settings/domain";

export type AppRoutingSettings = {
    domainSuggestion: string;
    port: number;
    exposePublicly: boolean;
    domains: AppRoutingDomain[];
    updateVer: number;
};

export type AppRoutingSslCert = SettingsBaseEntity & {
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

export type AppRoutingLBConfig = {
    strategy: ELBStrategy;
};

export type AppRoutingBasicAuthConfig = {
    id: string;
    name: string;
    enabled: boolean;
};

export type AppRoutingDomain = {
    enabled: boolean;
    domain: string;
    protocol?: ERoutingProtocol;
    domainRedirect?: string;
    sslCert?: { id: string; name: string } | null;
    containerPort: number;
    tlsPassthrough?: boolean;
    forceHttps?: boolean;
    basicAuth?: AppRoutingBasicAuthConfig | null;
    lbConfig?: AppRoutingLBConfig | null;
    clientConfig?: AppRoutingClientConfig | null;
    headerConfig?: AppRoutingHeaderConfig | null;
    compressionConfig?: AppRoutingCompressionConfig | null;
    rateLimitConfig?: AppRoutingRateLimitConfig | null;
    pathRewriteConfig?: AppRoutingPathRewriteConfig | null;
    circuitBreakerConfig?: AppRoutingCircuitBreakerConfig | null;
    websocketConfig?: AppRoutingWebsocketConfig | null;
    paths?: AppRoutingPathConfig[];
};

export type AppRoutingWebsocketConfig = {
    enabled: boolean;
};

export type AppRoutingClientConfig = {
    enabled: boolean;
    maxRequestBody: string;
    memRequestBody: string;
    allowedIPs: string[];
};

export type AppRoutingHeaderConfig = {
    enabled: boolean;
    autoContentType: boolean;
    toAddToRequests: Record<string, string>;
    toRemoveFromRequests: string[];
    toAddToResponses: Record<string, string>;
    toRemoveFromResponses: string[];
};

export type AppRoutingCompressionConfig = {
    enabled: boolean;
    excludedContentTypes: string[];
    includedContentTypes: string[];
    minResponseBody: string;
    defaultEncoding: string;
};

export type AppRoutingRateLimitConfig = {
    enabled: boolean;
    average: number;
    period: string;
    burst: number;
    maxInFlightReq: number;
};

export type AppRoutingPathRewriteConfig = {
    enabled: boolean;
    prefixAdd: string;
    prefixStrip: string;
    prefixStripIsRegex: boolean;
    pathReplace: string;
    pathReplaceIsRegex: boolean;
    pathReplaceWith: string;
};

export type AppRoutingCircuitBreakerConfig = {
    enabled: boolean;
    expression: string;
    checkPeriod: string;
    fallbackDuration: string;
    recoveryDuration: string;
    responseCode: number;
};

export type AppRoutingPathConfig = {
    enabled: boolean;
    path: string;
    mode: EHttpPathMode;
    basicAuth?: AppRoutingBasicAuthConfig | null;
    clientConfig?: AppRoutingClientConfig | null;
    headerConfig?: AppRoutingHeaderConfig | null;
    compressionConfig?: AppRoutingCompressionConfig | null;
    rateLimitConfig?: AppRoutingRateLimitConfig | null;
    pathRewriteConfig?: AppRoutingPathRewriteConfig | null;
    circuitBreakerConfig?: AppRoutingCircuitBreakerConfig | null;
    websocketConfig?: AppRoutingWebsocketConfig | null;
};

export type AppRoutingSettingsObjectIdReq = {
    id: string;
};

export type AppRoutingBasicAuthConfigReq = AppRoutingSettingsObjectIdReq & {
    enabled: boolean;
};

export type AppRoutingSettingsUpdateDomain = {
    enabled: boolean;
    domain: string;
    protocol: ERoutingProtocol;
    domainRedirect: string;
    sslCert: AppRoutingSettingsObjectIdReq;
    containerPort: number;
    tlsPassthrough: boolean;
    forceHttps: boolean;
    basicAuth: AppRoutingBasicAuthConfigReq;
    lbConfig?: AppRoutingLBConfig | null;
    clientConfig?: AppRoutingClientConfig | null;
    headerConfig?: AppRoutingHeaderConfig | null;
    compressionConfig?: AppRoutingCompressionConfig | null;
    rateLimitConfig?: AppRoutingRateLimitConfig | null;
    pathRewriteConfig?: AppRoutingPathRewriteConfig | null;
    circuitBreakerConfig?: AppRoutingCircuitBreakerConfig | null;
    websocketConfig?: AppRoutingWebsocketConfig | null;
    paths: AppRoutingSettingsUpdatePath[] | null;
};

export type AppRoutingSettingsUpdatePath = {
    enabled: boolean;
    path: string;
    mode: EHttpPathMode;
    basicAuth: AppRoutingBasicAuthConfigReq;
    clientConfig?: AppRoutingClientConfig | null;
    headerConfig?: AppRoutingHeaderConfig | null;
    compressionConfig?: AppRoutingCompressionConfig | null;
    rateLimitConfig?: AppRoutingRateLimitConfig | null;
    pathRewriteConfig?: AppRoutingPathRewriteConfig | null;
    circuitBreakerConfig?: AppRoutingCircuitBreakerConfig | null;
    websocketConfig?: AppRoutingWebsocketConfig | null;
};

export type AppRoutingSettingsUpdatePayload = {
    port: number;
    exposePublicly: boolean;
    domains: AppRoutingSettingsUpdateDomain[];
    updateVer: number;
};

// Aliases for backwards-compatibility if needed
export type AppHttpSettings = AppRoutingSettings;
export type AppHttpDomain = AppRoutingDomain;
export type AppHttpBasicAuthConfig = AppRoutingBasicAuthConfig;
export type AppHttpLBConfig = AppRoutingLBConfig;
export type AppHttpClientConfig = AppRoutingClientConfig;
export type AppHttpHeaderConfig = AppRoutingHeaderConfig;
export type AppHttpCompressionConfig = AppRoutingCompressionConfig;
export type AppHttpRateLimitConfig = AppRoutingRateLimitConfig;
export type AppHttpPathRewriteConfig = AppRoutingPathRewriteConfig;
export type AppHttpCircuitBreakerConfig = AppRoutingCircuitBreakerConfig;
export type AppHttpWebsocketConfig = AppRoutingWebsocketConfig;
export type AppHttpPathConfig = AppRoutingPathConfig;
export type AppHttpSettingsUpdatePayload = AppRoutingSettingsUpdatePayload;
