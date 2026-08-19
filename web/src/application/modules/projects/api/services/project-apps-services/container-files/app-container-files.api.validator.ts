import type { AxiosResponse } from "axios";
import { z } from "zod";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type { AppContainerFiles_UploadOne_Res } from "./app-container-files.api.contracts";

const UploadOneSchema = z.object({
    data: z.object({
        path: z.string(),
        message: z.string(),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

export class AppContainerFilesApiValidator {
    uploadOne = (response: AxiosResponse): AppContainerFiles_UploadOne_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: UploadOneSchema,
        });

        return { data, meta };
    };
}
