import { z } from "zod";

const ConfirmProjectDangerActionFormBaseSchema = z.object({
    projectName: z.string(),
    /** Delete what the apps stored as well. Off unless somebody ticks it. */
    removeStorage: z.boolean(),
});

export type ConfirmProjectDangerActionFormInput = z.input<typeof ConfirmProjectDangerActionFormBaseSchema>;
export type ConfirmProjectDangerActionFormOutput = z.output<typeof ConfirmProjectDangerActionFormBaseSchema>;

export function createConfirmProjectDangerActionFormSchema(
    expectedProjectName: string,
): z.ZodType<ConfirmProjectDangerActionFormOutput, z.ZodTypeDef, ConfirmProjectDangerActionFormInput> {
    return ConfirmProjectDangerActionFormBaseSchema.refine(values => values.projectName === expectedProjectName, {
        path: ["projectName"],
        message: "Project name does not match.",
    });
}
