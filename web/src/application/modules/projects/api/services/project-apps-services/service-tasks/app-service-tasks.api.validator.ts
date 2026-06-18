import type { AxiosResponse } from "axios";
import { z } from "zod";
import type { AppServiceTasks_FindMany_Res } from "~/projects/api/services";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

const NullableDateSchema = z.preprocess(
    value => (value === null || value === undefined || value === "" ? null : value),
    z.coerce.date().nullable(),
);

const OptionalStringSchema = z
    .string()
    .nullish()
    .transform(value => value ?? "");

const ContainerStatusSchema = z
    .object({
        containerId: OptionalStringSchema,
    })
    .nullish()
    .transform(value => value ?? null);

const ServiceTaskNodeSchema = z
    .object({
        id: OptionalStringSchema,
        name: OptionalStringSchema,
        hostname: OptionalStringSchema,
        addr: OptionalStringSchema,
        role: OptionalStringSchema,
        isLeader: z
            .boolean()
            .nullish()
            .transform(value => value ?? false),
    })
    .nullish()
    .transform(value => value ?? null);

const ServiceTaskStatusSchema = z
    .object({
        timestamp: NullableDateSchema,
        state: OptionalStringSchema,
        message: OptionalStringSchema,
        err: OptionalStringSchema,
        containerStatus: ContainerStatusSchema,
    })
    .nullish()
    .transform(value => value ?? null);

const ServiceTaskSchema = z.object({
    id: z.string(),
    slot: z.number(),
    node: ServiceTaskNodeSchema,
    status: ServiceTaskStatusSchema,
});

const FindManySchema = z.object({
    data: z.array(ServiceTaskSchema),
    meta: BaseMetaApiSchema.nullish().transform(value => value ?? null),
});

export class AppServiceTasksApiValidator {
    findMany = (response: AxiosResponse): AppServiceTasks_FindMany_Res => {
        return parseApiResponse({
            response,
            schema: FindManySchema,
        });
    };
}
