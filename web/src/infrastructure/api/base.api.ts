import { type ApiClient, createApiClient } from "@infrastructure/api/client";
import { QueryBuilder } from "@infrastructure/api/queries";

const ac = createApiClient();
const qb = new QueryBuilder();

export abstract class BaseApi {
    protected constructor(client: ApiClient = ac, queryBuilder: QueryBuilder = qb) {
        this.client = client;
        this.queryBuilder = queryBuilder;
    }

    protected readonly client: ApiClient;
    protected readonly queryBuilder: QueryBuilder;
}
