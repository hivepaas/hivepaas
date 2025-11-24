import { type AxiosResponse } from "axios";
import { z } from "zod";

import type {
    Profile_CreateOneApiKey_Res,
    Profile_FindManyApiKeysPaginated_Res,
    Profile_GetProfile2FASetup_Res,
} from "@application/shared/api/services";

import { PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

// /**
//  * Update profile photo schema
//  */
// const UpdatePhotoSchema = z.object({
//     data: z.tuple([
//         z.object({
//             url: z.string(),
//         }),
//     ]),
// });

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

export class ProfileApiValidator {
    // /**
    //  * Validate and transform update profile API response
    //  */
    // // update = (_: AxiosResponse): Profile_UpdateProfile_Res => {
    // //     return {
    // //         data: {
    // //             type: "success",
    // //         },
    // //     };
    // // };
    // // /**
    // //  * Validate and transform update profile photo API response
    // //  */
    // // updatePhoto = (response: AxiosResponse): Profile_UpdateProfilePhoto_Res => {
    // //     const { data } = parseApiResponse({
    // //         response,
    // //         schema: UpdatePhotoSchema,
    // //     });
    // //     return {
    // //         data: data[0],
    // //     };
    // // };
    // // /**
    // //  * Validate and transform update profile email API response
    // //  */
    // // updateEmail = (_: AxiosResponse): Profile_UpdateProfile_Res => {
    // //     return {
    // //         data: {
    // //             type: "success",
    // //         },
    // //     };
    // // };
    // // /**
    // //  * Validate and transform update profile password API response
    // //  */
    // // updatePassword = (_: AxiosResponse): Profile_UpdateProfile_Res => {
    // //     return {
    // //         data: {
    // //             type: "success",
    // //         },
    // //     };
    // // };
    // // /**
    // //  * Validate and transform update profile locale API response
    // //  */
    // // updateLocale = (_: AxiosResponse): Profile_UpdateProfileLocale_Res => {
    // //     return {
    // //         data: {
    // //             type: "success",
    // //         },
    // //     };
    // // };

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
     * Profile API Key schema
     */
    #ProfileApiKeySchema = z.object({
        id: z.string(),
        name: z.string(),
        key: z.string(),
        accessAction: z.object({
            read: z.boolean(),
            write: z.boolean(),
            delete: z.boolean(),
        }),
        expireAt: z.coerce.date(),
        status: z.string(),
    });

    /**
     * Find many profile API keys paginated API response schema
     */
    #FindManyApiKeysPaginatedSchema = z.object({
        data: z.array(this.#ProfileApiKeySchema),
        meta: PagingMetaApiSchema,
    });

    /**
     * Create one profile API key API response schema
     */
    #CreateOneApiKeySchema = z.object({
        data: this.#ProfileApiKeySchema,
    });

    /**
     * Validate and transform find many profile API keys paginated API response
     */
    findManyApiKeysPaginated = (response: AxiosResponse): Profile_FindManyApiKeysPaginated_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: this.#FindManyApiKeysPaginatedSchema,
        });

        return {
            data: data.map(apiKey => ({
                id: apiKey.id,
                name: apiKey.name,
                key: apiKey.key,
                accessAction: apiKey.accessAction,
                expireAt: apiKey.expireAt,
                status: apiKey.status,
            })),
            meta,
        };
    };

    /**
     * Validate and transform create one profile API key API response
     */
    createOneApiKey = (response: AxiosResponse): Profile_CreateOneApiKey_Res => {
        const { data } = parseApiResponse({
            response,
            schema: this.#CreateOneApiKeySchema,
        });

        return {
            data: {
                id: data.id,
                name: data.name,
                key: data.key,
                accessAction: data.accessAction,
                expireAt: data.expireAt,
                status: data.status,
            },
        };
    };
}
