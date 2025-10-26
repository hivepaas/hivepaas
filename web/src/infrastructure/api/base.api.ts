import { type ApiClient, createApiClient } from "@infrastructure/api/client";

const apiClient = createApiClient();

export abstract class BaseApi {
    protected constructor(client: ApiClient = apiClient) {
        this.client = client;
    }

    protected readonly client: ApiClient;
}
