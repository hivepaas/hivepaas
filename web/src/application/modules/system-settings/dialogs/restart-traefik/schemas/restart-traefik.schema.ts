import { z } from "zod";

export const RestartTraefikFormSchema = z
    .object({
        restartTraefikService: z.boolean(),
    })
    .superRefine((val, ctx) => {
        if (!val.restartTraefikService) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                message: "Please select Traefik service to restart",
                path: ["root"],
            });
        }
    });

export type RestartTraefikFormSchemaInput = z.input<typeof RestartTraefikFormSchema>;
export type RestartTraefikFormSchemaOutput = z.output<typeof RestartTraefikFormSchema>;
