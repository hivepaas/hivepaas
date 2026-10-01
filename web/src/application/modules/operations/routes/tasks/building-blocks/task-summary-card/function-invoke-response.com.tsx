import { Badge } from "@components/ui/badge";
import type { SystemTaskFunctionInvoke } from "~/operations/domain";

/** A response's body as text, or how many bytes of something else. */
function bodyText(body: Uint8Array): string {
    try {
        return new TextDecoder("utf-8", { fatal: true }).decode(body);
    } catch {
        return `${body.length} bytes that are not text`;
    }
}

/** The status and duration of a function's call, beside the run's other details. */
export function FunctionInvokeSummary({ response }: Props) {
    const failed = response.outcome !== "ok" || response.status >= 400;

    return (
        <div>
            <span className="text-muted-foreground">Response:</span>{" "}
            <Badge
                tone={failed ? "red" : "green"}
                className="font-mono"
            >
                {response.status}
            </Badge>{" "}
            <span className="text-muted-foreground">
                in {response.durationMs.toFixed(1)} ms
                {response.outcome !== "ok" ? ` (${response.outcome})` : ""}
            </span>
        </div>
    );
}

/** The body of a function's response, as the run kept it. */
export function FunctionInvokeBody({ response }: Props) {
    if (response.body.length === 0) {
        return null;
    }

    return (
        <div className="flex flex-col gap-1.5 pt-1">
            <span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                Response Body
            </span>
            <pre className="max-h-80 overflow-auto rounded-md border bg-muted p-3 font-mono text-xs whitespace-pre-wrap break-all select-all">
                {bodyText(response.body) + (response.bodyTruncated ? "\n… cut at 64 KB" : "")}
            </pre>
        </div>
    );
}

interface Props {
    response: SystemTaskFunctionInvoke;
}
