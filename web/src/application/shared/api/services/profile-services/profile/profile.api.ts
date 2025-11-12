import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import {
    type ProfileApiValidator,
    type Profile_Complete2FASetup_Req,
    type Profile_Complete2FASetup_Res,
    type Profile_GetProfile2FASetup_Req,
    type Profile_GetProfile2FASetup_Res,
} from "@application/shared/api/services";

import { BaseApi, parseApiError } from "@infrastructure/api";

export class ProfileApi extends BaseApi {
    public constructor(private readonly validator: ProfileApiValidator) {
        super();
    }

    // /**
    //  * Update profile
    //  */
    // async update(request: Profile_UpdateProfile_Req): Promise<Result<Profile_UpdateProfile_Res, Error>> {
    //     const { data } = request;

    //     return lastValueFrom(
    //         from(
    //             this.client.v1.patch("/users/current/profile", {
    //                 firstName: data.firstName,
    //                 lastName: data.lastName,
    //                 mobilePhone: data.mobilePhone,
    //                 officePhone: data.officePhone,
    //             }),
    //         ).pipe(
    //             map(this.validator.update),
    //             map(res => Ok(res)),
    //             catchError(error => of(Err(parseApiError(error)))),
    //         ),
    //     );
    // }

    // /**
    //  * Update profile photo
    //  */
    // async updatePhoto(request: Profile_UpdateProfilePhoto_Req): Promise<Result<Profile_UpdateProfilePhoto_Res, Error>> {
    //     const { data } = request;

    //     switch (data.photo) {
    //         /**
    //          * Delete photo
    //          */
    //         case null: {
    //             return lastValueFrom(
    //                 from(
    //                     this.client.v1.delete("/users/current/photo", {
    //                         headers: {
    //                             [WORKSPACE_HEADER]: request.meta.workspaceId,
    //                         },
    //                     }),
    //                 ).pipe(
    //                     map(() => Ok({ data: { url: null } })),
    //                     catchError(error => of(Err(parseApiError(error)))),
    //                 ),
    //             );
    //         }

    //         /**
    //          * Upload photo
    //          */
    //         default: {
    //             const form = new FormData();

    //             form.append("type", "avatar");
    //             form.append("file", data.photo, data.photo.name);

    //             return lastValueFrom(
    //                 from(
    //                     this.client.v1.post("/files/upload", form, {
    //                         headers: {
    //                             [WORKSPACE_HEADER]: request.meta.workspaceId,
    //                             "Content-Type": "multipart/form-data",
    //                         },
    //                     }),
    //                 ).pipe(
    //                     map(this.validator.updatePhoto),
    //                     map(res => Ok(res)),
    //                     catchError(error => of(Err(parseApiError(error)))),
    //                 ),
    //             );
    //         }
    //     }
    // }

    // /**
    //  * Update profile email
    //  */
    // async updateEmail(request: Profile_UpdateProfileEmail_Req): Promise<Result<Profile_UpdateProfileEmail_Res, Error>> {
    //     const { data } = request;

    //     return lastValueFrom(
    //         from(
    //             this.client.v1.patch("/users/current/email", {
    //                 email: data.email,
    //                 currentPassword: data.password,
    //             }),
    //         ).pipe(
    //             map(this.validator.updateEmail),
    //             map(res => Ok(res)),
    //             catchError(error => of(Err(parseApiError(error)))),
    //         ),
    //     );
    // }

    // /**
    //  * Update profile password
    //  */
    // async updatePassword(
    //     request: Profile_UpdateProfilePassword_Req,
    // ): Promise<Result<Profile_UpdateProfilePassword_Res, Error>> {
    //     const { data } = request;

    //     return lastValueFrom(
    //         from(
    //             this.client.v1.patch("/users/current/password", {
    //                 currentPassword: data.currentPassword,
    //                 newPassword: data.newPassword,
    //             }),
    //         ).pipe(
    //             map(this.validator.updatePassword),
    //             map(res => Ok(res)),
    //             catchError(error => of(Err(parseApiError(error)))),
    //         ),
    //     );
    // }

    // /**
    //  * Update profile locale
    //  */
    // async updateLocale(
    //     request: Profile_UpdateProfileLocale_Req,
    // ): Promise<Result<Profile_UpdateProfileLocale_Res, Error>> {
    //     const { data } = request;

    //     return lastValueFrom(
    //         from(
    //             this.client.v1.patch("/users/current/locale", {
    //                 language: data.language,
    //                 timezone: data.timezone,
    //             }),
    //         ).pipe(
    //             map(this.validator.updateLocale),
    //             map(res => Ok(res)),
    //             catchError(error => of(Err(parseApiError(error)))),
    //         ),
    //     );
    // }

    /**
     * Get profile 2FA setup
     */
    async getProfile2FASetup(
        request: Profile_GetProfile2FASetup_Req,
    ): Promise<Result<Profile_GetProfile2FASetup_Res, Error>> {
        const { data } = request;
        return lastValueFrom(
            from(
                this.client.v1.post("/users/current/mfa/totp-begin-setup", {
                    passcode: data.passcode,
                }),
            ).pipe(
                map(this.validator.getProfile2FASetup),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Verify profile 2FA setup
     */
    async complete2FASetup(
        request: Profile_Complete2FASetup_Req,
    ): Promise<Result<Profile_Complete2FASetup_Res, Error>> {
        const { data } = request;
        return lastValueFrom(
            from(
                this.client.v1.post("/users/current/mfa/totp-complete-setup", {
                    totpToken: data.totpToken,
                    passcode: data.passcode,
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" as const } })),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
