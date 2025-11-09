import axios from "axios";
import { jwtDecode } from "jwt-decode";
import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";
import { z } from "zod";

import { EnvConfig } from "@config";

import { type AppJwtPayload } from "@infrastructure/api/types";
import { parseApiError, parseApiResponse } from "@infrastructure/api/utils/api.data.utils";

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
 * Refresh the access token
 */
export async function refreshToken(): Promise<Result<string, Error>> {
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
