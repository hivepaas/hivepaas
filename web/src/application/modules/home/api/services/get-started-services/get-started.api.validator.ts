import { type AxiosResponse } from "axios";
import { z } from "zod";

import { SetupChecklistItemSchema } from "@application/shared/api/services";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type { GetStarted_Dismiss_Res, GetStarted_RequestDashboardCert_Res } from "./get-started.api.contracts";

const RequestDashboardCertSchema = z.object({
    data: SetupChecklistItemSchema,
    meta: BaseMetaApiSchema.nullish(),
});

export class GetStartedApiValidator {
    requestDashboardCert = (response: AxiosResponse): GetStarted_RequestDashboardCert_Res => {
        const { data, meta } = parseApiResponse({ response, schema: RequestDashboardCertSchema });
        return { data, meta };
    };

    dismiss = (_: AxiosResponse): GetStarted_Dismiss_Res => {
        return { data: { type: "success" } };
    };
}
