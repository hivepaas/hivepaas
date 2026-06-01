import { z } from "zod";

export const AccessSchema = z.object({
    id: z.string(),
    name: z.string(),
    access: z.object({
        read: z.boolean(),
        execute: z.boolean(),
        write: z.boolean(),
        delete: z.boolean(),
    }),
});
