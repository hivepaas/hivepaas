import { z } from "zod";

const ConfirmAppDangerActionFormBaseSchema = z.object({
    appName: z.string(),
    /** Delete what the apps stored as well. Off unless somebody ticks it. */
    removeStorage: z.boolean(),
    /** Apply the status change to the apps created to serve this one (dependencies, components) as well. On by default. */
    cascade: z.boolean(),
    /** Delete the cluster the app holds, every member with it. Off unless somebody ticks it. */
    deleteCluster: z.boolean(),
});

export type ConfirmAppDangerActionFormInput = z.input<typeof ConfirmAppDangerActionFormBaseSchema>;
export type ConfirmAppDangerActionFormOutput = z.output<typeof ConfirmAppDangerActionFormBaseSchema>;

export function createConfirmAppDangerActionFormSchema(
    expectedAppName: string,
    clusterRequired: boolean = false,
): z.ZodType<ConfirmAppDangerActionFormOutput, z.ZodTypeDef, ConfirmAppDangerActionFormInput> {
    return ConfirmAppDangerActionFormBaseSchema.refine(values => values.appName === expectedAppName, {
        path: ["appName"],
        message: "App name does not match.",
    }).refine(values => !clusterRequired || values.deleteCluster, {
        path: ["deleteCluster"],
        message: "The app holds its cluster: it is deleted only with all of it.",
    });
}
