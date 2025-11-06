import { z } from "zod";

export const PagingMetaApiSchema = z.object({
    code: z.string().optional(),
    message: z.string().optional(),
    page: z.object({
        limit: z.number(),
        offset: z.number(),
        total: z.number(),
    }),
});
