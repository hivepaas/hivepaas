import { type AppHttpDomain, type AppHttpSettings, type AppHttpSettingsUpdatePayload } from "~/projects/domain";
import { type EHttpPathMode, ELBStrategy } from "~/projects/module-shared/enums";

import {
    type AppConfigHttpSettingsFormSchemaInput,
    type AppConfigHttpSettingsFormSchemaOutput,
    createDefaultLBConfig,
    emptyDomain,
} from "../schemas";

function isLBStrategy(value: string): value is (typeof ELBStrategy)[keyof typeof ELBStrategy] {
    return Object.values(ELBStrategy).some(item => item === value);
}

function mapDomainToFormInput(domain: AppHttpDomain): AppConfigHttpSettingsFormSchemaInput["domains"][number] {
    return {
        enabled: domain.enabled,
        domain: domain.domain,
        containerPort: domain.containerPort,
        domainRedirect: domain.domainRedirect ?? "",
        sslCert: domain.sslCert?.id ? { id: domain.sslCert.id, name: domain.sslCert.name } : undefined,
        forceHttps: domain.forceHttps ?? false,
        basicAuth: domain.basicAuth?.id
            ? { id: domain.basicAuth.id, name: domain.basicAuth.name, enabled: domain.basicAuth.enabled }
            : undefined,
        lbConfig: domain.lbConfig ? { strategy: domain.lbConfig.strategy } : createDefaultLBConfig(),
        clientConfig: domain.clientConfig
            ? {
                  enabled: domain.clientConfig.enabled,
                  maxRequestBody: domain.clientConfig.maxRequestBody,
                  memRequestBody: domain.clientConfig.memRequestBody,
                  allowedIPs: domain.clientConfig.allowedIPs.join(","),
              }
            : undefined,
        headerConfig: domain.headerConfig
            ? {
                  enabled: domain.headerConfig.enabled,
                  autoContentType: domain.headerConfig.autoContentType,
                  toAddToRequests: Object.entries(domain.headerConfig.toAddToRequests).map(([key, value]) => ({
                      key,
                      value,
                  })),
                  toRemoveFromRequests: domain.headerConfig.toRemoveFromRequests.map(value => ({ value })),
                  toAddToResponses: Object.entries(domain.headerConfig.toAddToResponses).map(([key, value]) => ({
                      key,
                      value,
                  })),
                  toRemoveFromResponses: domain.headerConfig.toRemoveFromResponses.map(value => ({ value })),
              }
            : undefined,
        compressionConfig: domain.compressionConfig
            ? {
                  enabled: domain.compressionConfig.enabled,
                  excludedContentTypes: domain.compressionConfig.excludedContentTypes.join("\n"),
                  includedContentTypes: domain.compressionConfig.includedContentTypes.join("\n"),
                  minResponseBody: domain.compressionConfig.minResponseBody,
                  defaultEncoding: domain.compressionConfig.defaultEncoding,
              }
            : undefined,
        rateLimitConfig: domain.rateLimitConfig
            ? {
                  enabled: domain.rateLimitConfig.enabled,
                  average: domain.rateLimitConfig.average,
                  period: domain.rateLimitConfig.period,
                  burst: domain.rateLimitConfig.burst,
                  maxInFlightReq: domain.rateLimitConfig.maxInFlightReq,
              }
            : undefined,
        pathRewriteConfig: domain.pathRewriteConfig
            ? {
                  enabled: domain.pathRewriteConfig.enabled,
                  prefixAdd: domain.pathRewriteConfig.prefixAdd,
                  prefixStrip: domain.pathRewriteConfig.prefixStrip,
                  prefixStripIsRegex: domain.pathRewriteConfig.prefixStripIsRegex,
                  pathReplace: domain.pathRewriteConfig.pathReplace,
                  pathReplaceIsRegex: domain.pathRewriteConfig.pathReplaceIsRegex,
                  pathReplaceWith: domain.pathRewriteConfig.pathReplaceWith,
              }
            : undefined,
        circuitBreakerConfig: domain.circuitBreakerConfig
            ? {
                  enabled: domain.circuitBreakerConfig.enabled,
                  expression: domain.circuitBreakerConfig.expression,
                  checkPeriod: domain.circuitBreakerConfig.checkPeriod,
                  fallbackDuration: domain.circuitBreakerConfig.fallbackDuration,
                  recoveryDuration: domain.circuitBreakerConfig.recoveryDuration,
                  responseCode: domain.circuitBreakerConfig.responseCode,
              }
            : undefined,
        paths: (domain.paths ?? []).map(path => ({
            enabled: path.enabled,
            path: path.path,
            mode: path.mode,
            basicAuth: path.basicAuth?.id
                ? { id: path.basicAuth.id, name: path.basicAuth.name, enabled: path.basicAuth.enabled }
                : undefined,
            clientConfig: path.clientConfig
                ? {
                      enabled: path.clientConfig.enabled,
                      maxRequestBody: path.clientConfig.maxRequestBody,
                      memRequestBody: path.clientConfig.memRequestBody,
                      allowedIPs: path.clientConfig.allowedIPs.join(","),
                  }
                : undefined,
            headerConfig: path.headerConfig
                ? {
                      enabled: path.headerConfig.enabled,
                      autoContentType: path.headerConfig.autoContentType,
                      toAddToRequests: Object.entries(path.headerConfig.toAddToRequests).map(([key, value]) => ({
                          key,
                          value,
                      })),
                      toRemoveFromRequests: path.headerConfig.toRemoveFromRequests.map(value => ({ value })),
                      toAddToResponses: Object.entries(path.headerConfig.toAddToResponses).map(([key, value]) => ({
                          key,
                          value,
                      })),
                      toRemoveFromResponses: path.headerConfig.toRemoveFromResponses.map(value => ({ value })),
                  }
                : undefined,
            compressionConfig: path.compressionConfig
                ? {
                      enabled: path.compressionConfig.enabled,
                      excludedContentTypes: path.compressionConfig.excludedContentTypes.join("\n"),
                      includedContentTypes: path.compressionConfig.includedContentTypes.join("\n"),
                      minResponseBody: path.compressionConfig.minResponseBody,
                      defaultEncoding: path.compressionConfig.defaultEncoding,
                  }
                : undefined,
            rateLimitConfig: path.rateLimitConfig
                ? {
                      enabled: path.rateLimitConfig.enabled,
                      average: path.rateLimitConfig.average,
                      period: path.rateLimitConfig.period,
                      burst: path.rateLimitConfig.burst,
                      maxInFlightReq: path.rateLimitConfig.maxInFlightReq,
                  }
                : undefined,
            pathRewriteConfig: path.pathRewriteConfig
                ? {
                      enabled: path.pathRewriteConfig.enabled,
                      prefixAdd: path.pathRewriteConfig.prefixAdd,
                      prefixStrip: path.pathRewriteConfig.prefixStrip,
                      prefixStripIsRegex: path.pathRewriteConfig.prefixStripIsRegex,
                      pathReplace: path.pathRewriteConfig.pathReplace,
                      pathReplaceIsRegex: path.pathRewriteConfig.pathReplaceIsRegex,
                      pathReplaceWith: path.pathRewriteConfig.pathReplaceWith,
                  }
                : undefined,
            circuitBreakerConfig: path.circuitBreakerConfig
                ? {
                      enabled: path.circuitBreakerConfig.enabled,
                      expression: path.circuitBreakerConfig.expression,
                      checkPeriod: path.circuitBreakerConfig.checkPeriod,
                      fallbackDuration: path.circuitBreakerConfig.fallbackDuration,
                      recoveryDuration: path.circuitBreakerConfig.recoveryDuration,
                      responseCode: path.circuitBreakerConfig.responseCode,
                  }
                : undefined,
        })),
    };
}

