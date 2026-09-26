import { z } from "zod";

/** At most, as the server bounds them. */
export const DOCKER_API_MAX_CONTAINERS = 100;

/** The groups of endpoints beyond the core, as the proxy names them, and what each lets children do. */
export const DOCKER_API_GROUPS = [
    { value: "exec", label: "Exec", description: "Run commands inside its containers." },
    { value: "files", label: "Files", description: "Copy files into and out of its containers." },
    { value: "volumes", label: "Volumes", description: "Create volumes of its own, such as caches." },
    { value: "networks", label: "Networks", description: "Create networks of its own, such as one per CI job." },
    {
        value: "nestedSocket",
        label: "Nested socket",
        description: "Give its containers the same Docker API, for jobs that run docker themselves.",
    },
] as const;

const ValueSchema = z.object({ value: z.string() });

export const AppConfigDockerApiFormSchema = z
    .object({
        enabled: z.boolean(),
        mode: z.enum(["proxy", "host"]),
        images: z.array(ValueSchema),
        sharedDirs: z.array(ValueSchema),
        sharedVolumes: z.array(z.object({ key: z.string(), value: z.string() })),
        envNetwork: z.boolean(),
        allow: z.array(z.string()),
        limits: z.object({
            containers: z.number().int().min(0).max(DOCKER_API_MAX_CONTAINERS).optional(),
            memory: z.string().optional(),
            cpus: z.number().min(0).optional(),
        }),
    })
    .superRefine((values, ctx) => {
        if (values.enabled && values.mode === "proxy" && values.images.length === 0) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["images"],
                message: "Add at least one image, or * for any image",
            });
        }
    });

export type AppConfigDockerApiFormSchemaInput = z.input<typeof AppConfigDockerApiFormSchema>;
export type AppConfigDockerApiFormSchemaOutput = z.output<typeof AppConfigDockerApiFormSchema>;

export const emptyAppConfigDockerApiFormDefaults: AppConfigDockerApiFormSchemaInput = {
    enabled: false,
    mode: "proxy",
    images: [],
    sharedDirs: [],
    sharedVolumes: [],
    envNetwork: false,
    allow: [],
    limits: { containers: undefined, memory: undefined, cpus: undefined },
};
