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
    /**
     * Required, and within a year, as the server takes a key: one that never
     * expires is the one nobody remembers to revoke.
     */
    expireAt: z
        .date({ required_error: "Choose when the key expires", invalid_type_error: "Choose when the key expires" })
        .refine(
            date => {
                const now = new Date();
                now.setSeconds(0, 0);
                return date > now;
            },
            { message: "Expiration date must be in the future" },
        )
        .refine(date => date <= latestExpiration(), { message: "A key expires within a year" }),
});

/** latestExpiration is the last moment a key may be made to expire: a year from now. */
export function latestExpiration(): Date {
    const date = new Date();
    date.setFullYear(date.getFullYear() + 1);
    return date;
}

export type CreateProfileApiKeyFormSchemaInput = z.output<typeof CreateProfileApiKeyFormSchema>;
export type CreateProfileApiKeyFormSchemaOutput = z.output<typeof CreateProfileApiKeyFormSchema>;
