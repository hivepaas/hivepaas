import z from "zod";

export const F2aSetupSchema = z.object({
    passcode: z.string().trim().min(1, "Passcode is required"),
});

export type F2aSetupSchemaInput = z.input<typeof F2aSetupSchema>;
export type F2aSetupSchemaOutput = z.output<typeof F2aSetupSchema>;
