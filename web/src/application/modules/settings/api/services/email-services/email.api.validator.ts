import { type AxiosResponse } from "axios";
import { z } from "zod";
import { EmailSettingEntitySchema } from "~/settings/module-shared/schemas";

import { BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    Email_CreateOne_Res,
    Email_DeleteOne_Res,
    Email_FindManyPaginated_Res,
    Email_FindOneById_Res,
    Email_TestSendMail_Res,
    Email_UpdateOne_Res,
    Email_UpdateStatus_Res,
} from "./email.api.contracts";

const FindManyPaginatedSchema = z.object({
    data: z.array(EmailSettingEntitySchema),
    meta: PagingMetaApiSchema,
});

const FindOneByIdSchema = z.object({
    data: EmailSettingEntitySchema,
    meta: BaseMetaApiSchema.nullish(),
});

const CreateOneSchema = z.object({
    data: z.object({
        id: z.string(),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class EmailApiValidator {
    findManyPaginated = (response: AxiosResponse): Email_FindManyPaginated_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManyPaginatedSchema,
        });

        return { data, meta };
    };

    findOneById = (response: AxiosResponse): Email_FindOneById_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindOneByIdSchema,
        });

        return { data, meta };
    };

    createOne = (response: AxiosResponse): Email_CreateOne_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: CreateOneSchema,
        });

        return { data, meta };
    };

    updateOne = (response: AxiosResponse): Email_UpdateOne_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };

    updateStatus = (response: AxiosResponse): Email_UpdateStatus_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };

    deleteOne = (response: AxiosResponse): Email_DeleteOne_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };

    testSendMail = (response: AxiosResponse): Email_TestSendMail_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };
}
