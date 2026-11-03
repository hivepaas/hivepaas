import { z } from "zod";

export const CreateOrEditPositionFormSchema = z.object({
    name: z.string().trim().min(1, "Position is required").max(255, "Position must be less than 255 characters"),
});

export type CreateOrEditPositionFormInput = z.input<typeof CreateOrEditPositionFormSchema>;
export type CreateOrEditPositionFormOutput = z.output<typeof CreateOrEditPositionFormSchema>;
