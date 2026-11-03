import { z } from "zod";
import { ENodeAvailability } from "~/cluster/module-shared/enums";

export const SingleNodeFormSchema = z.object({
    name: z.string().trim().min(1, "Name is required"),
    availability: z.nativeEnum(ENodeAvailability),
    labels: z.array(
        z.object({
            key: z.string().trim().min(1, "Label name is required"),
            value: z.string(),
        }),
    ),
});

export type SingleNodeFormSchemaInput = z.input<typeof SingleNodeFormSchema>;
export type SingleNodeFormSchemaOutput = z.output<typeof SingleNodeFormSchema>;
