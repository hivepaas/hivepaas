import { z } from "zod";
import { EProjectAppStatus } from "~/projects/module-shared/enums";

const CopyProjectAppSslCertRefSchema = z.object({
    id: z.string(),
    name: z.string(),
});

const CopyProjectAppToggleSchema = z.object({
    copy: z.boolean(),
});

const CopyProjectAppDomainSchema = z.object({
    sourceDomain: z.string(),
    targetDomain: z.string().trim(),
    sourceSslCert: CopyProjectAppSslCertRefSchema.nullable(),
    targetSslCert: CopyProjectAppSslCertRefSchema.nullable(),
});

const CopyProjectAppTargetStatusSchema = z.union([
    z.literal(EProjectAppStatus.Active),
    z.literal(EProjectAppStatus.Disabled),
]);

const CopyProjectAppFormSchemaBase = z.object({
    sourceName: z.string(),
    targetName: z.string().trim().min(1, "Name is required"),
    sourceEnv: z.string(),
    targetEnv: z.string().trim().max(50, "Environment must be at most 50 characters"),
    sourceStatus: z.nativeEnum(EProjectAppStatus),
    targetStatus: CopyProjectAppTargetStatusSchema,
    copyConfigFiles: CopyProjectAppToggleSchema,
    copyDeploymentSettings: CopyProjectAppToggleSchema,
    copyEnvVars: CopyProjectAppToggleSchema,
    copyHealthChecks: CopyProjectAppToggleSchema,
    copyHttpSettings: z.object({
        copy: z.boolean(),
        copyDomainSettings: z.array(CopyProjectAppDomainSchema),
    }),
    copySchedJobs: CopyProjectAppToggleSchema,
    copySecrets: CopyProjectAppToggleSchema,
    updateVer: z.number(),
});

export function createCopyProjectAppFormSchema(
    envNames: string[] = [],
): z.ZodEffects<typeof CopyProjectAppFormSchemaBase> {
    return CopyProjectAppFormSchemaBase.superRefine((values, ctx) => {
        if (values.targetEnv !== "" && !envNames.includes(values.targetEnv)) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["targetEnv"],
                message: "Environment must be one of the project environments",
            });
        }
    });
}

export const CopyProjectAppFormSchema = createCopyProjectAppFormSchema();

export type CopyProjectAppFormInput = z.input<ReturnType<typeof createCopyProjectAppFormSchema>>;
export type CopyProjectAppFormOutput = z.output<ReturnType<typeof createCopyProjectAppFormSchema>>;
