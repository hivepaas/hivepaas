import axios from "axios";
import { jwtDecode } from "jwt-decode";
import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";
import { z } from "zod";

import { EnvConfig } from "@config";

import { type AppJwtPayload } from "@infrastructure/api/types";
import { parseApiError, parseApiResponse } from "@infrastructure/api/utils/api.data.utils";
import { isTransientException } from "@infrastructure/api/utils/api.exceptions.utils";

/**
 * Check if the token is expired
 */
export function isTokenExpired(token: string): boolean {
    try {
        const payload = jwtDecode<AppJwtPayload>(token);

        return payload.exp !== undefined && payload.exp * 1000 < Date.now();
    } catch {
        console.error(token, "Invalid access token");
    }

    return false;
}

/**
 * Refresh token API response schema
 */
const RefreshSchema = z.object({
    data: z.object({
        accessToken: z.string(),
    }),
});

/**
 * How long to wait before each retry of a refresh, and how many there are
 *
 * Short and few on purpose: a refresh runs in front of a request somebody is
 * waiting on, so this is spending their time. Two tries past the first cover the
 * case this exists for - a dropped connection, a gateway being restarted - and
 * anything still failing after a second is not a blip.
 */
const refreshRetryDelaysMs = [300, 900];

function wait(ms: number): Promise<void> {
    return new Promise(resolve => {
        setTimeout(resolve, ms);
    });
}

function requestRefresh(): Promise<Result<string, Error>> {
    return lastValueFrom(
        from(
            axios.post(
                "/sessions/refresh",
                {},
                {
                    baseURL: EnvConfig.API_URL,
                    // withCredentials: true,
                },
            ),
        ).pipe(
            map(response => {
                return parseApiResponse({
                    response,
                    schema: RefreshSchema,
                });
            }),
            map(res => Ok(res.data.accessToken)),
            catchError(error => of(Err(parseApiError(error)))),
        ),
    );
}

/**
 * Refresh the access token
 *
 * Retries a failure that says nothing about the session, because the session is
 * held in a cookie the server still honours: a connection that dropped at the
 * wrong moment is not a reason to end it, and before this it was - every failure
 * here signed the person out.
 *
 * The error of the last attempt is what comes back, so the caller can still tell
 * "the server says this session is finished" from "we could not reach it".
 */
export async function refreshToken(): Promise<Result<string, Error>> {
    for (let attempt = 0; ; attempt++) {
        const result = await requestRefresh();

        if (result.isOk() || attempt >= refreshRetryDelaysMs.length) {
            return result;
        }

        const error = result.unwrapErr();

        if (!isTransientException(error)) {
            return result;
        }

        await wait(refreshRetryDelaysMs[attempt] ?? 0);
    }
}
