import type { AxiosResponse } from "axios";
import { z } from "zod";
import { EProjectAppStatus } from "~/projects/module-shared/enums";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type { AppCloneSettings_FindOne_Res } from "./app-clone-settings.api.contracts";

const SettingRefSchema = z
    .object({
        id: z.string(),
        name: z
            .string()
            .nullish()
            .transform(value => value ?? ""),
    })
    .nullish()
    .transform(value => value ?? null);

const RoutingDomainSchema = z.object({
    sourceDomain: z.string(),
    targetDomain: z.string(),
    sourceSslCert: SettingRefSchema,
    targetSslCert: SettingRefSchema,
});

const NotificationSchema = z
    .object({
        successUseDefault: z.boolean().optional().default(true),
        success: SettingRefSchema,
        failureUseDefault: z.boolean().optional().default(true),
        failure: SettingRefSchema,
    })
    .nullish()
    .transform(value => value ?? null);

const AppCloneSettingsSchema = z.object({
    targetName: z.string().optional().default(""),
    targetEnv: z.string().optional().default(""),
    targetStatus: z.nativeEnum(EProjectAppStatus).optional().default(EProjectAppStatus.Active),
    targetReplicas: z.number().optional().default(-1),
    cloneDeploymentSettings: z.boolean().optional().default(true),
    cloneRoutingSettings: z.boolean().optional().default(true),
    cloneRoutingDomains: z
        .array(RoutingDomainSchema)
        .nullish()
        .transform(value => value ?? []),

    cloneVolumes: z.boolean().optional().default(true),
    cloneVolumeData: z.boolean().optional().default(true),
    liveVolumeClone: z.boolean().optional().default(true),
    includedVolumes: z
        .array(z.string())
        .nullish()
        .transform(value => value ?? []),
    excludedVolumes: z
        .array(z.string())
        .nullish()
        .transform(value => value ?? []),
    cloneEnvVars: z.boolean().optional().default(true),
    cloneSecrets: z.boolean().optional().default(true),
    cloneConfigFiles: z.boolean().optional().default(true),
    clonePeriodicJobs: z.boolean().optional().default(true),
    cloneSchedJobs: z.boolean().optional().default(true),

    commandPipes: z
        .array(
            z.object({
                id: z.string(),
                name: z
                    .string()
                    .nullish()
                    .transform(value => value ?? ""),
            }),
        )
        .nullish()
        .transform(value => value ?? []),
    notification: NotificationSchema,
    updateVer: z.number(),
});

const FindOneSchema = z.object({
    data: AppCloneSettingsSchema,
    meta: BaseMetaApiSchema.nullable(),
});

export class AppCloneSettingsApiValidator {
    findOne = (response: AxiosResponse): AppCloneSettings_FindOne_Res => {
        return parseApiResponse({ response, schema: FindOneSchema });
    };
}
