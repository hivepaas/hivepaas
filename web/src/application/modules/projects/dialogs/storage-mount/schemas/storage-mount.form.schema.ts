import { z } from "zod";
import { EMountConsistency, EMountPropagation, EMountType } from "~/projects/module-shared/enums";

const VolumeDriverSchema = z.object({
    name: z.string().optional(),
    options: z.record(z.string()).optional(),
});

/** Form holds `{ key, value }[]`; domain/API use `Record<string, string>` via form mappers. */
const LabelsSchema = z
    .array(
        z.object({
            key: z.string(),
            value: z.string(),
        }),
    )
    .optional();

const BindMountSchema = z.object({
    type: z.literal(EMountType.Bind),
    target: z.string().min(1, "Target is required"),
    readOnly: z.boolean().optional(),
    consistency: z.nativeEnum(EMountConsistency).optional(),
    bindOptions: z.object({
        baseDir: z.string().min(1, "Base directory is required"),
        subpath: z.string().optional(),
        subpathRequired: z.string().optional(),
        propagation: z.nativeEnum(EMountPropagation).optional(),
        nonRecursive: z.boolean().optional(),
        createMountpoint: z.boolean().optional(),
        readOnlyNonRecursive: z.boolean().optional(),
        readOnlyForceRecursive: z.boolean().optional(),
    }),
});

const VolumeMountSchema = z.object({
    type: z.literal(EMountType.Volume),
    target: z.string().min(1, "Target is required"),
    readOnly: z.boolean().optional(),
    consistency: z.nativeEnum(EMountConsistency).optional(),
    volumeOptions: z.object({
        volume: z.string().min(1, "Volume is required"),
        subpath: z.string().optional(),
        subpathRequired: z.string().optional(),
        noCopy: z.boolean().optional(),
        labels: LabelsSchema,
        driverConfig: VolumeDriverSchema.optional(),
    }),
});

const ClusterMountSchema = z.object({
    type: z.literal(EMountType.Cluster),
    target: z.string().min(1, "Target is required"),
    readOnly: z.boolean().optional(),
    consistency: z.nativeEnum(EMountConsistency).optional(),
    clusterOptions: z.object({
        volume: z.string().min(1, "Volume is required"),
        subpath: z.string().optional(),
        subpathRequired: z.string().optional(),
        noCopy: z.boolean().optional(),
        labels: LabelsSchema,
        driverConfig: VolumeDriverSchema.optional(),
    }),
});

const TmpfsMountSchema = z.object({
    type: z.literal(EMountType.Tmpfs),
    target: z.string().min(1, "Target is required"),
    readOnly: z.boolean().optional(),
    consistency: z.nativeEnum(EMountConsistency).optional(),
    tmpfsOptions: z.object({
        size: z.string().min(1, "Size is required"),
        mode: z.string().optional(),
        options: z.array(z.array(z.string())).nullish(),
    }),
});

const BaseStorageMountFormSchema = z.discriminatedUnion("type", [
    BindMountSchema,
    VolumeMountSchema,
    ClusterMountSchema,
    TmpfsMountSchema,
]);

export const StorageMountFormSchema = BaseStorageMountFormSchema.superRefine((val, ctx) => {
    if (val.type === EMountType.Bind) {
        const req = val.bindOptions.subpathRequired ?? "";
        const sub = val.bindOptions.subpath ?? "";
        if (req && !sub.startsWith(req)) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                message: "Subpath must include the required prefix",
                path: ["bindOptions", "subpath"],
            });
        }
    }
    if (val.type === EMountType.Volume) {
        const req = val.volumeOptions.subpathRequired ?? "";
        const sub = val.volumeOptions.subpath ?? "";
        if (req && !sub.startsWith(req)) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                message: "Subpath must include the required prefix",
                path: ["volumeOptions", "subpath"],
            });
        }
    }
    if (val.type === EMountType.Cluster) {
        const req = val.clusterOptions.subpathRequired ?? "";
        const sub = val.clusterOptions.subpath ?? "";
        if (req && !sub.startsWith(req)) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                message: "Subpath must include the required prefix",
                path: ["clusterOptions", "subpath"],
            });
        }
    }
});

export type StorageMountFormInput = z.input<typeof StorageMountFormSchema>;
export type StorageMountFormOutput = z.output<typeof StorageMountFormSchema>;
