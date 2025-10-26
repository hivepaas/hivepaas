import { parseEnv, z } from "znv";

const parsedConfig = parseEnv(import.meta.env, {
    VITE_API_URL: z.string(),
    VITE_PRODUCTION_VIEW: z.coerce.boolean().default(true),
    VITE_LOGOUT_TIMEOUT: z.coerce
        .number()
        .int()
        .min(0)
        .max(24 * 60, "Logout timeout should be less than 24 hours in minutes")
        .default(0),
});

/**
 * Values from environment variables
 */
export const EnvConfig = {
    API_URL: parsedConfig.VITE_API_URL,
    PRODUCTION_VIEW: parsedConfig.VITE_PRODUCTION_VIEW,
    LOGOUT_TIMEOUT: parsedConfig.VITE_LOGOUT_TIMEOUT,
} as const;
