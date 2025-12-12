import { type PaginationState, type SortingState } from "@infrastructure/data";
import { type NodeBase } from "~/cluster/domain";

import { type ApiRequestBase, type ApiResponseBase, type ApiResponsePaginated } from "@infrastructure/api";

/**
 * Find many nodes paginated
 */
export type Nodes_FindManyPaginated_Req = ApiRequestBase<{
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type Nodes_FindManyPaginated_Res = ApiResponsePaginated<NodeBase>;

/**
 * Find one node by id
 */
export type Nodes_FindOneById_Req = ApiRequestBase<{
    id: string;
}>;

export type Nodes_FindOneById_Res = ApiResponseBase<NodeBase>;

/**
 * Delete one node
 */
export type Nodes_DeleteOne_Req = ApiRequestBase<{
    id: string;
}>;

export type Nodes_DeleteOne_Res = ApiResponseBase<{
    id: string;
}>;

/**
 * Update one node
 */
export type Nodes_UpdateOne_Req = ApiRequestBase<{
    id: string;
    node: Partial<Omit<NodeBase, "createdAt" | "updatedAt" | "id">>;
}>;

export type Nodes_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

/**
 * Create a node
 */
export type Nodes_CreateOne_Req = ApiRequestBase<{
    node: Omit<NodeBase, "id" | "createdAt" | "updatedAt">;
}>;

export type Nodes_CreateOne_Res = ApiResponseBase<NodeBase>;
