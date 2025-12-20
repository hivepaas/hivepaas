import { z } from "zod";

export const ProblemApiResponseSchema = z.object({
    type: z.string().optional(),
    title: z.string(),
    status: z.number(),
    code: z.string(),
    detail: z.string(),
    displayLevel: z.enum(["high", "medium", "low"]).optional(),
});
