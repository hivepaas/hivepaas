import type { AxiosResponse } from "axios";
import { z } from "zod";
import type { AppDeployments_FindManyPaginated_Res } from "~/projects/api/services/project-apps-services";
import {
    EAppDeploymentMethod,
    EAppDeploymentStatus,
    EAppDeploymentTriggerSource,
} from "~/projects/module-shared/enums";

import { EUserRole } from "@application/shared/enums";

import { PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

const SourceUserSchema = z
    .object({
        id: z.string(),
        username: z.string(),
        email: z.string(),
        fullName: z.string(),
        photo: z
            .string()
            .nullish()
            .transform(value => value ?? null),
        role: z.nativeEnum(EUserRole),
    })
    .nullish()
    .transform(value => value ?? null);

const DeploymentTriggerSchema = z
    .object({
        source: z.nativeEnum(EAppDeploymentTriggerSource),
        sourceUser: SourceUserSchema,
    })
    .nullish()
    .transform(value => value ?? null);

const DeploymentOutputSchema = z
    .object({
        commitHash: z.string().optional().default(""),
        commitHashShort: z.string().optional().default(""),
        commitURL: z.string().optional().default(""),
        commitTitle: z.string().optional().default(""),
        commitMessage: z.string().optional().default(""),
        commitAuthor: z.string().optional().default(""),
        imageTags: z
            .array(z.string())
            .nullish()
            .transform(value => value ?? []),
    })
    .nullish()
    .transform(value => value ?? null);

const DeploymentSettingsSchema = z
    .object({
        activeMethod: z.nativeEnum(EAppDeploymentMethod),
    })
    .passthrough();

const NullableDateSchema = z.preprocess(
    value => (value === null || value === undefined || value === "" ? null : value),
    z.coerce.date().nullable(),
);

const AppDeploymentSchema = z.object({
    id: z.string(),
    status: z.nativeEnum(EAppDeploymentStatus),
    updateVer: z.number(),
    settings: DeploymentSettingsSchema,
    trigger: DeploymentTriggerSchema,
    output: DeploymentOutputSchema,
    startedAt: NullableDateSchema,
    endedAt: NullableDateSchema,
    createdAt: z.coerce.date(),
    updatedAt: NullableDateSchema,
});

const FindManyPaginatedSchema = z.object({
    data: z.array(AppDeploymentSchema),
    meta: PagingMetaApiSchema,
});

export class AppDeploymentsApiValidator {
    findManyPaginated = (response: AxiosResponse): AppDeployments_FindManyPaginated_Res => {
        return parseApiResponse({
            response,
            schema: FindManyPaginatedSchema,
        });
    };
}
