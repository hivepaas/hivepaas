import { z } from "zod";

export const REGISTRY_STORAGE_TYPES = ["volume", "s3"] as const;

export const HivePaaSRegistrySettingsFormSchema = z
    .object({
        enabled: z.boolean(),
        domain: z.string().trim(),
        storageType: z.enum(REGISTRY_STORAGE_TYPES),
        // The selected setting's id, because that is what a Select holds. On the
        // wire it is an object either way, and the form mappers convert at both
        // ends.
        volumeId: z.string(),
        cloudStorageId: z.string(),
        dashboardEnabled: z.boolean(),
        cleanupEnabled: z.boolean(),
        keepLast: z.number().int().min(1).max(1000),
        keepDays: z.number().int().min(1).max(3650),
        memoryLimit: z
            .string()
            .trim()
            .regex(/^$|^\d+(b|kb|mb|gb|tb)$/i, "Use a size such as 1gb or 512mb"),
    })
    .superRefine((value, ctx) => {
        if (!value.enabled) {
            // Nothing is provisioned while it is off, so nothing below has to be
            // answered yet: filling the form in over two sittings is not an error.
            return;
        }

        // Images are named after it and docker only speaks to a registry over
        // HTTPS at a name, so there is nothing to provision without one.
        if (!value.domain) {
            ctx.addIssue({ code: "custom", path: ["domain"], message: "A domain is required" });
        }
        if (value.storageType === "volume" && !value.volumeId) {
            ctx.addIssue({ code: "custom", path: ["volumeId"], message: "Choose a volume" });
        }
        if (value.storageType === "s3" && !value.cloudStorageId) {
            ctx.addIssue({ code: "custom", path: ["cloudStorageId"], message: "Choose a cloud storage" });
        }
    });

export type HivePaaSRegistrySettingsFormInput = z.input<typeof HivePaaSRegistrySettingsFormSchema>;
export type HivePaaSRegistrySettingsFormOutput = z.output<typeof HivePaaSRegistrySettingsFormSchema>;
