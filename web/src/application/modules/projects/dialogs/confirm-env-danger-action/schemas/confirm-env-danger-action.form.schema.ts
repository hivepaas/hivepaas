import { z } from "zod";

const ConfirmEnvDangerActionFormBaseSchema = z.object({
    envName: z.string(),
});

export type ConfirmEnvDangerActionFormInput = z.input<typeof ConfirmEnvDangerActionFormBaseSchema>;
export type ConfirmEnvDangerActionFormOutput = z.output<typeof ConfirmEnvDangerActionFormBaseSchema>;

export function createConfirmEnvDangerActionFormSchema(
    expectedEnvName: string,
): z.ZodType<ConfirmEnvDangerActionFormOutput, z.ZodTypeDef, ConfirmEnvDangerActionFormInput> {
    return ConfirmEnvDangerActionFormBaseSchema.refine(values => values.envName === expectedEnvName, {
        path: ["envName"],
        message: "Environment name does not match.",
    });
}
