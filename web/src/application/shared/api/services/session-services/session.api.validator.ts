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
        user: z.object({
            id: z.string(),
            firstName: z.string(),
            lastName: z.string(),
            fullName: z.string(),
            photo: z.string().transform(arg => arg.trim() || null),
            email: z.string(),
            mobilePhone: z.string(),
            officePhone: z.string(),
            securityOption: z.nativeEnum(ESecuritySettings),

            timezone: z
                .string()
                .trim()
                .transform(arg => arg || null),
        }),
    }),
});

export class SessionApiValidator {
    /**
     * Validate and transform the get profile API response
     */
    getProfile = (response: AxiosResponse): Session_GetProfile_Res => {
        const {
            data: { user },
        } = parseApiResponse({
            response,
            schema: GetProfileSchema,
        });

        return {
            data: {
                id: user.id,
                firstName: user.firstName,
                lastName: user.lastName,
                fullName: user.fullName,
                photo: user.photo,
                email: user.email,
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
