import { type AxiosResponse } from "axios";
import { z } from "zod";
import {
    type Nodes_CreateOne_Res,
    type Nodes_FindManyPaginated_Res,
    type Nodes_FindOneById_Res,
} from "~/cluster/api/services/nodes-services/nodes/nodes.api.contracts";

import { PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

/**
 * Node schema
 */
const NodeSchema = z.object({
    id: z.string(),
    name: z.string(),
    description: z.string().nullable(),
    status: z.enum(["active", "inactive", "maintenance"]),
    createdAt: z.coerce.date(),
    updatedAt: z.coerce.date().nullable(),
    nodeCount: z.number(),
    region: z.string(),
    version: z.string(),
});

/**
 * Find many nodes paginated API response schema
 */
const FindManyPaginatedSchema = z.object({
    data: z.array(NodeSchema),
    meta: PagingMetaApiSchema,
});

/**
 * Find one node by id API response schema
 */
const FindOneByIdSchema = z.object({
    data: NodeSchema,
});

/**
 * Create one node API response schema
 */
const CreateOneSchema = z.object({
    data: NodeSchema,
});

export class NodesApiValidator {
    /**
     * Validate and transform find many nodes paginated API response
     */
    findManyPaginated = (response: AxiosResponse): Nodes_FindManyPaginated_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManyPaginatedSchema,
        });

        return {
            data: data.map(node => ({
                id: node.id,
                name: node.name,
                description: node.description,
                status: node.status,
                createdAt: node.createdAt,
                updatedAt: node.updatedAt,
                nodeCount: node.nodeCount,
                region: node.region,
                version: node.version,
            })),
            meta,
        };
    };

    /**
     * Validate and transform find one node by id API response
     */
    findOneById = (response: AxiosResponse): Nodes_FindOneById_Res => {
        const { data } = parseApiResponse({
            response,
            schema: FindOneByIdSchema,
        });

        return {
            data: {
                id: data.id,
                name: data.name,
                description: data.description,
                status: data.status,
                createdAt: data.createdAt,
                updatedAt: data.updatedAt,
                nodeCount: data.nodeCount,
                region: data.region,
                version: data.version,
            },
        };
    };

    /**
     * Validate and transform create one node API response
     */
    createOne = (response: AxiosResponse): Nodes_CreateOne_Res => {
        const { data } = parseApiResponse({
            response,
            schema: CreateOneSchema,
        });

        return {
            data: {
                id: data.id,
                name: data.name,
                description: data.description,
                status: data.status,
                createdAt: data.createdAt,
                updatedAt: data.updatedAt,
                nodeCount: data.nodeCount,
                region: data.region,
                version: data.version,
            },
        };
    };
}
