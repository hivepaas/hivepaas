import { type AppHttpDomain, type AppHttpSettings, type AppHttpSettingsUpdatePayload } from "~/projects/domain";
import { type EHttpPathMode } from "~/projects/module-shared/enums";

import {
    type AppConfigHttpSettingsFormSchemaInput,
    type AppConfigHttpSettingsFormSchemaOutput,
    emptyDomain,
} from "../schemas";

function mapDomainToFormInput(domain: AppHttpDomain): AppConfigHttpSettingsFormSchemaInput["domains"][number] {
    return {
        enabled: domain.enabled,
        domain: domain.domain,
        containerPort: domain.containerPort,
        domainRedirect: domain.domainRedirect ?? "",
        sslCert: domain.sslCert?.id ? { id: domain.sslCert.id, name: domain.sslCert.name } : undefined,
        forceHttps: domain.forceHttps ?? false,
        basicAuth: domain.basicAuth?.id ? { id: domain.basicAuth.id, name: domain.basicAuth.name } : undefined,
        clientConfig: domain.clientConfig
            ? {
                  enabled: domain.clientConfig.enabled,
                  maxRequestBody: domain.clientConfig.maxRequestBody,
                  memRequestBody: domain.clientConfig.memRequestBody,
                  allowedIPs: domain.clientConfig.allowedIPs.map(value => ({ value })),
              }
            : undefined,
        headerConfig: domain.headerConfig
            ? {
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
                  excludedContentTypes: domain.compressionConfig.excludedContentTypes.map(value => ({ value })),
                  includedContentTypes: domain.compressionConfig.includedContentTypes.map(value => ({ value })),
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
        paths: (domain.paths ?? []).map(path => ({
            path: path.path,
            mode: path.mode,
            basicAuth: path.basicAuth?.id ? { id: path.basicAuth.id, name: path.basicAuth.name } : undefined,
            clientConfig: path.clientConfig
                ? {
                      enabled: path.clientConfig.enabled,
                      maxRequestBody: path.clientConfig.maxRequestBody,
                      memRequestBody: path.clientConfig.memRequestBody,
                      allowedIPs: path.clientConfig.allowedIPs.map(value => ({ value })),
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
        })),
    };
}

export function mapAppHttpSettingsToFormInput(data: AppHttpSettings): AppConfigHttpSettingsFormSchemaInput {
    return {
        exposePublicly: data.exposePublicly,
        domains: data.domains.map(mapDomainToFormInput),
    };
}

export function mapFormValuesToPayload(values: AppConfigHttpSettingsFormSchemaOutput): AppHttpSettingsUpdatePayload {
    return {
        exposePublicly: values.exposePublicly,
        domains: values.domains.map(domain => ({
            enabled: domain.enabled,
            domain: domain.domain,
            containerPort: domain.containerPort,
            domainRedirect: domain.domainRedirect,
            sslCert: { id: domain.sslCert?.id ?? "" },
            forceHttps: domain.forceHttps,
            basicAuth: { id: domain.basicAuth?.id ?? "" },
            clientConfig: domain.clientConfig
                ? {
                      enabled: domain.clientConfig.enabled,
                      maxRequestBody: domain.clientConfig.maxRequestBody,
                      memRequestBody: domain.clientConfig.memRequestBody,
                      allowedIPs: domain.clientConfig.allowedIPs.map(item => item.value).filter(Boolean),
                  }
                : null,
            headerConfig: domain.headerConfig
                ? {
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
                          .map(item => item.value)
                          .filter(Boolean),
                      includedContentTypes: domain.compressionConfig.includedContentTypes
                          .map(item => item.value)
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
            paths: domain.paths.map(path => ({
                path: path.path,
                mode: path.mode as EHttpPathMode,
                basicAuth: { id: path.basicAuth?.id ?? "" },
                clientConfig: path.clientConfig
                    ? {
                          enabled: path.clientConfig.enabled,
                          maxRequestBody: path.clientConfig.maxRequestBody,
                          memRequestBody: path.clientConfig.memRequestBody,
                          allowedIPs: path.clientConfig.allowedIPs.map(item => item.value).filter(Boolean),
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
            })),
        })),
        updateVer: 0,
    };
}

export { emptyDomain };
