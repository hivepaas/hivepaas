import type { AxiosResponse } from "axios";
import { z } from "zod";

import { parseApiResponse } from "@infrastructure/api";

import type { SupportFeedbacks_CreateOne_Res } from "./support-feedbacks.api.contracts";

const CreateOneSchema = z.object({
    data: z.unknown().optional(),
});

export class SupportFeedbacksApiValidator {
    createOne = (response: AxiosResponse): SupportFeedbacks_CreateOne_Res => {
        parseApiResponse({
            response,
            schema: CreateOneSchema,
        });

        return { data: { type: "success" } };
    };
}
