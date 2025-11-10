import { parseEnv, z } from "znv";

const parsedConfig = parseEnv(import.meta.env, {
    VITE_API_URL: z.string(),
    VITE_APP_URL: z.string().optional().default(""),
});

/**
 * Values from environment variables
 */
export const EnvConfig = {
    API_URL: parsedConfig.VITE_API_URL,
    APP_URL: parsedConfig.VITE_APP_URL,
} as const;
