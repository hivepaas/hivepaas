import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import {
    type ProfileApiValidator,
    type Profile_Complete2FASetup_Req,
    type Profile_Complete2FASetup_Res,
    type Profile_GetProfile2FASetup_Req,
    type Profile_GetProfile2FASetup_Res,
    type Profile_UpdateProfile_Req,
    type Profile_UpdateProfile_Res,
} from "@application/shared/api/services";

import { BaseApi, JsonTransformer, parseApiError } from "@infrastructure/api";

export class ProfileApi extends BaseApi {
    public constructor(private readonly validator: ProfileApiValidator) {
        super();
    }

    /**
     * Update profile
     */
    async update(
        request: Profile_UpdateProfile_Req,
        signal?: AbortSignal,
    ): Promise<Result<Profile_UpdateProfile_Res, Error>> {
        const { profile } = request.data;

        const json = {
            username: JsonTransformer.string({
                data: profile.username,
            }),
            email: JsonTransformer.string({
                data: profile.email,
            }),
            fullName: JsonTransformer.string({
                data: profile.fullName,
            }),
            position: JsonTransformer.string({
                data: profile.position,
            }),
            notes: JsonTransformer.string({
                data: profile.notes,
            }),
            photo: JsonTransformer.object({
                data: profile.photo,
            }),
        };

        return lastValueFrom(
            from(this.client.v1.put("/users/current/profile", json, { signal })).pipe(
                map(() => Ok({ data: { type: "success" } as const })),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

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
