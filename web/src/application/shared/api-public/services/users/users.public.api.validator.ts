import { z } from "zod";

import { type Public_Users_FindMany_Res } from "@application/shared/api-public/services";
import { EUserRole } from "@application/shared/enums";

import { type ApiHttpResponse, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

const UserBaseSchema = z.object({
    id: z.string(),
    username: z.string(),
    email: z.string(),
    fullName: z.string(),
    photo: z.string().nullable(),
    role: z.nativeEnum(EUserRole),
});

const FindManySchema = z.object({
    data: z.array(UserBaseSchema),
    meta: PagingMetaApiSchema,
});

export class UsersPublicApiValidator {
    findMany = (response: ApiHttpResponse): Public_Users_FindMany_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManySchema,
        });

        return {
            data,
            meta,
        };
    };
}
