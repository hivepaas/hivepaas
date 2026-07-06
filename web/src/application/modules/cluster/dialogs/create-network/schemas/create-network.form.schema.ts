import { z } from "zod";

const KeyValueSchema = z.object({
    key: z.string().trim().min(1, "Name is required"),
    value: z.string().trim(),
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

export const CreateNetworkFormSchema = z
    .object({
        name: z.string().trim().min(1, "Name is required").max(64, "Name must be 64 characters or less"),
        enableIPv4: z.boolean(),
        enableIPv6: z.boolean(),
        internal: z.boolean(),
        attachable: z.boolean(),
        ingress: z.boolean(),
        labels: z.array(KeyValueSchema),
        options: z.array(KeyValueSchema),
        availableInProjects: z.boolean(),
        default: z.boolean(),
    })
    .superRefine((values, ctx) => {
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

export type CreateNetworkFormInput = z.input<typeof CreateNetworkFormSchema>;
export type CreateNetworkFormOutput = z.output<typeof CreateNetworkFormSchema>;
