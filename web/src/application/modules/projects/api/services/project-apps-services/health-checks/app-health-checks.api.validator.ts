import type { AxiosResponse } from "axios";
import { z } from "zod";
import type {
    AppHealthChecks_CreateOne_Res,
    AppHealthChecks_FindManyPaginated_Res,
    AppHealthChecks_FindOneById_Res,
} from "~/projects/api/services/project-apps-services";
import {
    EAppHealthCheckGrpcStatus,
    EAppHealthCheckGrpcVersion,
    EAppHealthCheckRestMethod,
    EAppHealthCheckType,
} from "~/projects/module-shared/enums";

import { ESettingStatus, ESettingType } from "@application/shared/enums";

import { BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

const NamedRefSchema = z
    .object({
        id: z.string(),
        name: z.string(),
    })
    .nullish()
    .transform(value => value ?? undefined);

const RestReturnTextSchema = z.object({
    exact: z.string().optional().default(""),
    regex: z.string().optional().default(""),
});

const RestReturnJSONSchema = z.object({
    exact: z.string().optional().default(""),
    contain: z.string().optional().default(""),
});

const RestSchema = z.object({
    url: z.string(),
    method: z.nativeEnum(EAppHealthCheckRestMethod),
    contentType: z.string().optional().default(""),
    body: z.string().optional().default(""),
    returnCode: z.string().optional().default(""),
    returnText: RestReturnTextSchema.nullish().transform(value => value ?? null),
    returnJSON: RestReturnJSONSchema.nullish().transform(value => value ?? null),
});

const GrpcSchema = z.object({
    version: z.nativeEnum(EAppHealthCheckGrpcVersion),
    addr: z.string(),
    service: z.string().optional().default(""),
    returnStatus: z.nativeEnum(EAppHealthCheckGrpcStatus),
});

const NotificationSchema = z
    .object({
        successUseDefault: z.boolean(),
        success: NamedRefSchema,
        failureUseDefault: z.boolean(),
        failure: NamedRefSchema,
        minSendInterval: z.string(),
    })
    .nullish()
    .transform(value => value ?? null);

const AppHealthCheckSchema = z.object({
    id: z.string(),
    type: z.nativeEnum(ESettingType),
    name: z.string(),
    kind: z.string().optional(),
    status: z.nativeEnum(ESettingStatus),
    inherited: z.boolean().optional().default(false),
    availableInProjects: z.boolean().optional().default(false),
    default: z.boolean().optional().default(false),
    updateVer: z.number(),
    createdAt: z.coerce.date(),
    updatedAt: z.coerce.date(),
    expireAt: z.coerce.date().nullable().optional().default(null),
    healthcheckType: z.nativeEnum(EAppHealthCheckType),
    interval: z.string(),
    maxRetry: z.number(),
    retryDelay: z.string(),
    timeout: z.string(),
    saveResultTasks: z.boolean(),
    rest: RestSchema.nullish().transform(value => value ?? null),
    grpc: GrpcSchema.nullish().transform(value => value ?? null),
    notification: NotificationSchema,
});

const FindManyPaginatedSchema = z.object({
    data: z.array(AppHealthCheckSchema),
    meta: PagingMetaApiSchema,
});

const FindOneByIdSchema = z.object({
    data: AppHealthCheckSchema,
    meta: BaseMetaApiSchema.nullable(),
});

const CreateOneSchema = z.object({
    data: z.object({
        id: z.string(),
    }),
    meta: BaseMetaApiSchema.nullable(),
});

export class AppHealthChecksApiValidator {
    findManyPaginated = (response: AxiosResponse): AppHealthChecks_FindManyPaginated_Res => {
        return parseApiResponse({
            response,
            schema: FindManyPaginatedSchema,
        });
    };

    findOneById = (response: AxiosResponse): AppHealthChecks_FindOneById_Res => {
        return parseApiResponse({
            response,
            schema: FindOneByIdSchema,
        });
    };

    createOne = (response: AxiosResponse): AppHealthChecks_CreateOne_Res => {
        return parseApiResponse({
            response,
            schema: CreateOneSchema,
        });
    };
}
