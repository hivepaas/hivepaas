import { z } from "zod";

export const TraefikGeneralFormSchema = z.object({
    appSettings: z.object({
        replicas: z.number().int().min(1).max(100),
    }),
});

export type TraefikGeneralFormInput = z.input<typeof TraefikGeneralFormSchema>;
export type TraefikGeneralFormOutput = z.output<typeof TraefikGeneralFormSchema>;

export const emptyTraefikGeneralFormDefaults: TraefikGeneralFormInput = {
    appSettings: {
        replicas: 1,
    },
};
