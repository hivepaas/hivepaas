import { z } from "zod";

const StringValueListSchema = z
    .array(
        z.object({
            value: z.string().transform(value => value.trim()),
        }),
    )
    .superRefine((items, ctx) => {
        const seen = new Set<string>();

        items.forEach((item, index) => {
            if (!item.value) {
                return;
            }

            if (seen.has(item.value)) {
                ctx.addIssue({
                    code: "custom",
                    message: `"${item.value}" already exists`,
                    path: [index],
                });
                return;
            }

            seen.add(item.value);
        });
    });

const VolumeListSchema = z
    .array(
        z.object({
            id: z.string(),
            name: z.string(),
        }),
    )
    .superRefine((items, ctx) => {
        const seen = new Set<string>();

        items.forEach((item, index) => {
            if (seen.has(item.id)) {
                ctx.addIssue({
                    code: "custom",
                    message: `"${item.name}" already exists`,
                    path: [index],
                });
                return;
            }

            seen.add(item.id);
        });
    });

const trimmedString = z.string().transform(value => value.trim());

export const ProjectStorageSettingsFormSchema = z.object({
    bindSettings: z.object({
        enabled: z.boolean(),
        baseDirs: StringValueListSchema,
        subpathTemplate: trimmedString,
    }),
    volumeSettings: z.object({
        enabled: z.boolean(),
        volumes: VolumeListSchema,
        subpathTemplate: trimmedString,
    }),
    clusterVolumeSettings: z.object({
        enabled: z.boolean(),
        volumes: VolumeListSchema,
        subpathTemplate: trimmedString,
    }),
    tmpfsSettings: z.object({
        enabled: z.boolean(),
        maxSize: trimmedString,
    }),
});

export type ProjectStorageSettingsFormSchemaInput = z.input<typeof ProjectStorageSettingsFormSchema>;
export type ProjectStorageSettingsFormSchemaOutput = z.output<typeof ProjectStorageSettingsFormSchema>;

export const emptyProjectStorageSettingsFormDefaults: ProjectStorageSettingsFormSchemaInput = {
    bindSettings: {
        enabled: false,
        baseDirs: [],
        subpathTemplate: "",
    },
    volumeSettings: {
        enabled: false,
        volumes: [],
        subpathTemplate: "",
    },
    clusterVolumeSettings: {
        enabled: false,
        volumes: [],
        subpathTemplate: "",
    },
    tmpfsSettings: {
        enabled: false,
        maxSize: "",
    },
};
