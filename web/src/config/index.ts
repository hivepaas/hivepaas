export const envConfig = {
    API_URL: (import.meta.env.VITE_API_URL as string | undefined) ?? "",
} as const;

export * from "./env.config";
export * from "./project.config";
