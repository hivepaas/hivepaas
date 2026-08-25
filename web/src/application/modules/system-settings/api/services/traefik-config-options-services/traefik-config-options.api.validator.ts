import type { AxiosResponse } from "axios";
import { z } from "zod";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    TraefikConfigOptions_FindOne_Res,
    TraefikConfigOptions_UpdateOne_Res,
} from "./traefik-config-options.api.contracts";

const StartupCommandSchema = z.object({
    logLevel: z.string().optional().default(""),
    accessLog: z.boolean().optional().default(false),
    http3: z.boolean().optional().default(false),
    fastProxy: z.boolean().optional().default(false),
    openPorts: z.array(z.string()).optional().default([]),
    args: z.array(z.string()).default([]),
});

const FindOneSchema = z.object({
    data: z.object({
        startupCommand: StartupCommandSchema,
    }),
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class TraefikConfigOptionsApiValidator {
    findOne = (response: AxiosResponse): TraefikConfigOptions_FindOne_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneSchema });
        return { data, meta };
    };

    updateOne = (response: AxiosResponse): TraefikConfigOptions_UpdateOne_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };
}
