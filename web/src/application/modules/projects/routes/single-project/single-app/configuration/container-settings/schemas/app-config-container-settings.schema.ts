import { z } from "zod";
import { EAppArmorMode, EHealthcheckMode, ERestartPolicyCondition, ESeccompMode } from "~/projects/module-shared/enums";

/** Key/value rows before mapping to API `Record<string, string>` */
export const ContainerSettingsFormLabelRowSchema = z.object({
    key: z.string(),
    value: z.string(),
});

/** General/runtime fields aligned with `ContainerSpec` (groups as one string in the form). */
export const ContainerSettingsFormGeneralSchema = z.object({
    image: z.string(),
    command: z.string(),
    workingDir: z.string(),
    hostname: z.string(),
    user: z.string(),
    /** Space-separated; mapped to `string[]` on submit */
    groups: z.string(),
    init: z.boolean(),
    tty: z.boolean(),
    openStdin: z.boolean(),
    readOnly: z.boolean(),
    stopSignal: z.string(),
    stopGracePeriod: z.string(),
});

/** Aligned with `RestartPolicy`; duration fields stay as strings for text inputs. */
export const ContainerSettingsFormRestartPolicySchema = z.object({
    condition: z.nativeEnum(ERestartPolicyCondition),
    delay: z.string(),
    window: z.string(),
    maxAttempts: z.number().optional(),
});

/** Aligned with `Healthcheck`; duration fields stay as strings for text inputs. */
export const ContainerSettingsFormHealthcheckSchema = z.object({
    enabled: z.boolean(),
    mode: z.nativeEnum(EHealthcheckMode),
    command: z.string(),
    interval: z.string(),
    timeout: z.string(),
    startPeriod: z.string(),
    startInterval: z.string(),
    retries: z.number().optional(),
});

/** Aligned with Docker swarm log driver config. */
export const ContainerSettingsFormLogDriverSchema = z.object({
    driver: z.string().trim(),
    options: z.array(ContainerSettingsFormLabelRowSchema),
});

/** Aligned with `Privileges` plus form-only `selinuxEnabled` (maps to `seLinuxContext.disable`). */
export const ContainerSettingsFormPrivilegesSchema = z.object({
    noNewPrivileges: z.boolean(),
    selinuxEnabled: z.boolean(),
    seLinuxUser: z.string(),
    seLinuxRole: z.string(),
    seLinuxType: z.string(),
    seLinuxLevel: z.string(),
    seccompMode: z.nativeEnum(ESeccompMode),
    seccompProfile: z.string(),
    appArmorMode: z.nativeEnum(EAppArmorMode),
});

export const AppConfigContainerSettingsFormSchema = z
    .object({
        general: ContainerSettingsFormGeneralSchema,
        serviceLabels: z.array(ContainerSettingsFormLabelRowSchema),
        containerLabels: z.array(ContainerSettingsFormLabelRowSchema),
        healthcheck: ContainerSettingsFormHealthcheckSchema,
        restartPolicy: ContainerSettingsFormRestartPolicySchema,
        logDriver: ContainerSettingsFormLogDriverSchema,
        privileges: ContainerSettingsFormPrivilegesSchema,
    })
    .superRefine((values, ctx) => {
        const hasOptions = values.logDriver.options.some(row => row.key.trim() !== "");
        if (hasOptions && !values.logDriver.driver) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["logDriver", "driver"],
                message: "Driver is required when options are configured",
            });
        }
    });

export type ContainerSettingsFormLabelRow = z.infer<typeof ContainerSettingsFormLabelRowSchema>;
export type ContainerSettingsFormGeneral = z.infer<typeof ContainerSettingsFormGeneralSchema>;
export type ContainerSettingsFormHealthcheck = z.infer<typeof ContainerSettingsFormHealthcheckSchema>;
export type ContainerSettingsFormLogDriver = z.infer<typeof ContainerSettingsFormLogDriverSchema>;
export type ContainerSettingsFormRestartPolicy = z.infer<typeof ContainerSettingsFormRestartPolicySchema>;
export type ContainerSettingsFormPrivileges = z.infer<typeof ContainerSettingsFormPrivilegesSchema>;

export type AppConfigContainerSettingsFormSchemaInput = z.input<typeof AppConfigContainerSettingsFormSchema>;
export type AppConfigContainerSettingsFormSchemaOutput = z.output<typeof AppConfigContainerSettingsFormSchema>;

export const emptyAppConfigContainerSettingsFormDefaults: AppConfigContainerSettingsFormSchemaInput = {
    general: {
        image: "",
        command: "",
        workingDir: "",
        hostname: "",
        user: "",
        groups: "",
        init: false,
        tty: false,
        openStdin: false,
        readOnly: false,
        stopSignal: "",
        stopGracePeriod: "",
    },
    serviceLabels: [],
    containerLabels: [],
    healthcheck: {
        enabled: false,
        mode: EHealthcheckMode.Inherit,
        command: "",
        interval: "",
        timeout: "",
        startPeriod: "",
        startInterval: "",
        retries: undefined,
    },
    restartPolicy: {
        condition: ERestartPolicyCondition.None,
        delay: "",
        window: "",
        maxAttempts: undefined,
    },
    logDriver: {
        driver: "",
        options: [],
    },
    privileges: {
        noNewPrivileges: false,
        selinuxEnabled: true,
        seLinuxUser: "",
        seLinuxRole: "",
        seLinuxType: "",
        seLinuxLevel: "",
        seccompMode: ESeccompMode.Default,
        seccompProfile: "",
        appArmorMode: EAppArmorMode.Default,
    },
};
