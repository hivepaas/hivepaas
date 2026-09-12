import { z } from "zod";

const EndpointForm = z.object({
    url: z.string().trim(),
    username: z.string().trim(),
    password: z.string(),
    bearerToken: z.string(),
    tlsSkipVerify: z.boolean(),
});

export const HivePaaSLoggingSettingsFormSchema = z
    .object({
        enabled: z.boolean(),
        sources: z.object({ apps: z.boolean(), hivepaas: z.boolean(), traefikAccess: z.boolean(), nodes: z.boolean() }),
        backendManaged: z.boolean(),
        nodeId: z.string(),
        volumeId: z.string(),
        volumeSubpath: z
            .string()
            .trim()
            // It is a path inside the volume: the server refuses anything that
            // would climb out, and saying so here costs a round trip less.
            .refine(v => !v.startsWith("/"), "Must not start with /")
            .refine(v => !v.split("/").includes(".."), "Must not contain .."),
        retention: z
            .string()
            .trim()
            .regex(/^(\d+(w|d|h|m|s))+$/, "Use a duration such as 30d or 12h"),
        maxDiskUsagePercent: z.number().int().min(1).max(100).nullable(),
        ingest: EndpointForm,
        query: EndpointForm,
        forwards: z.array(EndpointForm.extend({ name: z.string().trim().min(1, "Required"), format: z.string() })),
    })
    .superRefine((v, ctx) => {
        // Nothing is required while logging is off: that is the default, and
        // turning the feature off must not demand a complete form.
        if (!v.enabled) {
            return;
        }
        if (!v.sources.apps && !v.sources.hivepaas) {
            ctx.addIssue({ code: "custom", path: ["sources", "apps"], message: "Select at least one source" });
        }
        if (v.backendManaged && !v.nodeId) {
            ctx.addIssue({ code: "custom", path: ["nodeId"], message: "Required" });
        }
        if (v.backendManaged && !v.volumeId) {
            ctx.addIssue({ code: "custom", path: ["volumeId"], message: "Required" });
        }
        if (!v.backendManaged && !v.ingest.url) {
            ctx.addIssue({ code: "custom", path: ["ingest", "url"], message: "Required" });
        }
        const seen = new Set<string>();
        v.forwards.forEach((f, i) => {
            if (seen.has(f.name)) {
                ctx.addIssue({ code: "custom", path: ["forwards", i, "name"], message: "Names must be unique" });
            }
            seen.add(f.name);
            if (!f.url) {
                ctx.addIssue({ code: "custom", path: ["forwards", i, "url"], message: "Required" });
            }
        });
    });

export type HivePaaSLoggingSettingsFormInput = z.input<typeof HivePaaSLoggingSettingsFormSchema>;
export type HivePaaSLoggingSettingsFormOutput = z.output<typeof HivePaaSLoggingSettingsFormSchema>;
