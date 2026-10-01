import { z } from "zod";
import { EAppScheduledJobScheduleMode, EAppScheduledJobTaskPriority } from "~/projects/module-shared/enums";

/** The methods a function's call sends, as the server takes them. */
export const FUNCTION_INVOKE_METHODS = ["GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"] as const;

/** A method that sends no body. */
export function methodSendsNoBody(method: string): boolean {
    return method === "GET" || method === "HEAD";
}

/** The server's limits: a test run's body, a path with its query, the headers. */
const BODY_MAX_BYTES = 1024 * 1024;
const PATH_MAX_LENGTH = 2048;
const HEADERS_MAX = 50;
/** An HTTP header's name: a token. */
const HEADER_NAME_PATTERN = /^[A-Za-z0-9!#$%&'*+.^_`|~-]+$/;

const NamedRefSchema = z.object({
    id: z.string(),
    name: z.string(),
});

/** A path with its query, without a space or a control character. */
function isPlainPath(path: string): boolean {
    if (/\s/.test(path)) {
        return false;
    }
    for (let i = 0; i < path.length; i++) {
        const code = path.charCodeAt(i);
        if (code < 0x20 || code === 0x7f) {
            return false;
        }
    }
    return true;
}

/** Why header lines cannot be sent, or null when they can. */
export function headerLinesProblem(headerLines: string): string | null {
    const lines = headerLines.split("\n").filter(line => line.trim() !== "");
    const names = new Set<string>();
    for (const line of lines) {
        const separator = line.indexOf(":");
        const name = separator > 0 ? line.slice(0, separator).trim() : "";
        if (!HEADER_NAME_PATTERN.test(name)) {
            return `"${line.trim()}" is not "name: value"`;
        }
        names.add(name.toLowerCase());
    }
    if (names.size > HEADERS_MAX) {
        return `At most ${HEADERS_MAX} headers`;
    }
    return null;
}

export const FunctionInvokeFormSchema = z
    .object({
        name: z.string().trim().min(1, "Name is required"),
        scheduleMode: z.nativeEnum(EAppScheduledJobScheduleMode),
        scheduleInterval: z.string().trim(),
        scheduleCronExpr: z.string().trim(),
        scheduleFrom: z.date().nullable(),
        scheduleTo: z.date().nullable(),
        method: z.enum(FUNCTION_INVOKE_METHODS),
        path: z.string().trim(),
        /** One per line, as "name: value". */
        headers: z.string(),
        body: z.string(),
        timeout: z.string().trim(),
        maxRetry: z.number().int().min(0, "Max retry must be greater than or equal to 0").optional(),
        retryDelay: z.string().trim(),
        priority: z.nativeEnum(EAppScheduledJobTaskPriority),
        controlEnabled: z.boolean(),
        notification: z.object({
            successUseDefault: z.boolean(),
            success: NamedRefSchema.optional(),
            failureUseDefault: z.boolean(),
            failure: NamedRefSchema.optional(),
        }),
    })
    .superRefine((value, ctx) => {
        const issue = (message: string, path: string[]) => {
            ctx.addIssue({ code: z.ZodIssueCode.custom, message, path });
        };

        if (value.path.length > PATH_MAX_LENGTH || !isPlainPath(value.path)) {
            issue(`A path with its query, without spaces, up to ${PATH_MAX_LENGTH} characters`, ["path"]);
        }
        const headersProblem = headerLinesProblem(value.headers);
        if (headersProblem) {
            issue(headersProblem, ["headers"]);
        }
        if (!methodSendsNoBody(value.method) && new TextEncoder().encode(value.body).length > BODY_MAX_BYTES) {
            issue("At most 1 MB", ["body"]);
        }
    });

export type FunctionInvokeFormInput = z.input<typeof FunctionInvokeFormSchema>;
export type FunctionInvokeFormOutput = z.output<typeof FunctionInvokeFormSchema>;
