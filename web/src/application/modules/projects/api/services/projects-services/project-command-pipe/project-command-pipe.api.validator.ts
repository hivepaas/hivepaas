import type { AxiosResponse } from "axios";
import { z } from "zod";

import { ESettingStatus, ESettingType } from "@application/shared/enums";

import { BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    ProjectCommandPipe_CreateOne_Res,
    ProjectCommandPipe_DeleteOne_Res,
    ProjectCommandPipe_FindManyPaginated_Res,
    ProjectCommandPipe_FindOneById_Res,
    ProjectCommandPipe_UpdateOne_Res,
    ProjectCommandPipe_UpdateStatus_Res,
} from "./project-command-pipe.api.contracts";

const NullableDateSchema = z.preprocess(value => {
    if (value === null || value === undefined || value === "") {
        return null;
    }

    if (value instanceof Date) {
        return Number.isNaN(value.getTime()) ? null : value;
    }

    if (typeof value !== "string" && typeof value !== "number") {
        return null;
    }

    const date = new Date(value);

    return Number.isNaN(date.getTime()) ? null : date;
}, z.date().nullable());

const SettingRefSchema = z
    .object({
        id: z.string(),
        type: z.string().optional().default(""),
        name: z
            .string()
            .nullish()
            .transform(value => value ?? ""),
        kind: z.string().optional().default(""),
        status: z.nativeEnum(ESettingStatus),
        inherited: z.boolean().optional().default(false),
        availableInProjects: z.boolean().optional().default(false),
        default: z.boolean().optional().default(false),
        updateVer: z.number(),
        createdAt: z.coerce.date(),
        updatedAt: NullableDateSchema,
        expireAt: NullableDateSchema,
    })
    .nullish()
    .transform(value => value ?? null);

const CommandPipeSchema = z.object({
    id: z.string(),
    type: z.literal(ESettingType.CommandPipe),
    name: z
        .string()
        .nullish()
        .transform(value => value ?? ""),
    status: z.nativeEnum(ESettingStatus),
    inherited: z.boolean().optional().default(false),
    availableInProjects: z.boolean().optional().default(false),
    default: z.boolean().optional().default(false),
    updateVer: z.number(),
    createdAt: z.coerce.date(),
    updatedAt: NullableDateSchema,
    expireAt: NullableDateSchema,
    size: z.number().optional().default(0),
    sourceCommand: SettingRefSchema,
    targetCommand: SettingRefSchema,
});

const FindManyPaginatedSchema = z.object({
    data: z.array(CommandPipeSchema),
    meta: PagingMetaApiSchema,
});

const FindOneByIdSchema = z.object({
    data: CommandPipeSchema,
    meta: BaseMetaApiSchema.nullish(),
});

const CreateOneSchema = z.object({
    data: z.object({
        id: z.string(),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class ProjectCommandPipeApiValidator {
    findManyPaginated = (response: AxiosResponse): ProjectCommandPipe_FindManyPaginated_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManyPaginatedSchema,
        });

        return { data, meta };
    };

    findOneById = (response: AxiosResponse): ProjectCommandPipe_FindOneById_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindOneByIdSchema,
        });

        return { data, meta };
    };

    createOne = (response: AxiosResponse): ProjectCommandPipe_CreateOne_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: CreateOneSchema,
        });

        return { data, meta };
    };

    updateOne = (response: AxiosResponse): ProjectCommandPipe_UpdateOne_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };

    updateStatus = (response: AxiosResponse): ProjectCommandPipe_UpdateStatus_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };

    deleteOne = (response: AxiosResponse): ProjectCommandPipe_DeleteOne_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };
}
