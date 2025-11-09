import axios, { type AxiosInstance } from "axios";

import { envConfig } from "@config";

import { initAuthInterceptors, initLangInterceptors } from "@infrastructure/api/interceptors";

export interface ApiClient {
    readonly v1: AxiosInstance;
}

export function createApiClient(): ApiClient {
    const client: ApiClient = {
        v1: axios.create({
            baseURL: envConfig.API_URL,
            headers: {
                "Content-Type": "application/json",
            },
        }),
    } as const;

    initLangInterceptors(client.v1);
    initAuthInterceptors(client.v1);

    return client;
}
