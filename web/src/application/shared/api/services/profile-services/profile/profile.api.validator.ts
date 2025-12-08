import { type AxiosResponse } from "axios";
import { z } from "zod";

import type {
    Profile_CreateOneApiKey_Res,
    Profile_FindManyApiKeysPaginated_Res,
    Profile_GetProfile2FASetup_Res,
} from "@application/shared/api/services";
import { EProfileApiKeyStatus } from "@application/shared/enums";

import { PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

/**
 * Get profile 2FA setup API response schema
 */
const GetProfile2FASetupSchema = z.object({
    data: z.object({
        totpToken: z.string(),
        secret: z.string(),
        qrCode: z.object({
            dataBase64: z.string(),
        }),
    }),
});

/**
 * Account API Key schema
 */
const ProfileApiKeySchema = z.object({
    id: z.string(),
    name: z.string(),
    keyId: z.string(),
    updateVer: z.number(),
    accessAction: z
        .object({
            read: z.boolean(),
            write: z.boolean(),
            delete: z.boolean(),
        })
        .optional(),
    expireAt: z.coerce.date().optional(),
    status: z.nativeEnum(EProfileApiKeyStatus),
});

/**
 * Find many account API keys paginated API response schema
 */
const FindManyApiKeysPaginatedSchema = z.object({
    data: z.array(ProfileApiKeySchema),
    meta: PagingMetaApiSchema,
});

/**
 * Create one account API key API response schema
 */
const CreateOneApiKeySchema = z.object({
    data: z.object({
        id: z.string(),
        keyId: z.string(),
        secretKey: z.string(),
    }),
});

export class ProfileApiValidator {
    /**
     * Validate and transform get profile 2FA setup API response
     */
    getProfile2FASetup = (response: AxiosResponse): Profile_GetProfile2FASetup_Res => {
        const { data } = parseApiResponse({
            response,
            schema: GetProfile2FASetupSchema,
        });

        return {
            data: {
                totpToken: data.totpToken,
                totpQRCode: data.qrCode.dataBase64,
                secretKey: data.secret,
            },
        };
    };

    /**
     * Validate and transform find many account API keys paginated API response
     */
    findManyApiKeysPaginated = (response: AxiosResponse): Profile_FindManyApiKeysPaginated_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManyApiKeysPaginatedSchema,
        });

        return {
            data: data.map(apiKey => ({
                id: apiKey.id,
                name: apiKey.name,
                keyId: apiKey.keyId,
                updateVer: apiKey.updateVer,
                accessAction: apiKey.accessAction ?? null,
                expireAt: apiKey.expireAt,
                status: apiKey.status,
            })),
            meta,
        };
    };

    /**
     * Validate and transform create one account API key API response
     */
    createOneApiKey = (response: AxiosResponse): Profile_CreateOneApiKey_Res => {
        const { data } = parseApiResponse({
            response,
            schema: CreateOneApiKeySchema,
        });

        return {
            data: {
                id: data.id,
                keyId: data.keyId,
                secretKey: data.secretKey,
            },
        };
    };
}
