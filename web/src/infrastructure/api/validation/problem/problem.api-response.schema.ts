import { z } from "zod";

export const ProblemApiResponseSchema = z.object({
    type: z.string().optional(),
    title: z.string(),
    status: z.number(),
    code: z.string(),
    detail: z.string(),
    displayLevel: z.enum(["high", "medium", "low"]).optional(),
    /**
     * Values a client may act on, beside the message they appear in: the limit a
     * refused creation reached, for one. Most problems have none.
     *
     * Every failed request passes through this schema, so a params it cannot read
     * is dropped rather than failing the problem: the message is still worth
     * showing.
     */
    params: z.record(z.unknown()).nullish().catch(null),
});
