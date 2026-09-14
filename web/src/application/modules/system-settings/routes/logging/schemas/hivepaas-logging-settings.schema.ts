import { z } from "zod";

/** The one way an endpoint authenticates. Only the selected mode is stored. */
export const LOGGING_AUTH_MODES = ["none", "basic", "bearer"] as const;

const EndpointForm = z.object({
    url: z.string().trim(),
    authMode: z.enum(LOGGING_AUTH_MODES),
    username: z.string().trim(),
    password: z.string(),
    bearerToken: z.string(),
    headers: z.array(z.object({ key: z.string().trim(), value: z.string() })),
    tlsSkipVerify: z.boolean(),
});

type EndpointFormOutput = z.output<typeof EndpointForm>;

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
        // Basic auth with no username sends no credentials at all: the request
        // would go out unauthenticated while the form claims otherwise.
        const requireBasicUsername = (ep: EndpointFormOutput, path: (string | number)[]) => {
            if (ep.authMode === "basic" && !ep.username) {
                ctx.addIssue({ code: "custom", path: [...path, "username"], message: "Required" });
            }
        };
        if (!v.backendManaged) {
            if (!v.ingest.url) {
                ctx.addIssue({ code: "custom", path: ["ingest", "url"], message: "Required" });
            }
            requireBasicUsername(v.ingest, ["ingest"]);
            // The query endpoint is optional, and is only sent when it has a URL.
            if (v.query.url) {
                requireBasicUsername(v.query, ["query"]);
            }
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
            requireBasicUsername(f, ["forwards", i]);
        });
    });

export type HivePaaSLoggingSettingsFormInput = z.input<typeof HivePaaSLoggingSettingsFormSchema>;
export type HivePaaSLoggingSettingsFormOutput = z.output<typeof HivePaaSLoggingSettingsFormSchema>;
