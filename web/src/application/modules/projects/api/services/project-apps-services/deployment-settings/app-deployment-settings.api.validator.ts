import { type AxiosResponse } from "axios";
import { z } from "zod";
import { EAppDeploymentMethod, EBuildTool, ERepoType } from "~/projects/module-shared/enums";
import { SettingsBaseEntitySchema } from "~/settings/module-shared/schemas";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type { AppDeploymentSettings_FindOne_Res } from "./app-deployment-settings.api.contracts";

const OptionalStringSchema = z
    .string()
    .nullish()
    .transform(value => value ?? "");

const SettingsRefSchema = SettingsBaseEntitySchema.nullish().transform(value => value ?? null);

const ImageTagsSchema = z.union([z.string(), z.array(z.string()), z.null(), z.undefined()]).transform(value => {
    if (Array.isArray(value)) {
        return value.join(", ");
    }

    return value ?? "";
});

const BaseDeploymentSettingsSchema = z.object({
    command: OptionalStringSchema.optional(),
    workingDir: OptionalStringSchema.optional(),
    preDeploymentCommand: OptionalStringSchema.optional(),
    postDeploymentCommand: OptionalStringSchema.optional(),
    notification: z
        .object({
            successUseDefault: z.boolean(),
            success: z
                .object({
                    id: z.string(),
                    name: z.string(),
                })
                .nullish(),
            failureUseDefault: z.boolean(),
            failure: z
                .object({
                    id: z.string(),
                    name: z.string(),
                })
                .nullish(),
        })
        .optional(),
    updateVer: z.number().optional().default(0),
});

const RepoSourceSchema = z.preprocess(
    value => {
        if (!value || typeof value !== "object") {
            return {};
        }

        const input = value as Record<string, unknown>;

        return {
            ...input,
            repoUrl: input["repoUrl"] ?? input["repoURL"],
        };
    },
    z.object({
        buildTool: z.nativeEnum(EBuildTool).optional().default(EBuildTool.Docker),
        repoType: z.nativeEnum(ERepoType).optional().default(ERepoType.Git),
        repoUrl: OptionalStringSchema,
        repoRef: OptionalStringSchema,
        commitHash: OptionalStringSchema,
        repoOptions: z
            .object({
                gitSubmodulesEnabled: z.boolean().optional().default(true),
                gitLfsEnabled: z.boolean().optional().default(true),
            })
            .nullish()
            .transform(val => val ?? { gitSubmodulesEnabled: true, gitLfsEnabled: true }),
        credentials: SettingsRefSchema,
        dockerfilePath: OptionalStringSchema,
        imageName: OptionalStringSchema,
        imageTags: ImageTagsSchema,
        pushToRegistry: SettingsRefSchema,
    }),
);

const RepoMethodSchema = BaseDeploymentSettingsSchema.extend({
    activeMethod: z.literal(EAppDeploymentMethod.Repo),
    repoSource: RepoSourceSchema,
});

const ImageSourceSchema = z.preprocess(
    value => {
        if (!value || typeof value !== "object") {
            return {};
        }

        return value;
    },
    z.object({
        image: OptionalStringSchema,
        registryAuth: SettingsRefSchema,
    }),
);

const ImageMethodSchema = BaseDeploymentSettingsSchema.extend({
    activeMethod: z.literal(EAppDeploymentMethod.Image),
    imageSource: ImageSourceSchema,
});

const AppDeploymentSettingsSchema = z.preprocess(
    value => {
        if (!value || typeof value !== "object") {
            return {
                activeMethod: EAppDeploymentMethod.Image,
                imageSource: {},
                updateVer: 0,
            };
        }

        const input = value as Record<string, unknown>;

        if (
            input["activeMethod"] === EAppDeploymentMethod.Repo ||
            input["activeMethod"] === EAppDeploymentMethod.Image
        ) {
            return input;
        }

        return {
            ...input,
            activeMethod: EAppDeploymentMethod.Image,
            imageSource: input["imageSource"] ?? {},
            updateVer: input["updateVer"] ?? 0,
        };
    },
    z.discriminatedUnion("activeMethod", [RepoMethodSchema, ImageMethodSchema]),
);

const FindOneSchema = z.object({
    data: AppDeploymentSettingsSchema,
    meta: BaseMetaApiSchema.nullable(),
});

export class AppDeploymentSettingsApiValidator {
    findOne = (response: AxiosResponse): AppDeploymentSettings_FindOne_Res => {
        return parseApiResponse({ response, schema: FindOneSchema });
    };
}
