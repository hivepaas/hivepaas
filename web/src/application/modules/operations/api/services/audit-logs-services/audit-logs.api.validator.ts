import { type AxiosResponse } from "axios";
import { z } from "zod";
import type { AuditLogResult, AuditLogSource, AuditLogType } from "~/operations/domain";

import { BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    AuditLogs_FindManyPaginated_Res,
    AuditLogs_FindOneById_Res,
    AuditLogs_FindTypes_Res,
} from "./audit-logs.api.contracts";

const AuditLogResultSchema = z.string().transform(val => val as AuditLogResult);
const AuditLogTypeSchema = z.string().transform(val => val as AuditLogType);
const AuditLogSourceSchema = z
    .string()
    .optional()
    .transform(val => (val ? (val as AuditLogSource) : undefined));

const DateSchema = z.preprocess(value => {
    if (value instanceof Date) {
        return Number.isNaN(value.getTime()) ? new Date() : value;
    }
    if (typeof value === "string" || typeof value === "number") {
        const date = new Date(value);
        return Number.isNaN(date.getTime()) ? new Date() : date;
    }
    return new Date();
}, z.date());

const ActorSchema = z
    .object({
        id: z.string(),
        type: z.string(),
        name: z.string().optional(),
        loggedName: z.string().optional(),
        photo: z.string().optional(),
    })
    .nullish()
    .transform(val => val ?? undefined);

const ResourceSchema = z
    .object({
        id: z.string(),
        type: z.string(),
        name: z.string().optional(),
        loggedName: z.string().optional(),
        photo: z.string().optional(),
    })
    .nullish()
    .transform(val => val ?? undefined);

const ScopeProjectSchema = z
    .object({
        id: z.string(),
        name: z.string(),
        key: z.string().optional(),
        photo: z
            .string()
            .nullish()
            .transform(val => val ?? undefined),
        status: z.string().optional(),
    })
    .nullish()
    .transform(val => val ?? undefined);

const ScopeProjectEnvSchema = z
    .object({
        id: z.string(),
        name: z.string(),
        color: z.string().optional(),
        status: z.string().optional(),
    })
    .nullish()
    .transform(val => val ?? undefined);

const ScopeAppSchema = z
    .object({
        id: z.string(),
        name: z.string(),
        key: z.string().optional(),
        photo: z
            .string()
            .nullish()
            .transform(val => val ?? undefined),
        status: z.string().optional(),
        env: z.string().optional(),
    })
    .nullish()
    .transform(val => val ?? undefined);

const ScopeUserSchema = z
    .object({
        id: z.string(),
        username: z.string(),
        email: z.string().optional(),
        fullName: z.string().optional(),
        photo: z
            .string()
            .nullish()
            .transform(val => val ?? undefined),
        role: z.string().optional(),
    })
    .nullish()
    .transform(val => val ?? undefined);

const AuditLogSchema = z.object({
    id: z.string(),
    type: AuditLogTypeSchema,
    source: AuditLogSourceSchema,
    section: z
        .string()
        .nullish()
        .transform(val => (val?.trim() ? val.trim() : undefined)),
    result: AuditLogResultSchema,
    scopeProject: ScopeProjectSchema,
    scopeProjectEnv: ScopeProjectEnvSchema,
    scopeApp: ScopeAppSchema,
    scopeUser: ScopeUserSchema,
    actor: ActorSchema,
    resource: ResourceSchema,
    viaApiKey: z.boolean().optional(),
    sessionUid: z.string().optional(),
    clientIp: z.string().optional(),
    remoteAddr: z.string().optional(),
    userAgent: z.string().optional(),
    requestId: z.string().optional(),
    detail: z.string().optional(),
    hasDetail: z.boolean().optional(),
    createdAt: DateSchema,
});

const FindManyPaginatedSchema = z.object({
    data: z.array(AuditLogSchema).default([]),
    meta: PagingMetaApiSchema,
});

const FindOneByIdSchema = z.object({
    data: AuditLogSchema,
    meta: BaseMetaApiSchema.nullish(),
});

const FindTypesSchema = z.object({
    data: z.array(z.string()).default([]),
    meta: BaseMetaApiSchema.nullish(),
});

export class AuditLogsApiValidator {
    findManyPaginated = (response: AxiosResponse): AuditLogs_FindManyPaginated_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindManyPaginatedSchema });
        return { data, meta };
    };

    findOneById = (response: AxiosResponse): AuditLogs_FindOneById_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneByIdSchema });
        return { data, meta };
    };

    findTypes = (response: AxiosResponse): AuditLogs_FindTypes_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindTypesSchema });
        return { data, meta };
    };
}
