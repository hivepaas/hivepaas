import { type AxiosResponse } from "axios";
import { z } from "zod";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import { type AppFunction_TestRun_Res } from "./app-function.api.contracts";

const OptionalStringSchema = z
    .string()
    .nullish()
    .transform(value => value ?? "");

/**
 * A response's body, which may not be text: base64 on the wire, bytes here.
 */
const Base64BytesSchema = z
    .string()
    .nullish()
    .transform(value => {
        const binary = atob(value ?? "");
        const bytes = new Uint8Array(binary.length);
        for (let i = 0; i < binary.length; i++) {
            bytes[i] = binary.charCodeAt(i);
        }
        return bytes;
    });

const ValuesSchema = z
    .record(z.array(z.string()))
    .nullish()
    .transform(value => value ?? {});

const TestRunSchema = z.object({
    data: z.object({
        outcome: z.string(),
        status: z
            .number()
            .nullish()
            .transform(value => value ?? 0),
        headers: ValuesSchema,
        body: Base64BytesSchema,
        bodyTruncated: z
            .boolean()
            .nullish()
            .transform(value => value ?? false),
        requestId: OptionalStringSchema,
        durationMs: z
            .number()
            .nullish()
            .transform(value => value ?? 0),
        logs: OptionalStringSchema,
        logsTruncated: z
            .boolean()
            .nullish()
            .transform(value => value ?? false),
        error: OptionalStringSchema,
        exitCode: z
            .number()
            .nullish()
            .transform(value => value ?? 0),
        librariesBuilt: z
            .boolean()
            .nullish()
            .transform(value => value ?? false),
        librariesLog: OptionalStringSchema,
        lockFiles: z
            .array(z.object({ path: z.string(), content: OptionalStringSchema }))
            .nullish()
            .transform(value => value ?? []),
    }),
    meta: BaseMetaApiSchema.nullable(),
});

export class AppFunctionApiValidator {
    testRun = (response: AxiosResponse): AppFunction_TestRun_Res => {
        return parseApiResponse({ response, schema: TestRunSchema });
    };
}
