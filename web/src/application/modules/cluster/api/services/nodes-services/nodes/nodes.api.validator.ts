import { type AxiosResponse } from "axios";
import { z } from "zod";
import {
    type Nodes_CreateOne_Res,
    type Nodes_FindManyPaginated_Res,
    type Nodes_FindOneById_Res,
} from "~/cluster/api/services/nodes-services/nodes/nodes.api.contracts";
import { ENodeAvailability, ENodeRole, ENodeStatus } from "~/cluster/module-shared/enums";

import { PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

/**
 * Node schema
 */
const NodeSchema = z.object({
    id: z.string(),
    name: z.string(),
    labels: z.record(z.string(), z.string()).nullable(),
    hostname: z.string(),
    addr: z.string(),
    status: z.nativeEnum(ENodeStatus),
    availability: z.nativeEnum(ENodeAvailability),
    role: z.nativeEnum(ENodeRole),
    isLeader: z.boolean(),
    platform: z
        .object({
            architecture: z.string(),
            os: z.string(),
        })
        .nullable(),
    resources: z
        .object({
            cpus: z.number(),
            memoryMB: z.number(),
        })
        .nullable(),
    engineDesc: z
        .object({
            engineVersion: z.string(),
            labels: z.record(z.string(), z.string()).nullable(),
            plugins: z.array(z.object({ name: z.string(), type: z.string() })).optional(),
        })
        .nullable(),
    updateVer: z.number(),
    createdAt: z.coerce.date(),
    updatedAt: z.coerce.date().nullable(),
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
                labels: node.labels,
                hostname: node.hostname,
                addr: node.addr,
                availability: node.availability,
                status: node.status,
                role: node.role,
                isLeader: node.isLeader,
                platform: node.platform,
                resources: node.resources,
                engineDesc: node.engineDesc,
                updateVer: node.updateVer,
                createdAt: node.createdAt,
                updatedAt: node.updatedAt,
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
                labels: data.labels,
                hostname: data.hostname,
                addr: data.addr,
                availability: data.availability,
                status: data.status,
                role: data.role,
                isLeader: data.isLeader,
                platform: data.platform,
                resources: data.resources,
                engineDesc: data.engineDesc,
                updateVer: data.updateVer,
                createdAt: data.createdAt,
                updatedAt: data.updatedAt,
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
            },
        };
    };
}
