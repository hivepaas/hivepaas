import { z } from "zod";

export const CreateProfileApiKeyFormSchema = z.object({
    name: z.string().trim().min(1, "Name is required"),
    accessAction: z.object({
        read: z.boolean(),
        write: z.boolean(),
        delete: z.boolean(),
    }),
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

export type CreateProfileApiKeyFormSchemaInput = z.input<typeof CreateProfileApiKeyFormSchema>;
export type CreateProfileApiKeyFormSchemaOutput = z.output<typeof CreateProfileApiKeyFormSchema>;
