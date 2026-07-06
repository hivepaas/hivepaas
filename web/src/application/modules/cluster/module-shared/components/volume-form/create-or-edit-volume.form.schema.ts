import { z } from "zod";
import {
    EClusterVolumeDriverMode,
    EClusterVolumeLocalType,
    EClusterVolumePropagation,
} from "~/cluster/module-shared/enums";

const KeyValueSchema = z.object({
    key: z.string().trim().min(1, "Name is required"),
    value: z.string().trim(),
});

const BindOptionsSchema = z.object({
    directory: z.string().trim(),
    propagation: z.nativeEnum(EClusterVolumePropagation),
    readonly: z.boolean(),
    extraOptions: z.string().trim(),
});

const NfsOptionsSchema = z.object({
    addr: z.string().trim(),
    device: z.string().trim(),
    version: z.string().trim(),
    readonly: z.boolean(),
    extraOptions: z.string().trim(),
});

const TmpfsOptionsSchema = z.object({
    device: z.string().trim(),
    size: z.string().trim(),
    uid: z.string().trim(),
    extraOptions: z.string().trim(),
});

function hasDuplicateKeys(items: { key: string }[]) {
    const seen = new Set<string>();

    return items.some(item => {
        const key = item.key.trim();
        if (seen.has(key)) {
            return true;
        }

        seen.add(key);
        return false;
    });
}

export const CreateOrEditVolumeFormSchema = z
    .object({
        name: z.string().trim().min(1, "Name is required").max(100, "Name must be 100 characters or less"),
        driverMode: z.nativeEnum(EClusterVolumeDriverMode),
        customDriverName: z.string().trim(),
        localType: z.nativeEnum(EClusterVolumeLocalType),
        bindOptions: BindOptionsSchema,
        nfsOptions: NfsOptionsSchema,
        tmpfsOptions: TmpfsOptionsSchema,
        labels: z.array(KeyValueSchema),
        options: z.array(KeyValueSchema),
        availableInProjects: z.boolean(),
        default: z.boolean(),
    })
    .superRefine((values, ctx) => {
        // if (values.driverMode === EClusterVolumeDriverMode.Custom && !values.customDriverName) {
        //     ctx.addIssue({
        //         code: z.ZodIssueCode.custom,
        //         path: ["customDriverName"],
        //         message: "Driver name is required",
        //     });
        // }

        if (values.driverMode === EClusterVolumeDriverMode.Local && values.localType === EClusterVolumeLocalType.Nfs) {
            if (!values.nfsOptions.addr) {
                ctx.addIssue({
                    code: z.ZodIssueCode.custom,
                    path: ["nfsOptions", "addr"],
                    message: "Address is required",
                });
            }

            if (!values.nfsOptions.device) {
                ctx.addIssue({
                    code: z.ZodIssueCode.custom,
                    path: ["nfsOptions", "device"],
                    message: "Device is required",
                });
            }
        }

        if (
            values.driverMode === EClusterVolumeDriverMode.Local &&
            values.localType === EClusterVolumeLocalType.Tmpfs &&
            !values.tmpfsOptions.size
        ) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["tmpfsOptions", "size"],
                message: "Size is required",
            });
        }

        if (hasDuplicateKeys(values.labels)) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["labels"],
                message: "Duplicate label names are not allowed",
            });
        }

        if (hasDuplicateKeys(values.options)) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["options"],
                message: "Duplicate option names are not allowed",
            });
        }
    });

export type CreateOrEditVolumeFormInput = z.input<typeof CreateOrEditVolumeFormSchema>;
export type CreateOrEditVolumeFormOutput = z.output<typeof CreateOrEditVolumeFormSchema>;
