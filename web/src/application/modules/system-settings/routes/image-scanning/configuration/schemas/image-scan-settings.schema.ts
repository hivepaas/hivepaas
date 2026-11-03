import { z } from "zod";

/** The daily scan's intervals: no more often than the DB is rebuilt, no less than weekly. */
export const IMAGE_SCAN_INTERVALS = ["6h", "12h", "1d", "2d", "7d"] as const;

/** An OCI repository Trivy fetches a DB from: registry.example.com:5000/aquasec/trivy-db:2. */
const REPOSITORY_PATTERN =
    /^$|^[a-zA-Z0-9][a-zA-Z0-9.-]*(:[0-9]+)?(\/[a-z0-9][a-z0-9._-]*)+(:[A-Za-z0-9_][A-Za-z0-9._-]*)?$/;

const SettingsRefSchema = z.object({
    id: z.string(),
    name: z.string(),
});

export const ImageScanSettingsFormSchema = z
    .object({
        enabled: z.boolean(),
        // The selected volume's id, because that is what a Combobox holds; on the
        // wire it is `volume`, an object.
        volumeId: z.string(),
        cpuLimit: z.number().min(0.25).max(256).nullable(),
        memoryLimit: z
            .string()
            .trim()
            .regex(/^$|^\d+(b|kb|mb|gb|tb)$/i, "Use a size such as 512mb or 1gb"),
        dbRepository: z.string().trim().regex(REPOSITORY_PATTERN, "Use an OCI repository: registry/path[:tag]"),
        javaDbRepository: z.string().trim().regex(REPOSITORY_PATTERN, "Use an OCI repository: registry/path[:tag]"),
        scheduleInterval: z.string().min(1, "Required"),
        scheduleFrom: z.date().nullable(),
        notification: z.object({
            successUseDefault: z.boolean(),
            success: SettingsRefSchema.nullish(),
            failureUseDefault: z.boolean(),
            failure: SettingsRefSchema.nullish(),
        }),
    })
    .superRefine((v, ctx) => {
        if (v.enabled && !v.volumeId) {
            ctx.addIssue({ code: "custom", path: ["volumeId"], message: "Required" });
        }
    });

export type ImageScanSettingsFormInput = z.input<typeof ImageScanSettingsFormSchema>;
export type ImageScanSettingsFormOutput = z.output<typeof ImageScanSettingsFormSchema>;
