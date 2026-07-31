import { z } from "zod";

export const BaseMetaApiSchema = z.object({
    code: z.string().optional(),
    message: z.string().optional(),
    warning: z.string().optional(),
    error: z.string().optional(),
});
