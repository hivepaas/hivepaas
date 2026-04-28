import { z } from "zod";
import { EMountConsistency, EMountPropagation, EMountType } from "~/projects/module-shared/enums";

const VolumeDriverSchema = z.object({
    name: z.string().optional(),
    options: z.record(z.string()).optional(),
});

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
        labels: z.record(z.string()).optional(),
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
        labels: z.record(z.string()).optional(),
        driverConfig: VolumeDriverSchema.optional(),
    }),
});

const TmpfsMountSchema = z.object({
    type: z.literal(EMountType.Tmpfs),
    target: z.string().min(1, "Target is required"),
    readOnly: z.boolean().optional(),
    consistency: z.nativeEnum(EMountConsistency).optional(),
    tmpfsOptions: z.object({
        size: z.string().optional(),
        mode: z.number().optional(),
        options: z.array(z.array(z.string())).optional(),
    }),
});

export const StorageMountFormSchema = z.discriminatedUnion("type", [
    BindMountSchema,
    VolumeMountSchema,
    ClusterMountSchema,
    TmpfsMountSchema,
]);

export type StorageMountFormInput = z.input<typeof StorageMountFormSchema>;
export type StorageMountFormOutput = z.output<typeof StorageMountFormSchema>;