export function mapAppHttpSettingsToFormInput(data: AppHttpSettings): AppConfigHttpSettingsFormSchemaInput {
    return {
        port: data.port >= 1 && data.port <= 65535 ? data.port : 8080,
        exposePublicly: data.exposePublicly,
        domains: data.domains.map(mapDomainToFormInput),
    };
}

export function mapFormValuesToPayload(values: AppConfigHttpSettingsFormSchemaOutput): AppHttpSettingsUpdatePayload {
    return {
        port: values.port,
        exposePublicly: values.exposePublicly,
        domains: values.domains.map(domain => ({
            enabled: domain.enabled,
            domain: domain.domain,
            containerPort: domain.containerPort,
            domainRedirect: domain.domainRedirect,
            sslCert: { id: domain.sslCert?.id ?? "" },
            forceHttps: domain.forceHttps,
            basicAuth: { id: domain.basicAuth?.id ?? "", enabled: domain.basicAuth?.enabled ?? false },
            lbConfig:
                domain.lbConfig && isLBStrategy(domain.lbConfig.strategy)
                    ? { strategy: domain.lbConfig.strategy }
                    : null,
            clientConfig: domain.clientConfig
                ? {
                      enabled: domain.clientConfig.enabled,
                      maxRequestBody: domain.clientConfig.maxRequestBody,
                      memRequestBody: domain.clientConfig.memRequestBody,
                      allowedIPs: domain.clientConfig.allowedIPs
                          .replace(/\n/g, ",")
                          .split(",")
                          .map(s => s.trim())
                          .filter(Boolean),
                  }
                : null,
            headerConfig: domain.headerConfig
                ? {
                      enabled: domain.headerConfig.enabled,
                      autoContentType: domain.headerConfig.autoContentType,
                      toAddToRequests: Object.fromEntries(
                          domain.headerConfig.toAddToRequests.map(({ key, value }) => [key, value]),
                      ),
                      toRemoveFromRequests: domain.headerConfig.toRemoveFromRequests
                          .map(item => item.value)
                          .filter(Boolean),
                      toAddToResponses: Object.fromEntries(
                          domain.headerConfig.toAddToResponses.map(({ key, value }) => [key, value]),
                      ),
                      toRemoveFromResponses: domain.headerConfig.toRemoveFromResponses
                          .map(item => item.value)
                          .filter(Boolean),
                  }
                : null,
            compressionConfig: domain.compressionConfig
                ? {
                      enabled: domain.compressionConfig.enabled,
                      excludedContentTypes: domain.compressionConfig.excludedContentTypes
                          .replace(/\n/g, ",")
                          .split(",")
                          .map(item => item.trim())
                          .filter(Boolean),
                      includedContentTypes: domain.compressionConfig.includedContentTypes
                          .replace(/\n/g, ",")
                          .split(",")
                          .map(item => item.trim())
                          .filter(Boolean),
                      minResponseBody: domain.compressionConfig.minResponseBody,
                      defaultEncoding: domain.compressionConfig.defaultEncoding,
                  }
                : null,
            rateLimitConfig: domain.rateLimitConfig
                ? {
                      enabled: domain.rateLimitConfig.enabled,
                      average: domain.rateLimitConfig.average,
                      period: domain.rateLimitConfig.period,
                      burst: domain.rateLimitConfig.burst,
                      maxInFlightReq: domain.rateLimitConfig.maxInFlightReq,
                  }
                : null,
            pathRewriteConfig: domain.pathRewriteConfig
                ? {
                      enabled: domain.pathRewriteConfig.enabled,
                      prefixAdd: domain.pathRewriteConfig.prefixAdd,
                      prefixStrip: domain.pathRewriteConfig.prefixStrip,
                      prefixStripIsRegex: domain.pathRewriteConfig.prefixStripIsRegex,
                      pathReplace: domain.pathRewriteConfig.pathReplace,
                      pathReplaceIsRegex: domain.pathRewriteConfig.pathReplaceIsRegex,
                      pathReplaceWith: domain.pathRewriteConfig.pathReplaceWith,
                  }
                : null,
            circuitBreakerConfig: domain.circuitBreakerConfig
                ? {
                      enabled: domain.circuitBreakerConfig.enabled,
                      expression: domain.circuitBreakerConfig.expression,
                      checkPeriod: domain.circuitBreakerConfig.checkPeriod,
                      fallbackDuration: domain.circuitBreakerConfig.fallbackDuration,
                      recoveryDuration: domain.circuitBreakerConfig.recoveryDuration,
                      responseCode: domain.circuitBreakerConfig.responseCode,
                  }
                : null,
            paths: domain.paths.map(path => ({
                enabled: path.enabled,
                path: path.path,
                mode: path.mode as EHttpPathMode,
                basicAuth: { id: path.basicAuth?.id ?? "", enabled: path.basicAuth?.enabled ?? false },
                clientConfig: path.clientConfig
                    ? {
                          enabled: path.clientConfig.enabled,
                          maxRequestBody: path.clientConfig.maxRequestBody,
                          memRequestBody: path.clientConfig.memRequestBody,
                          allowedIPs: path.clientConfig.allowedIPs
                              .replace(/\n/g, ",")
                              .split(",")
                              .map(s => s.trim())
                              .filter(Boolean),
                      }
                    : null,
                headerConfig: path.headerConfig
                    ? {
                          enabled: path.headerConfig.enabled,
                          autoContentType: path.headerConfig.autoContentType,
                          toAddToRequests: Object.fromEntries(
                              path.headerConfig.toAddToRequests.map(({ key, value }) => [key, value]),
                          ),
                          toRemoveFromRequests: path.headerConfig.toRemoveFromRequests
                              .map(item => item.value)
                              .filter(Boolean),
                          toAddToResponses: Object.fromEntries(
                              path.headerConfig.toAddToResponses.map(({ key, value }) => [key, value]),
                          ),
                          toRemoveFromResponses: path.headerConfig.toRemoveFromResponses
                              .map(item => item.value)
                              .filter(Boolean),
                      }
                    : null,
                compressionConfig: path.compressionConfig
                    ? {
                          enabled: path.compressionConfig.enabled,
                          excludedContentTypes: path.compressionConfig.excludedContentTypes
                              .replace(/\n/g, ",")
                              .split(",")
                              .map(item => item.trim())
                              .filter(Boolean),
                          includedContentTypes: path.compressionConfig.includedContentTypes
                              .replace(/\n/g, ",")
                              .split(",")
                              .map(item => item.trim())
                              .filter(Boolean),
                          minResponseBody: path.compressionConfig.minResponseBody,
                          defaultEncoding: path.compressionConfig.defaultEncoding,
                      }
                    : null,
                rateLimitConfig: path.rateLimitConfig
                    ? {
                          enabled: path.rateLimitConfig.enabled,
                          average: path.rateLimitConfig.average,
                          period: path.rateLimitConfig.period,
                          burst: path.rateLimitConfig.burst,
                          maxInFlightReq: path.rateLimitConfig.maxInFlightReq,
                      }
                    : null,
                pathRewriteConfig: path.pathRewriteConfig
                    ? {
                          enabled: path.pathRewriteConfig.enabled,
                          prefixAdd: path.pathRewriteConfig.prefixAdd,
                          prefixStrip: path.pathRewriteConfig.prefixStrip,
                          prefixStripIsRegex: path.pathRewriteConfig.prefixStripIsRegex,
                          pathReplace: path.pathRewriteConfig.pathReplace,
                          pathReplaceIsRegex: path.pathRewriteConfig.pathReplaceIsRegex,
                          pathReplaceWith: path.pathRewriteConfig.pathReplaceWith,
                      }
                    : null,
                circuitBreakerConfig: path.circuitBreakerConfig
                    ? {
                          enabled: path.circuitBreakerConfig.enabled,
                          expression: path.circuitBreakerConfig.expression,
                          checkPeriod: path.circuitBreakerConfig.checkPeriod,
                          fallbackDuration: path.circuitBreakerConfig.fallbackDuration,
                          recoveryDuration: path.circuitBreakerConfig.recoveryDuration,
                          responseCode: path.circuitBreakerConfig.responseCode,
                      }
                    : null,
            })),
        })),
        updateVer: 0,
    };
}

export { emptyDomain };
