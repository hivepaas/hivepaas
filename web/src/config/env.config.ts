import { parseEnv, z } from "znv";

const parsedConfig = parseEnv(import.meta.env, {
    VITE_HP_DASHBOARD_BASE_URL: z.string().optional().default(""),
    VITE_HP_API_BASE_PATH: z.string().optional().default("/_"),
});

function stripTrailingSlash(value: string): string {
    return value.replace(/\/+$/, "");
}

function ensureLeadingSlash(value: string): string {
    return value.startsWith("/") ? value : `/${value}`;
}

function resolveDashboardBaseUrl(envValue: string): string {
    const trimmed = envValue.trim();

    if (trimmed === "") {
        return window.location.origin;
    }

    return stripTrailingSlash(trimmed);
}

function joinApiBaseUrl(dashboardBaseUrl: string, apiBasePath: string): string {
    return `${stripTrailingSlash(dashboardBaseUrl)}${ensureLeadingSlash(apiBasePath.trim() || "/_")}`;
}

const dashboardBaseUrl = resolveDashboardBaseUrl(parsedConfig.VITE_HP_DASHBOARD_BASE_URL);
const apiBasePath = ensureLeadingSlash(parsedConfig.VITE_HP_API_BASE_PATH.trim() || "/_");

/**
 * Values from environment variables, resolved once at boot
 */
export const EnvConfig = {
    DASHBOARD_BASE_URL: dashboardBaseUrl,
    API_BASE_PATH: apiBasePath,
    API_URL: joinApiBaseUrl(dashboardBaseUrl, apiBasePath),
} as const;
