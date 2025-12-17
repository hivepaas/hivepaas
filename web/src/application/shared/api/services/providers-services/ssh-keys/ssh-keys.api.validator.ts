import { type AxiosResponse } from "axios";
import { z } from "zod";

import { PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import { type SshKeys_FindManyPaginated_Res } from "./ssh-keys.api.contracts";

const SshKeySchema = z.object({
    id: z.string(),
    name: z.string(),
});

const FindManyPaginatedSchema = z.object({
    data: z.array(SshKeySchema),
    meta: PagingMetaApiSchema,
});

export class SshKeysApiValidator {
    findManyPaginated = (response: AxiosResponse): SshKeys_FindManyPaginated_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindManyPaginatedSchema });
        return { data, meta };
    };
}
