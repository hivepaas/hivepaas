import { z } from "zod";

export const CreateProfileApiKeyFormSchema = z.object({
    name: z.string().trim().min(1, "Name is required"),
    accessAction: z.object({
        read: z.boolean(),
        execute: z.boolean().optional().default(false),
        write: z.boolean(),
        delete: z.boolean(),
    }),
    /** Gives the key the owner's capability to reveal secrets. */
    allowRevealSecrets: z.boolean().optional().default(false),
    expireAt: z
        .date()
        .refine(
            date => {
                const now = new Date();
                now.setSeconds(0, 0);
                return date > now;
            },
            { message: "Expiration date must be in the future" },
        )
        .optional(),
});

export type CreateProfileApiKeyFormSchemaInput = z.output<typeof CreateProfileApiKeyFormSchema>;
export type CreateProfileApiKeyFormSchemaOutput = z.output<typeof CreateProfileApiKeyFormSchema>;
