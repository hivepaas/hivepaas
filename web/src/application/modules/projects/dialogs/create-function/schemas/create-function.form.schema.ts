import { z } from "zod";
import { type FunctionSourcePayload } from "~/projects/api/services";
import { functionTemplateOf } from "~/projects/module-shared/constants";
import {
    ALL_FUNCTION_RUNTIMES,
    EFunctionLanguage,
    type EFunctionRuntime,
    ERepoType,
} from "~/projects/module-shared/enums";
import { functionFilePathProblem } from "~/projects/module-shared/utils";

/**
 * Where a new function's code comes from: its runtime's template, edited in
 * the dashboard afterwards, or a repository.
 */
export const EFunctionCodeSource = {
    Template: "template",
    Repository: "repository",
} as const;

export type EFunctionCodeSource = (typeof EFunctionCodeSource)[keyof typeof EFunctionCodeSource];

const CredentialsRefSchema = z
    .object({
        id: z.string(),
        name: z.string(),
        type: z.string().optional(),
        kind: z.string().optional(),
    })
    .nullable();

/** A domain name without a wildcard: labels of letters, digits and '-', at least two. */
const DOMAIN_PATTERN = /^(?=.{1,253}$)([a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$/i;

/** A function's name as a DNS label: lower case, '-' for anything else, at most 63. */
export function functionDomainLabel(name: string): string {
    return name
        .toLowerCase()
        .replace(/[^a-z0-9-]+/g, "-")
        .replace(/^-+/, "")
        .slice(0, 63)
        .replace(/-+$/, "");
}

const CreateFunctionFormSchemaBase = z.object({
    name: z
        .string({
            required_error: "Name is required",
        })
        .trim()
        .min(1, "Name is required"),
    // A function is created in an env: it is deployed as soon as it exists.
    env: z.string().min(1, "Environment is required").max(50, "Environment must be at most 50 characters"),
    runtime: z.enum(ALL_FUNCTION_RUNTIMES as [string, ...string[]]),
    /** For a runtime that takes either language: the template and the entrypoint. */
    language: z.nativeEnum(EFunctionLanguage),
    codeSource: z.nativeEnum(EFunctionCodeSource),
    repoUrl: z.string().trim(),
    repoRef: z.string().trim(),
    dir: z.string().trim(),
    credentials: CredentialsRefSchema,
    /** Whether a push to the ref deploys the function. */
    autoDeploy: z.boolean(),
    /** Whether the function is routed at domain from its first deployment. */
    expose: z.boolean(),
    domain: z.string().trim(),
});

export function createCreateFunctionFormSchema(
    envNames: string[] = [],
): z.ZodEffects<typeof CreateFunctionFormSchemaBase> {
    return CreateFunctionFormSchemaBase.superRefine((values, ctx) => {
        if (values.env !== "" && !envNames.includes(values.env)) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["env"],
                message: "Environment must be one of the project environments",
            });
        }
        if (values.expose && !DOMAIN_PATTERN.test(values.domain)) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["domain"],
                message: "A domain name, such as hello.example.com",
            });
        }
        if (values.codeSource !== EFunctionCodeSource.Repository) {
            return;
        }
        if (values.repoUrl === "") {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["repoUrl"],
                message: "Repository URL is required",
            });
        }
        if (values.autoDeploy && values.repoRef === "") {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["repoRef"],
                message: "Deploying on push needs a branch",
            });
        }
        const dirProblem = values.dir === "" ? null : functionFilePathProblem(values.dir, []);
        if (dirProblem) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["dir"],
                message: dirProblem,
            });
        }
    });
}

export type CreateFunctionFormInput = z.input<ReturnType<typeof createCreateFunctionFormSchema>>;
export type CreateFunctionFormOutput = z.output<ReturnType<typeof createCreateFunctionFormSchema>>;

/**
 * What a function is created from, by the form: the template of its runtime and
 * language, or the repository. What the form leaves out - the entrypoint of
 * JavaScript, the limits - the backend fills in with the runtime's defaults; a
 * TypeScript function's entrypoint is the template's, index.ts, from a
 * repository too.
 */
export function createFunctionSource(values: CreateFunctionFormOutput): FunctionSourcePayload {
    const runtime = values.runtime as EFunctionRuntime;
    const fromRepository = values.codeSource === EFunctionCodeSource.Repository;
    const template = functionTemplateOf(runtime, values.language);

    return {
        runtime,
        contract: "",
        entrypoint: { file: template.entrypoint, handler: "" },
        code: fromRepository
            ? {
                  repo: {
                      repoType: ERepoType.Git,
                      repoUrl: values.repoUrl,
                      repoRef: values.repoRef,
                      commitHash: "",
                      credentials: { id: values.credentials?.id ?? "" },
                      autoDeploy: values.autoDeploy,
                  },
                  dir: values.dir,
              }
            : {
                  inline: { files: template.files },
                  dir: "",
              },
        systemPackages: [],
        timeout: "",
        maxConcurrency: 0,
        maxBodySize: "",
        pushToRegistry: { id: "" },
    };
}
