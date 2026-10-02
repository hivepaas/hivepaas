import { z } from "zod";
import { type FunctionSourcePayload } from "~/projects/api/services";
import { type FunctionSource } from "~/projects/domain";
import { FUNCTION_TEMPLATES } from "~/projects/module-shared/constants";
import { ERepoType, isKnownFunctionRuntime } from "~/projects/module-shared/enums";
import { functionFilePathProblem, functionSourceToPayload } from "~/projects/module-shared/utils";

/**
 * Where a function's code is: in the dashboard's editor, or in a repository.
 */
export const EFunctionCodeLocation = {
    Inline: "inline",
    Repository: "repository",
} as const;

export type EFunctionCodeLocation = (typeof EFunctionCodeLocation)[keyof typeof EFunctionCodeLocation];

const SettingsRefSchema = z
    .object({
        id: z.string(),
        name: z.string(),
        type: z.string().optional(),
        kind: z.string().optional(),
    })
    .nullish();

// The same names the backend takes: a Debian package, with an optional version.
const DEBIAN_PACKAGE_PATTERN = /^[a-z0-9][a-z0-9+.-]+(=[A-Za-z0-9.+~:-]+)?$/;

export const FunctionSettingsFormSchema = z
    .object({
        runtime: z.string().min(1, "Runtime is required"),
        entrypointFile: z.string().trim(),
        entrypointHandler: z.string().trim(),
        timeout: z.string().trim().min(1, "Timeout is required"),
        maxConcurrency: z.coerce
            .number({ invalid_type_error: "Concurrency is a number" })
            .int("Concurrency is a whole number")
            .min(1, "At least 1")
            .max(1000, "At most 1000"),
        maxBodySize: z.string().trim().min(1, "Body size is required"),
        systemPackages: z.array(z.string()),
        codeLocation: z.nativeEnum(EFunctionCodeLocation),
        repoUrl: z.string().trim(),
        repoRef: z.string().trim(),
        dir: z.string().trim(),
        credentials: SettingsRefSchema,
        autoDeploy: z.boolean(),
        pushToRegistry: SettingsRefSchema,
    })
    .superRefine((values, ctx) => {
        values.systemPackages.forEach(pkg => {
            if (!DEBIAN_PACKAGE_PATTERN.test(pkg)) {
                ctx.addIssue({
                    code: z.ZodIssueCode.custom,
                    path: ["systemPackages"],
                    message: `"${pkg}" is not a Debian package name`,
                });
            }
        });
        if (values.codeLocation !== EFunctionCodeLocation.Repository) {
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

export type FunctionSettingsFormInput = z.input<typeof FunctionSettingsFormSchema>;
export type FunctionSettingsFormOutput = z.output<typeof FunctionSettingsFormSchema>;

/**
 * The form's values for a function's source.
 */
export function functionSettingsDefaultValues(source: FunctionSource): FunctionSettingsFormInput {
    const { repo } = source.code;

    return {
        runtime: source.runtime,
        entrypointFile: source.entrypoint.file,
        entrypointHandler: source.entrypoint.handler,
        timeout: source.timeout,
        maxConcurrency: source.maxConcurrency,
        maxBodySize: source.maxBodySize,
        systemPackages: source.systemPackages,
        codeLocation: repo ? EFunctionCodeLocation.Repository : EFunctionCodeLocation.Inline,
        repoUrl: repo?.repoUrl ?? "",
        repoRef: repo?.repoRef ?? "",
        dir: repo ? source.code.dir : "",
        credentials: repo?.credentials ?? null,
        // Code moved into a repository deploys on push, as a new function's does.
        autoDeploy: repo?.autoDeploy ?? true,
        pushToRegistry: source.pushToRegistry,
    };
}

/**
 * The source the form's values make of the function's: the code stays where it
 * is unless the form moves it. Code moved into the editor starts from what the
 * function had there, or from the runtime's template.
 */
export function functionSettingsToSource(
    values: FunctionSettingsFormOutput,
    source: FunctionSource,
): FunctionSourcePayload {
    const { runtime } = values;
    const fromRepository = values.codeLocation === EFunctionCodeLocation.Repository;
    const template = isKnownFunctionRuntime(runtime) ? FUNCTION_TEMPLATES[runtime] : [];
    const files = source.code.inline?.files.length ? source.code.inline.files : template;
    const base = functionSourceToPayload(source, fromRepository ? undefined : files);

    return {
        ...base,
        runtime,
        entrypoint: { file: values.entrypointFile, handler: values.entrypointHandler },
        code: fromRepository
            ? {
                  repo: {
                      repoType: ERepoType.Git,
                      repoUrl: values.repoUrl,
                      repoRef: values.repoRef,
                      // A commit pins the ref it was read from; another ref starts afresh.
                      commitHash:
                          source.code.repo?.repoUrl === values.repoUrl && source.code.repo.repoRef === values.repoRef
                              ? source.code.repo.commitHash
                              : "",
                      credentials: { id: values.credentials?.id ?? "" },
                      autoDeploy: values.autoDeploy,
                  },
                  dir: values.dir,
              }
            : base.code,
        systemPackages: values.systemPackages,
        timeout: values.timeout,
        maxConcurrency: values.maxConcurrency,
        maxBodySize: values.maxBodySize,
        pushToRegistry: { id: values.pushToRegistry?.id ?? "" },
    };
}

const FUNCTION_SOURCE_ERROR_FIELDS: Record<string, keyof FunctionSettingsFormInput> = {
    "runtime": "runtime",
    "entrypoint.file": "entrypointFile",
    "entrypoint.handler": "entrypointHandler",
    "timeout": "timeout",
    "maxConcurrency": "maxConcurrency",
    "maxBodySize": "maxBodySize",
    "systemPackages": "systemPackages",
    "code.repo.repoURL": "repoUrl",
    "code.repo.repoRef": "repoRef",
    "code.repo.credentials": "credentials",
    "code.dir": "dir",
    "pushToRegistry": "pushToRegistry",
};

/**
 * The form's field a backend's validation error is about, or null when the
 * error is about something the form does not show.
 */
export function functionSettingsErrorField(path: string): keyof FunctionSettingsFormInput | null {
    const inSource = path.replace(/^functionSource\./, "").replace(/\[\d+\]$/, "");
    return FUNCTION_SOURCE_ERROR_FIELDS[inSource] ?? null;
}
