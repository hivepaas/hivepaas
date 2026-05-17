import { type AxiosResponse } from "axios";
import { z } from "zod";
import { SSHKeySettingEntitySchema } from "~/settings/module-shared/schemas";

import { BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    SSHKey_CreateOne_Res,
    SSHKey_DeleteOne_Res,
    SSHKey_FindManyPaginated_Res,
    SSHKey_FindOneById_Res,
    SSHKey_UpdateMeta_Res,
    SSHKey_UpdateOne_Res,
} from "./ssh-key.api.contracts";

const FindManyPaginatedSchema = z.object({
    data: z.array(SSHKeySettingEntitySchema),
    meta: PagingMetaApiSchema,
});

const FindOneByIdSchema = z.object({
    data: SSHKeySettingEntitySchema,
    meta: BaseMetaApiSchema.nullish(),
});

const CreateOneSchema = z.object({
    data: z.object({ id: z.string() }),
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class SSHKeyApiValidator {
    findManyPaginated = (response: AxiosResponse): SSHKey_FindManyPaginated_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindManyPaginatedSchema });
        return { data, meta };
    };

    findOneById = (response: AxiosResponse): SSHKey_FindOneById_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneByIdSchema });
        return { data, meta };
    };

    createOne = (response: AxiosResponse): SSHKey_CreateOne_Res => {
        const { data, meta } = parseApiResponse({ response, schema: CreateOneSchema });
        return { data, meta };
    };

    updateOne = (response: AxiosResponse): SSHKey_UpdateOne_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };

    updateMeta = (response: AxiosResponse): SSHKey_UpdateMeta_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };

    deleteOne = (response: AxiosResponse): SSHKey_DeleteOne_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };
}
