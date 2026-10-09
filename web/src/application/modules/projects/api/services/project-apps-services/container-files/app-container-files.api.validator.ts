import { Err, Ok, type Result } from "oxide.ts";
import { z } from "zod";

import { parseApiProblem } from "@infrastructure/api";

import { UnexpectedApiErrorException, UnexpectedApiResponseException } from "@infrastructure/exceptions/api";

import type { AppContainerFiles_UploadOne_Res } from "./app-container-files.api.contracts";

// What the server sends over an upload's stream: how much the copy has taken,
// then done, with what the file became, or the error the server would have
// answered a request with.
const UploadMessageSchema = z.discriminatedUnion("type", [
    z.object({
        type: z.literal("progress"),
        received: z.number(),
    }),
    z.object({
        type: z.literal("done"),
        data: z.object({
            path: z.string(),
            message: z.string(),
        }),
    }),
    z.object({
        type: z.literal("error"),
        error: z.unknown(),
    }),
]);

export type AppContainerFiles_UploadMessage =
    | { type: "progress"; received: number }
    | { type: "answer"; result: Result<AppContainerFiles_UploadOne_Res, Error> };

export class AppContainerFilesApiValidator {
    uploadMessage = (message: string): AppContainerFiles_UploadMessage => {
        let json: unknown;
        try {
            json = JSON.parse(message);
        } catch {
            return { type: "answer", result: Err(new UnexpectedApiResponseException()) };
        }

        const parsed = UploadMessageSchema.safeParse(json);
        if (!parsed.success) {
            if (import.meta.env["NODE_ENV"] !== "production") {
                console.warn(json, parsed.error.format());
            }

            return { type: "answer", result: Err(new UnexpectedApiResponseException()) };
        }

        switch (parsed.data.type) {
            case "progress":
                return { type: "progress", received: parsed.data.received };
            case "error":
                return {
                    type: "answer",
                    result: Err(parseApiProblem(parsed.data.error) ?? new UnexpectedApiErrorException()),
                };
            case "done":
                return { type: "answer", result: Ok({ data: parsed.data.data }) };
        }
    };
}
