import type { SupportFeedbackCategory } from "@application/shared/enums";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export interface SupportFeedbackPayload {
    category: SupportFeedbackCategory;
    name?: string;
    email?: string;
    company?: string;
    subject: string;
    description: string;
}

export type SupportFeedbacks_CreateOne_Req = ApiRequestBase<SupportFeedbackPayload>;
export type SupportFeedbacks_CreateOne_Res = ApiResponseBase<{
    type: "success";
}>;
