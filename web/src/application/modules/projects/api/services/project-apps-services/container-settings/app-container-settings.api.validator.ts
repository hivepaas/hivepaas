import { type AxiosResponse } from "axios";
import { z } from "zod";
import { EAppArmorMode, EHealthcheckMode, ERestartPolicyCondition, ESeccompMode } from "~/projects/module-shared/enums";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import {
    type AppContainerSettings_CheckPort_Res,
    type AppContainerSettings_FindOne_Res,
} from "./app-container-settings.api.contracts";

const SELinuxContextSchema = z.object({
    disable: z.boolean().optional(),
    user: z.string().optional(),
    role: z.string().optional(),
    type: z.string().optional(),
    level: z.string().optional(),
});

const SeccompOptsSchema = z.object({
    mode: z.nativeEnum(ESeccompMode).optional(),
    profile: z.string().optional(),
});

const AppArmorOptsSchema = z.object({
    mode: z.nativeEnum(EAppArmorMode).optional(),
});

const PrivilegesSchema = z.object({
    seLinuxContext: SELinuxContextSchema.nullish(),
    seccomp: SeccompOptsSchema.nullish(),
    appArmor: AppArmorOptsSchema.nullish(),
    noNewPrivileges: z.boolean().optional(),
});

const RestartPolicySchema = z.object({
    condition: z.nativeEnum(ERestartPolicyCondition).optional(),
    delay: z.union([z.string(), z.null()]).optional(),
    maxAttempts: z.union([z.number(), z.null()]).optional(),
    window: z.union([z.string(), z.null()]).optional(),
});

const HealthcheckSchema = z.object({
    enabled: z.boolean().optional(),
    mode: z.nativeEnum(EHealthcheckMode).optional(),
    command: z.string().optional(),
    interval: z.string().optional(),
    timeout: z.string().optional(),
    startPeriod: z.string().optional(),
    startInterval: z.string().optional(),
    retries: z.number().optional(),
});

const LogDriverSchema = z.object({
    name: z.string(),
    options: z.record(z.string(), z.string()).nullish(),
});

const ContainerSpecSchema = z.object({
    serviceLabels: z.record(z.string(), z.string()).nullish(),
    containerLabels: z.record(z.string(), z.string()).nullish(),
    image: z.string(),
    command: z.string(),
    workingDir: z.string(),
    hostname: z.string(),
    user: z.string(),
    groups: z.array(z.string()).nullish(),
    stopSignal: z.string(),
    tty: z.boolean(),
    openStdin: z.boolean(),
    readOnly: z.boolean(),
    stopGracePeriod: z.union([z.string(), z.null()]),
    privileges: PrivilegesSchema.nullable(),
    healthcheck: HealthcheckSchema.nullable(),
    restartPolicy: RestartPolicySchema.nullable(),
    logDriver: LogDriverSchema.nullish(),
});

const AppContainerSettingsSchema = ContainerSpecSchema.extend({
    updateVer: z.number(),
});

const FindOneSchema = z.object({
    data: AppContainerSettingsSchema,
    meta: BaseMetaApiSchema.nullable(),
});

const CheckPortSchema = z.object({
    data: z.object({
        open: z.boolean(),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

export class AppContainerSettingsApiValidator {
    findOne = (response: AxiosResponse): AppContainerSettings_FindOne_Res => {
        const data = parseApiResponse({ response, schema: FindOneSchema });
        return {
            data: {
                ...data.data,
                serviceLabels: data.data.serviceLabels ?? {},
                containerLabels: data.data.containerLabels ?? {},
                groups: data.data.groups ?? [],
                logDriver: data.data.logDriver
                    ? { ...data.data.logDriver, options: data.data.logDriver.options ?? {} }
                    : null,
            },
            meta: data.meta,
        };
    };

    checkPort = (response: AxiosResponse): AppContainerSettings_CheckPort_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: CheckPortSchema,
        });

        return { data, meta };
    };
}
