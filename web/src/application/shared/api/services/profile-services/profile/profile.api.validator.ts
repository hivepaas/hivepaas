// import { type AxiosResponse } from "axios";
// import { z } from "zod";

// import {
//     type Profile_UpdateProfileLocale_Res,
//     type Profile_UpdateProfilePhoto_Res,
//     type Profile_UpdateProfile_Res,
// } from "@application/shared/api/services";

// import { parseApiResponse } from "@infrastructure/api";

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

// export class ProfileApiValidator {
//     /**
//      * Validate and transform update profile API response
//      */
//     update = (_: AxiosResponse): Profile_UpdateProfile_Res => {
//         return {
//             data: {
//                 type: "success",
//             },
//         };
//     };

//     /**
//      * Validate and transform update profile photo API response
//      */
//     updatePhoto = (response: AxiosResponse): Profile_UpdateProfilePhoto_Res => {
//         const { data } = parseApiResponse({
//             response,
//             schema: UpdatePhotoSchema,
//         });

//         return {
//             data: data[0],
//         };
//     };

//     /**
//      * Validate and transform update profile email API response
//      */
//     updateEmail = (_: AxiosResponse): Profile_UpdateProfile_Res => {
//         return {
//             data: {
//                 type: "success",
//             },
//         };
//     };

//     /**
//      * Validate and transform update profile password API response
//      */
//     updatePassword = (_: AxiosResponse): Profile_UpdateProfile_Res => {
//         return {
//             data: {
//                 type: "success",
//             },
//         };
//     };

//     /**
//      * Validate and transform update profile locale API response
//      */
//     updateLocale = (_: AxiosResponse): Profile_UpdateProfileLocale_Res => {
//         return {
//             data: {
//                 type: "success",
//             },
//         };
//     };
// }
