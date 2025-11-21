import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { type Session_GetProfile_Res, type Session_Logout_Res } from "@application/shared/api/services";

import { BaseApi, parseApiError } from "@infrastructure/api";

import { type SessionApiValidator } from "./session.api.validator";

export class SessionApi extends BaseApi {
    public constructor(private readonly validator: SessionApiValidator) {
        super();
    }

    /**
     * Get profile
     */
    async getProfile(signal?: AbortSignal): Promise<Result<Session_GetProfile_Res, Error>> {
        return lastValueFrom(
            from(
                this.client.v1.get("/sessions/me", {
                    signal,
                    params: {
                        getAccesses: true,
                    },
                }),
            ).pipe(
                map(this.validator.getProfile),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Logout
     */
    async logout(): Promise<Result<Session_Logout_Res, Error>> {
        return lastValueFrom(
            from(
                this.client.v1.delete("/sessions", {
                    // withCredentials: true,
                }),
            ).pipe(
                map(this.validator.logout),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
