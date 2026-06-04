import { z } from "zod";

const ConfirmAppDangerActionFormBaseSchema = z.object({
    appName: z.string(),
});

export type ConfirmAppDangerActionFormInput = z.input<typeof ConfirmAppDangerActionFormBaseSchema>;
export type ConfirmAppDangerActionFormOutput = z.output<typeof ConfirmAppDangerActionFormBaseSchema>;

export function createConfirmAppDangerActionFormSchema(
    expectedAppName: string,
): z.ZodType<ConfirmAppDangerActionFormOutput, z.ZodTypeDef, ConfirmAppDangerActionFormInput> {
    return ConfirmAppDangerActionFormBaseSchema.refine(values => values.appName === expectedAppName, {
        path: ["appName"],
        message: "App name does not match.",
    });
}
