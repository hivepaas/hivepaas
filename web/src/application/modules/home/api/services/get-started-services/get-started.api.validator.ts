import { type AxiosResponse } from "axios";
import { z } from "zod";
import type { DashboardCert } from "~/home/domain";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    GetStarted_Dismiss_Res,
    GetStarted_GetDashboardCert_Res,
    GetStarted_RequestDashboardCert_Res,
} from "./get-started.api.contracts";

const DashboardCertSchema = z
    .object({
        status: z.enum(["todo", "obtaining", "failed", "done"]).catch("todo"),
        domain: z.string().nullish(),
        error: z.string().nullish(),
    })
    .transform(
        (cert): DashboardCert => ({
            status: cert.status,
            domain: cert.domain ?? "",
            error: cert.error ?? "",
        }),
    );

const DashboardCertResponseSchema = z.object({
    data: DashboardCertSchema,
    meta: BaseMetaApiSchema.nullish(),
});

export class GetStartedApiValidator {
    getDashboardCert = (response: AxiosResponse): GetStarted_GetDashboardCert_Res => {
        const { data, meta } = parseApiResponse({ response, schema: DashboardCertResponseSchema });
        return { data, meta };
    };

    requestDashboardCert = (response: AxiosResponse): GetStarted_RequestDashboardCert_Res => {
        const { data, meta } = parseApiResponse({ response, schema: DashboardCertResponseSchema });
        return { data, meta };
    };

    dismiss = (_: AxiosResponse): GetStarted_Dismiss_Res => {
        return { data: { type: "success" } };
    };
}
