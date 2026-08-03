import { z } from "zod";

import { PROXY_PROVIDER_UNSPECIFIED } from "../hivepaas-general.constants";

const appReplicasSchema = z.number().int().min(1).max(100);
const workerReplicasSchema = z.number().int().min(0).max(100);
const workerConcurrencySchema = z.number().int().min(1).max(100);
const durationSchema = z.string().trim().min(1);

const proxyProviderSchema = z.enum(["", "cloudflare", "fastly", "aws-cloudfront", "imperva"]);

export const HivePaaSGeneralFormSchema = z
    .object({
        appSettings: z.object({
            replicas: appReplicasSchema,
        }),
        workerSettings: z.object({
            replicas: workerReplicasSchema,
            concurrency: workerConcurrencySchema,
            runWorkerInMainApp: z.boolean(),
        }),
        taskSettings: z.object({
            taskCheckInterval: durationSchema,
            taskCreateInterval: durationSchema,
        }),
        periodicSettings: z.object({
            baseInterval: durationSchema,
        }),
        proxySettings: z.object({
            proxyProvider: proxyProviderSchema,
            trustedIPsText: z.string(),
        }),
    })
    .superRefine((values, ctx) => {
        if (values.workerSettings.replicas === 0 && !values.workerSettings.runWorkerInMainApp) {
            ctx.addIssue({
                code: "custom",
                message: "Run Worker in Main App must be enabled when worker replicas is 0",
                path: ["workerSettings", "runWorkerInMainApp"],
            });
        }
    });

export type HivePaaSGeneralFormInput = z.input<typeof HivePaaSGeneralFormSchema>;
export type HivePaaSGeneralFormOutput = z.output<typeof HivePaaSGeneralFormSchema>;

export const emptyHivePaaSGeneralFormDefaults: HivePaaSGeneralFormInput = {
    appSettings: {
        replicas: 1,
    },
    workerSettings: {
        replicas: 1,
        concurrency: 1,
        runWorkerInMainApp: true,
    },
    taskSettings: {
        taskCheckInterval: "10m",
        taskCreateInterval: "10m",
    },
    periodicSettings: {
        baseInterval: "15s",
    },
    proxySettings: {
        proxyProvider: PROXY_PROVIDER_UNSPECIFIED,
        trustedIPsText: "",
    },
};
