import { Err, Ok, type Result } from "oxide.ts";
import { z } from "zod";

import { parseApiProblem } from "@infrastructure/api";

import { UnexpectedApiErrorException, UnexpectedApiResponseException } from "@infrastructure/exceptions/api";

import type { AppContainerFiles_UploadOne_Res } from "./app-container-files.api.contracts";

// What an upload's stream ends with: done, with what the file became, or the
// error the server would have answered a request with.
const UploadAnswerSchema = z.discriminatedUnion("type", [
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

export class AppContainerFilesApiValidator {
    uploadAnswer = (message: string): Result<AppContainerFiles_UploadOne_Res, Error> => {
        let json: unknown;
        try {
            json = JSON.parse(message);
        } catch {
            return Err(new UnexpectedApiResponseException());
        }

        const parsed = UploadAnswerSchema.safeParse(json);
        if (!parsed.success) {
            return Err(new UnexpectedApiResponseException());
        }

        if (parsed.data.type === "error") {
            return Err(parseApiProblem(parsed.data.error) ?? new UnexpectedApiErrorException());
        }

        return Ok({ data: parsed.data.data });
    };
}
