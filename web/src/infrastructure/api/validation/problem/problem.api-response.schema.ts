import { z } from "zod";

export const ProblemApiResponseSchema = z.object({
    type: z.string(),
    title: z.string(),
    status: z.number(),
    code: z.string(),
    detail: z.string(),
});
