import { type AxiosResponse } from "axios";
import { z } from "zod";
import { EAppDeploymentMethod, EBuildTool, ERepoType } from "~/projects/module-shared/enums";

import { SettingsBaseEntitySchema } from "@application/modules/settings/module-shared/schemas";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import {
    type AppDeploymentSettings_FindOne_Res,
    type AppDeploymentSettings_UpdateOne_Res,
} from "./app-deployment-settings.api.contracts";

const BaseDeploymentSettingsSchema = z.object({
    command: z.string().optional(),
    workingDir: z.string().optional(),
    preDeploymentCommand: z.string().optional(),
    postDeploymentCommand: z.string().optional(),
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
    updateVer: z.number(),
});

const RepoMethodSchema = BaseDeploymentSettingsSchema.extend({
    activeMethod: z.literal(EAppDeploymentMethod.Repo),
    repoSource: z.object({
        buildTool: z.nativeEnum(EBuildTool),
        repoType: z.nativeEnum(ERepoType),
        repoUrl: z.string(),
        repoRef: z.string(),
        credentials: SettingsBaseEntitySchema,
        dockerfilePath: z.string(),
        imageName: z.string(),
        imageTags: z.string(),
        pushToRegistry: SettingsBaseEntitySchema,
    }),
});

const ImageMethodSchema = BaseDeploymentSettingsSchema.extend({
    activeMethod: z.literal(EAppDeploymentMethod.Image),
    imageSource: z.object({
        image: z.string(),
        registryAuth: SettingsBaseEntitySchema,
    }),
});

const AppDeploymentSettingsSchema = z.discriminatedUnion("activeMethod", [RepoMethodSchema, ImageMethodSchema]);

const FindOneSchema = z.object({
    data: AppDeploymentSettingsSchema,
    meta: BaseMetaApiSchema.nullable(),
});

const UpdateOneSchema = z.object({
    data: z.object({ type: z.literal("success") }),
    meta: BaseMetaApiSchema.nullable(),
});

export class AppDeploymentSettingsApiValidator {
    findOne = (response: AxiosResponse): AppDeploymentSettings_FindOne_Res => {
        return parseApiResponse({ response, schema: FindOneSchema });
    };

    updateOne = (response: AxiosResponse): AppDeploymentSettings_UpdateOne_Res => {
        return parseApiResponse({ response, schema: UpdateOneSchema });
    };
}
