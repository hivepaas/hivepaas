import { type AxiosResponse } from "axios";
import { z } from "zod";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import {
    type AppResourceSettings_FindOne_Res,
    type AppResourceSettings_UpdateOne_Res,
} from "./app-resource-settings.api.contracts";

const GenericResourceSchema = z.object({
    kind: z.string(),
    value: z.string(),
});

const ResourceReservationsSchema = z.object({
    cpus: z.number().optional(),
    memoryMB: z.number().optional(),
    genericResources: z.array(GenericResourceSchema).nullish(),
});

const ResourceLimitsSchema = z.object({
    cpus: z.number().optional(),
    memoryMB: z.number().optional(),
    pids: z.number().optional(),
});

// BE Ulimit struct has no json tags, so fields are PascalCase
const UlimitSchema = z.object({
    Name: z.string(),
    Hard: z.number(),
    Soft: z.number(),
});

const CapabilitiesSchema = z.object({
    capabilityAdd: z.array(z.string()).nullish(),
    capabilityDrop: z.array(z.string()).nullish(),
    enableGPU: z.boolean().optional(),
    oomScoreAdj: z.number().optional(),
    sysctls: z.record(z.string(), z.string()).nullish(),
});

const AppResourceSettingsSchema = z.object({
    reservations: ResourceReservationsSchema.nullish(),
    limits: ResourceLimitsSchema.nullish(),
    ulimits: z.array(UlimitSchema).nullish(),
    capabilities: CapabilitiesSchema.nullish(),
    updateVer: z.number(),
});

const FindOneSchema = z.object({
    data: AppResourceSettingsSchema,
    meta: BaseMetaApiSchema.nullable(),
});

const UpdateOneSchema = z.object({
    data: z.object({ type: z.literal("success") }),
    meta: BaseMetaApiSchema.nullable(),
});

export class AppResourceSettingsApiValidator {
    findOne = (response: AxiosResponse): AppResourceSettings_FindOne_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneSchema });
        return {
            data: {
                reservations: data.reservations
                    ? {
                          cpus: data.reservations.cpus,
                          memoryMB: data.reservations.memoryMB,
                          genericResources:
                              data.reservations.genericResources?.map(item => ({
                                  kind: item.kind,
                                  value: item.value,
                              })) ?? [],
                      }
                    : null,
                limits: data.limits
                    ? {
                          cpus: data.limits.cpus,
                          memoryMB: data.limits.memoryMB,
                          pids: data.limits.pids,
                      }
                    : null,
                ulimits:
                    data.ulimits?.map(item => ({
                        name: item.Name,
                        hard: item.Hard,
                        soft: item.Soft,
                    })) ?? [],
                capabilities: data.capabilities
                    ? {
                          capabilityAdd: data.capabilities.capabilityAdd ?? [],
                          capabilityDrop: data.capabilities.capabilityDrop ?? [],
                          enableGPU: data.capabilities.enableGPU,
                          oomScoreAdj: data.capabilities.oomScoreAdj,
                          sysctls: data.capabilities.sysctls ?? {},
                      }
                    : null,
                updateVer: data.updateVer,
            },
            meta,
        };
    };

    updateOne = (response: AxiosResponse): AppResourceSettings_UpdateOne_Res => {
        return parseApiResponse({ response, schema: UpdateOneSchema });
    };
}
