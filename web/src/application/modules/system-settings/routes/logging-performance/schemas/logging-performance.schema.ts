import { z } from "zod";

export const LoggingPerformanceFormSchema = z.object({
    enabled: z.boolean(),
    /** Every node of the swarm, as listed: those enabled are saved, each with its capacity. */
    nodes: z.array(
        z.object({
            id: z.string(),
            enabled: z.boolean(),
            capacity: z.enum(["auto", "small", "medium", "large"]),
        }),
    ),
});

export type LoggingPerformanceFormInput = z.input<typeof LoggingPerformanceFormSchema>;
export type LoggingPerformanceFormOutput = z.output<typeof LoggingPerformanceFormSchema>;
