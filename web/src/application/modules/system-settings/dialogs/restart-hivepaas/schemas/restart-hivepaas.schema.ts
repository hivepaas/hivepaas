import { z } from "zod";

export const RestartHivePaaSFormSchema = z
    .object({
        restartMainApp: z.boolean(),
        restartDbApp: z.boolean(),
        restartCacheApp: z.boolean(),
        restartWorkers: z.boolean(),
        restartAgents: z.boolean(),
    })
    .superRefine((val, ctx) => {
        const hasSelectedService =
            val.restartMainApp || val.restartDbApp || val.restartCacheApp || val.restartWorkers || val.restartAgents;

        if (!hasSelectedService) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                message: "Please select at least one service to restart",
                path: ["root"],
            });
        }
    });

export type RestartHivePaaSFormSchemaInput = z.input<typeof RestartHivePaaSFormSchema>;
export type RestartHivePaaSFormSchemaOutput = z.output<typeof RestartHivePaaSFormSchema>;
