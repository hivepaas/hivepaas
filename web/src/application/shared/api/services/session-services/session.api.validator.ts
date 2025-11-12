import { type AxiosResponse } from "axios";
import { z } from "zod";

import { type Session_GetProfile_Res, type Session_Logout_Res } from "@application/shared/api/services";
import { ESecuritySettings } from "@application/shared/enums";

import { parseApiResponse } from "@infrastructure/api";

/**
 * Get profile API response schema
 */
const GetProfileSchema = z.object({
    data: z.object({
        nextStep: z.string().optional(),
        user: z.object({
            id: z.string(),
            fullName: z.string(),
            photo: z.string().transform(arg => arg.trim() || null),
            email: z.string(),
            securityOption: z.nativeEnum(ESecuritySettings),
            mfaSecret: z.string().optional(),
        }),
    }),
});

export class SessionApiValidator {
    /**
     * Validate and transform the get profile API response
     */
    getProfile = (response: AxiosResponse): Session_GetProfile_Res => {
        const {
            data: { user, nextStep },
        } = parseApiResponse({
            response,
            schema: GetProfileSchema,
        });

        return {
            data: {
                id: user.id,
                fullName: user.fullName,
                photo: user.photo,
                email: user.email,
                securityOption: user.securityOption,
                mfaSecret: user.mfaSecret ?? "",
                nextStep,
            },
        };
    };

    /**
     * Validate and transform the logout API response
     */
    logout = (_: AxiosResponse): Session_Logout_Res => {
        return {
            data: {
                type: "success",
            },
        };
    };
}
