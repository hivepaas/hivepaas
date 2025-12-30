import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";
import type {
    NodesApiValidator,
    Nodes_CreateOne_Req,
    Nodes_CreateOne_Res,
    Nodes_DeleteOne_Req,
    Nodes_DeleteOne_Res,
    Nodes_FindManyPaginated_Req,
    Nodes_FindManyPaginated_Res,
    Nodes_FindOneById_Req,
    Nodes_FindOneById_Res,
    Nodes_GetJoinNode_Req,
    Nodes_GetJoinNode_Res,
    Nodes_JoinNode_Req,
    Nodes_JoinNode_Res,
    Nodes_UpdateOne_Req,
    Nodes_UpdateOne_Res,
} from "~/cluster/api/services/nodes-services";

import { BaseApi, JsonTransformer, parseApiError } from "@infrastructure/api";

export class NodesApi extends BaseApi {
    public constructor(private readonly validator: NodesApiValidator) {
        super();
    }

    /**
     * Find many nodes paginated
     */
    async findManyPaginated(
        request: Nodes_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<Nodes_FindManyPaginated_Res, Error>> {
        const { search, pagination, sorting } = request.data;

        const query = this.queryBuilder.getInstance();

        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get("/cluster/nodes", {
                    params: query.build(),
                    signal,
                }),
            ).pipe(
                map(this.validator.findManyPaginated),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Find one node by id
     */
    async findOneById(
        request: Nodes_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<Nodes_FindOneById_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/cluster/nodes/${id}`, {
                    signal,
                }),
            ).pipe(
                map(this.validator.findOneById),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Delete a node
     */
    async deleteOne(request: Nodes_DeleteOne_Req): Promise<Result<Nodes_DeleteOne_Res, Error>> {
        const { id, force } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.delete(`/cluster/nodes/${id}`, {
                    params: {
                        force,
                    },
                }),
            ).pipe(
                map(() => Ok({ data: { id } })),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Update a node
     */
    async updateOne(request: Nodes_UpdateOne_Req, signal?: AbortSignal): Promise<Result<Nodes_UpdateOne_Res, Error>> {
        const { id, name, role, availability, updateVer, labels } = request.data;

        const json = {
            name: JsonTransformer.string({
                data: name,
            }),
            role,
            availability,
            updateVer,
            labels: labels.reduce<Record<string, string>>(
                (acc, current) => ({
                    ...acc,
                    [current.key]: current.value,
                }),
                {},
            ),
        };

        return lastValueFrom(
            from(
                this.client.v1.put(`/cluster/nodes/${id}`, json, {
                    signal,
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } as const })),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Create a node
     */
    async createOne(request: Nodes_CreateOne_Req, signal?: AbortSignal): Promise<Result<Nodes_CreateOne_Res, Error>> {
        const { node } = request.data;

        const json = {
            name: JsonTransformer.string({
                data: node.name,
            }),
        };

        return lastValueFrom(
            from(this.client.v1.post("/cluster/nodes", json, { signal })).pipe(
                map(this.validator.createOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Get join node command
     */
    async getJoinNode(
        request: Nodes_GetJoinNode_Req,
        signal?: AbortSignal,
    ): Promise<Result<Nodes_GetJoinNode_Res, Error>> {
        const { joinAsManager } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get("/cluster/nodes/join-command", {
                    params: {
                        joinAsManager,
                    },
                    signal,
                }),
            ).pipe(
                map(this.validator.getJoinNode),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
    /**
     * Join node
     */
    async joinNode(request: Nodes_JoinNode_Req, signal?: AbortSignal): Promise<Result<Nodes_JoinNode_Res, Error>> {
        const { sshKey, host, port, user, joinAsManager } = request.data;
        const json = {
            sshKey,
            host,
            port,
            user,
            joinAsManager,
        };

        return lastValueFrom(
            from(this.client.v1.post("/cluster/nodes/join", json, { signal })).pipe(
                map(() => Ok({ data: { type: "success" } as const })),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
