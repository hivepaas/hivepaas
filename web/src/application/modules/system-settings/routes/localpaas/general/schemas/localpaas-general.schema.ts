import { z } from "zod";

const appReplicasSchema = z.number().int().min(1).max(100);
const workerReplicasSchema = z.number().int().min(0).max(100);
const workerConcurrencySchema = z.number().int().min(1).max(100);
const durationSchema = z.string().trim().min(1);

export const LocalPaaSGeneralFormSchema = z
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
        healthcheckSettings: z.object({
            baseInterval: durationSchema,
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

export type LocalPaaSGeneralFormInput = z.input<typeof LocalPaaSGeneralFormSchema>;
export type LocalPaaSGeneralFormOutput = z.output<typeof LocalPaaSGeneralFormSchema>;

export const emptyLocalPaaSGeneralFormDefaults: LocalPaaSGeneralFormInput = {
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
    healthcheckSettings: {
        baseInterval: "15s",
    },
};
