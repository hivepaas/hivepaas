import { z } from "zod";
import { EProjectAppStatus } from "~/projects/module-shared/enums";

const AppCloneSettingRefSchema = z.object({
    id: z.string(),
    name: z.string(),
});

const AppCloneHttpDomainSchema = z.object({
    sourceDomain: z.string(),
    targetDomain: z.string(),
    sourceSslCert: AppCloneSettingRefSchema.nullable(),
    targetSslCert: AppCloneSettingRefSchema.nullable(),
});

const AppCloneNotificationSchema = z.object({
    successUseDefault: z.boolean(),
    success: AppCloneSettingRefSchema.nullable(),
    failureUseDefault: z.boolean(),
    failure: AppCloneSettingRefSchema.nullable(),
});

const AppCloneTargetStatusSchema = z.union([
    z.literal(EProjectAppStatus.Active),
    z.literal(EProjectAppStatus.Disabled),
]);

const AppCloneSettingsFormSchemaBase = z.object({
    targetName: z.string().trim().min(1, "Target name is required"),
    targetEnv: z.string().trim().min(1, "Target environment is required"),
    targetStatus: AppCloneTargetStatusSchema,
    targetReplicas: z.coerce.number(),
    cloneDeploymentSettings: z.boolean(),
    cloneHttpSettings: z.boolean(),
    cloneHttpDomains: z.array(AppCloneHttpDomainSchema),
    cloneVolumes: z.boolean(),
    cloneVolumeData: z.boolean(),
    stopSourceAppBeforeClone: z.boolean(),
    cloneEnvVars: z.boolean(),
    cloneSecrets: z.boolean(),
    cloneConfigFiles: z.boolean(),
    clonePeriodicJobs: z.boolean(),
    cloneSchedJobs: z.boolean(),
    postCloneCommandsEnabled: z.boolean(),
    commandPipes: z.array(AppCloneSettingRefSchema),
    includedVolumes: z.array(z.string()),
    excludedVolumes: z.array(z.string()),
    notification: AppCloneNotificationSchema,
    updateVer: z.number(),
});

export function createAppCloneSettingsFormSchema(
    envNames: string[] = [],
): z.ZodEffects<typeof AppCloneSettingsFormSchemaBase> {
    return AppCloneSettingsFormSchemaBase.superRefine((values, ctx) => {
        if (values.targetEnv !== "" && !envNames.includes(values.targetEnv)) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["targetEnv"],
                message: "Environment must be one of the project environments",
            });
        }
    });
}

export const AppCloneSettingsFormSchema = createAppCloneSettingsFormSchema();

export type AppCloneSettingsFormSchemaInput = z.input<ReturnType<typeof createAppCloneSettingsFormSchema>>;
export type AppCloneSettingsFormSchemaOutput = z.output<ReturnType<typeof createAppCloneSettingsFormSchema>>;

export const emptyAppCloneSettingsFormDefaults: AppCloneSettingsFormSchemaInput = {
    targetName: "",
    targetEnv: "",
    targetStatus: EProjectAppStatus.Active,
    targetReplicas: -1,
    cloneDeploymentSettings: true,
    cloneHttpSettings: true,
    cloneHttpDomains: [],
    cloneVolumes: true,
    cloneVolumeData: true,
    stopSourceAppBeforeClone: false,
    cloneEnvVars: true,
    cloneSecrets: true,
    cloneConfigFiles: true,
    clonePeriodicJobs: true,
    cloneSchedJobs: true,
    postCloneCommandsEnabled: false,
    commandPipes: [],
    includedVolumes: [],
    excludedVolumes: [],
    notification: {
        successUseDefault: true,
        success: null,
        failureUseDefault: true,
        failure: null,
    },
    updateVer: 0,
};
