import { type AxiosResponse } from "axios";
import { z } from "zod";

import { type Public_Users_FindManyBase_Res } from "@application/shared/api-public/services";
import { EUserRole } from "@application/shared/enums";

import { PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

const UserBaseSchema = z.object({
    id: z.string(),
    username: z.string(),
    email: z.string(),
    fullName: z.string(),
    photo: z.string().nullable(),
    role: z.nativeEnum(EUserRole),
});

const FindManyBaseSchema = z.object({
    data: z.array(UserBaseSchema),
    meta: PagingMetaApiSchema,
});

export class UsersPublicApiValidator {
    findManyBase = (response: AxiosResponse): Public_Users_FindManyBase_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManyBaseSchema,
        });

        return {
            data,
            meta,
        };
    };
}
