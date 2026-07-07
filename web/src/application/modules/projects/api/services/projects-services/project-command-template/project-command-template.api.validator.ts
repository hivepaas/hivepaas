import type { AxiosResponse } from "axios";
import { z } from "zod";
import { PROJECT_COMMAND_TEMPLATE_DEFAULT_CONSOLE_SIZE } from "~/projects/domain";
import { EAppScheduledJobArgSeparator } from "~/projects/module-shared/enums";

import { ECommandTemplateKind, ESettingStatus, ESettingType } from "@application/shared/enums";

import { BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    ProjectCommandTemplate_CreateOne_Res,
    ProjectCommandTemplate_DeleteOne_Res,
    ProjectCommandTemplate_FindManyPaginated_Res,
    ProjectCommandTemplate_FindOneById_Res,
    ProjectCommandTemplate_UpdateOne_Res,
    ProjectCommandTemplate_UpdateStatus_Res,
} from "./project-command-template.api.contracts";

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

const EnvVarSchema = z.object({
    key: z.string().optional().default(""),
    value: z.string().optional().default(""),
    isLiteral: z.boolean().optional().default(false),
});

const ArgSchema = z.object({
    use: z.boolean().optional().default(false),
    name: z.string().optional().default(""),
    value: z.string().optional().default(""),
});

const ArgGroupSchema = z.object({
    enabled: z.boolean().optional().default(false),
    exportEnv: z.string().optional().default(""),
    separator: z
        .string()
        .optional()
        .default(EAppScheduledJobArgSeparator.Whitespace)
        .transform(value =>
            value === EAppScheduledJobArgSeparator.Equal
                ? EAppScheduledJobArgSeparator.Equal
                : EAppScheduledJobArgSeparator.Whitespace,
        ),
    args: z
        .array(ArgSchema)
        .nullish()
        .transform(value => value ?? []),
});

const ConsoleSizeSchema = z
    .object({
        width: z.number(),
        height: z.number(),
    })
    .nullish()
    .transform(value => value ?? { ...PROJECT_COMMAND_TEMPLATE_DEFAULT_CONSOLE_SIZE });

const CommandTemplateSchema = z.object({
    id: z.string(),
    type: z.literal(ESettingType.CommandTemplate),
    name: z
        .string()
        .nullish()
        .transform(value => value ?? ""),
    kind: z.nativeEnum(ECommandTemplateKind),
    status: z.nativeEnum(ESettingStatus),
    inherited: z.boolean().optional().default(false),
    availableInProjects: z.boolean().optional().default(false),
    default: z.boolean().optional().default(false),
    updateVer: z.number(),
    createdAt: z.coerce.date(),
    updatedAt: NullableDateSchema,
    expireAt: NullableDateSchema,
    command: z.string().optional().default(""),
    script: z.string().optional().default(""),
    workingDir: z.string().optional().default(""),
    envVars: z
        .array(EnvVarSchema)
        .nullish()
        .transform(value => value ?? []),
    argGroups: z
        .array(ArgGroupSchema)
        .nullish()
        .transform(value => value ?? []),
    consoleSize: ConsoleSizeSchema,
    tty: z.boolean().optional().default(false),
    size: z.number().optional().default(0),
});

const FindManyPaginatedSchema = z.object({
    data: z.array(CommandTemplateSchema),
    meta: PagingMetaApiSchema,
});

const FindOneByIdSchema = z.object({
    data: CommandTemplateSchema,
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

export class ProjectCommandTemplateApiValidator {
    findManyPaginated = (response: AxiosResponse): ProjectCommandTemplate_FindManyPaginated_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManyPaginatedSchema,
        });

        return { data, meta };
    };

    findOneById = (response: AxiosResponse): ProjectCommandTemplate_FindOneById_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindOneByIdSchema,
        });

        return { data, meta };
    };

    createOne = (response: AxiosResponse): ProjectCommandTemplate_CreateOne_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: CreateOneSchema,
        });

        return { data, meta };
    };

    updateOne = (response: AxiosResponse): ProjectCommandTemplate_UpdateOne_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };

    updateStatus = (response: AxiosResponse): ProjectCommandTemplate_UpdateStatus_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };

    deleteOne = (response: AxiosResponse): ProjectCommandTemplate_DeleteOne_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };
}
