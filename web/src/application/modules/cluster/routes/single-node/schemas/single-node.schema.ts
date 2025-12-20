import { z } from "zod";
import { ENodeAvailability, ENodeRole } from "~/cluster/module-shared/enums";

export const SingleNodeFormSchema = z.object({
    name: z.string().min(1, "Name is required"),
    role: z.nativeEnum(ENodeRole),
    availability: z.nativeEnum(ENodeAvailability),
    labels: z.array(
        z.object({
            key: z.string().min(1, "Label name is required"),
            value: z.string().min(1, "Value is required"),
        }),
    ),
});

export type SingleNodeFormSchemaInput = z.input<typeof SingleNodeFormSchema>;
export type SingleNodeFormSchemaOutput = z.output<typeof SingleNodeFormSchema>;
