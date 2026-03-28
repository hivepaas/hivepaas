import { z } from "zod";
import { EAppArmorMode, ERestartPolicyCondition, ESeccompMode } from "~/projects/module-shared/enums";

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
    tty: z.boolean(),
    openStdin: z.boolean(),
    readOnly: z.boolean(),
    stopSignal: z.string(),
    stopGracePeriod: z.string(),
});

/** Aligned with `RestartPolicy`; `maxAttempts` as string for text inputs. */
export const ContainerSettingsFormRestartPolicySchema = z.object({
    condition: z.nativeEnum(ERestartPolicyCondition),
    delay: z.string(),
    window: z.string(),
    maxAttempts: z.string(),
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

export const AppConfigContainerSettingsFormSchema = z.object({
    general: ContainerSettingsFormGeneralSchema,
    labels: z.array(ContainerSettingsFormLabelRowSchema),
    restartPolicy: ContainerSettingsFormRestartPolicySchema,
    privileges: ContainerSettingsFormPrivilegesSchema,
});

export type ContainerSettingsFormLabelRow = z.infer<typeof ContainerSettingsFormLabelRowSchema>;
export type ContainerSettingsFormGeneral = z.infer<typeof ContainerSettingsFormGeneralSchema>;
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
        tty: false,
        openStdin: false,
        readOnly: false,
        stopSignal: "",
        stopGracePeriod: "",
    },
    labels: [],
    restartPolicy: {
        condition: ERestartPolicyCondition.None,
        delay: "",
        window: "",
        maxAttempts: "",
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
