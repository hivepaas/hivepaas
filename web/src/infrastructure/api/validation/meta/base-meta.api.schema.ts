import { z } from "zod";

export const BaseMetaApiSchema = z.object({
    code: z.string().optional(),
    message: z.string().optional(),
});
