import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, JsonTransformer, parseApiError } from "@infrastructure/api";

import type {
    Profile_Complete2FASetup_Req,
    Profile_Complete2FASetup_Res,
    Profile_CreateOneApiKey_Req,
    Profile_CreateOneApiKey_Res,
    Profile_DeleteOneApiKey_Req,
    Profile_DeleteOneApiKey_Res,
    Profile_FindManyApiKeysPaginated_Req,
    Profile_FindManyApiKeysPaginated_Res,
    Profile_GetProfile2FASetup_Req,
    Profile_GetProfile2FASetup_Res,
    Profile_UpdateProfilePassword_Req,
    Profile_UpdateProfilePassword_Res,
    Profile_UpdateProfile_Req,
    Profile_UpdateProfile_Res,
} from "./profile.api.contracts";
import type { ProfileApiValidator } from "./profile.api.validator";

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

    /**
     * Update profile password
     */
    async updatePassword(
        request: Profile_UpdateProfilePassword_Req,
    ): Promise<Result<Profile_UpdateProfilePassword_Res, Error>> {
        const { data } = request;

        return lastValueFrom(
            from(
                this.client.v1.put("/users/current/password", {
                    currentPassword: data.currentPassword,
                    newPassword: data.newPassword,
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } as const })),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

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

    /**
     * Find many profile API keys paginated
     */
    async findManyApiKeysPaginated(
        request: Profile_FindManyApiKeysPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<Profile_FindManyApiKeysPaginated_Res, Error>> {
        const { search, pagination, sorting } = request.data;

        const query = this.queryBuilder.getInstance();

        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get("/users/current/settings/api-keys", {
                    params: query.build(),
                    signal,
                }),
            ).pipe(
                map(this.validator.findManyApiKeysPaginated),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Create one profile API key
     */
    async createOneApiKey(
        request: Profile_CreateOneApiKey_Req,
        signal?: AbortSignal,
    ): Promise<Result<Profile_CreateOneApiKey_Res, Error>> {
        const { name, accessAction, expireAt } = request.data;

        const json = {
            name,
            accessAction,
            expireAt: expireAt ? JsonTransformer.date({ data: expireAt, some: date => date.toISOString() }) : undefined,
        };

        return lastValueFrom(
            from(
                this.client.v1.post("/users/current/api-keys", json, {
                    signal,
                }),
            ).pipe(
                map(this.validator.createOneApiKey),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Delete one profile API key
     */
    async deleteOneApiKey(request: Profile_DeleteOneApiKey_Req): Promise<Result<Profile_DeleteOneApiKey_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/users/current/api-keys/${id}`, {})).pipe(
                map(() => Ok({ data: { id } })),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
