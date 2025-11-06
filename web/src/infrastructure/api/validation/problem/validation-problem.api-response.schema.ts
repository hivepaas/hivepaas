import { z } from "zod";

import { ProblemApiResponseSchema } from "@infrastructure/api";

export const ValidationProblemApiResponseSchema = z
    .object({
        errors: z.array(
            z.object({
                path: z.string(),
                code: z.string(),
                message: z.string(),
            }),
        ),
    })
    .merge(ProblemApiResponseSchema);
