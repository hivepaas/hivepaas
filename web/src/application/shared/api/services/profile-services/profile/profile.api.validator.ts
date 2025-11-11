import { type AxiosResponse } from "axios";
import { z } from "zod";

import { type Profile_GetProfile2FASetup_Res } from "@application/shared/api/services";

import { parseApiResponse } from "@infrastructure/api";

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
            },
        };
    };
}
