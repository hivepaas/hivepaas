import { z } from "zod";

export const TraefikConfigOptionsFormSchema = z.object({
    startupCommand: z.object({
        logLevel: z.string(),
        accessLog: z.boolean(),
        http3: z.boolean(),
        fastProxy: z.boolean(),
        openPortsText: z.string().optional(),
        argsText: z.string(),
    }),
});

export type TraefikConfigOptionsFormInput = z.input<typeof TraefikConfigOptionsFormSchema>;
export type TraefikConfigOptionsFormOutput = z.output<typeof TraefikConfigOptionsFormSchema>;

export const emptyTraefikConfigOptionsFormDefaults: TraefikConfigOptionsFormInput = {
    startupCommand: {
        logLevel: "",
        accessLog: false,
        http3: false,
        fastProxy: false,
        openPortsText: "",
        argsText: "",
    },
};
