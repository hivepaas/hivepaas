import { useState } from "react";

import { Play, Plus } from "lucide-react";
import { AppFunctionCommands } from "~/projects/data";
import { type FunctionFile, type FunctionTestRunResult } from "~/projects/domain";
import { buildTestRequest, describeBody } from "~/projects/module-shared/utils";

import {
    Button,
    Input,
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
    Tabs,
    TabsContent,
    TabsList,
    TabsTrigger,
} from "@/components/ui";
import { Badge, type BadgeTone } from "@/components/ui/badge";
import { Textarea } from "@/components/ui/textarea";

const METHODS = ["GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"];

const OUTCOMES: Record<string, { label: string; tone: BadgeTone }> = {
    "ok": { label: "OK", tone: "green" },
    "error": { label: "Error", tone: "red" },
    "timeout": { label: "Timed out", tone: "orange" },
    "libraries-failed": { label: "Libraries not installed", tone: "red" },
    "not-loaded": { label: "Not loaded", tone: "red" },
    "bad-request": { label: "Request not read", tone: "red" },
    "killed": { label: "Killed", tone: "red" },
    "no-result": { label: "No result", tone: "red" },
};

function Pre({ children }: { children: string }) {
    return (
        <pre className="max-h-80 overflow-auto rounded-md bg-muted p-3 font-mono text-xs whitespace-pre-wrap break-all">
            {children}
        </pre>
    );
}

function TestResult({ result, onAddLockFiles }: { result: FunctionTestRunResult; onAddLockFiles?: () => void }) {
    const outcome = OUTCOMES[result.outcome] ?? { label: result.outcome, tone: "neutral" as const };
    const body = describeBody(result.body, result.headers);
    const hasResponse = result.status > 0;
    const headerLines = Object.entries(result.headers)
        .flatMap(([name, values]) => values.map(value => `${name}: ${value}`))
        .join("\n");

    return (
        <div className="flex flex-col gap-3">
            <div className="flex flex-wrap items-center gap-2 text-sm">
                <Badge
                    tone={outcome.tone}
                    appearance="solid"
                >
                    {outcome.label}
                </Badge>
                {hasResponse ? <span className="font-mono">{result.status}</span> : null}
                {hasResponse ? <span className="text-muted-foreground">{result.durationMs.toFixed(1)} ms</span> : null}
                {result.librariesBuilt ? <span className="text-muted-foreground">libraries installed</span> : null}
                {!hasResponse && result.exitCode !== 0 ? (
                    <span className="text-muted-foreground">exit code {result.exitCode}</span>
                ) : null}
            </div>

            {result.lockFiles.length > 0 && onAddLockFiles ? (
                <div className="flex flex-wrap items-center justify-between gap-2 rounded-md border px-3 py-2 text-sm">
                    <span>
                        The run made {result.lockFiles.map(file => file.path).join(", ")}: add it to the code, and the
                        next build installs the same versions.
                    </span>
                    <Button
                        type="button"
                        size="sm"
                        variant="outline"
                        onClick={onAddLockFiles}
                    >
                        <Plus /> Add to the code
                    </Button>
                </div>
            ) : null}

            <Tabs defaultValue={hasResponse ? "body" : result.librariesLog ? "libraries" : "error"}>
                <TabsList>
                    {hasResponse ? <TabsTrigger value="body">Body</TabsTrigger> : null}
                    {hasResponse ? <TabsTrigger value="headers">Headers</TabsTrigger> : null}
                    <TabsTrigger value="logs">Logs</TabsTrigger>
                    {result.error ? <TabsTrigger value="error">Errors</TabsTrigger> : null}
                    {result.librariesLog ? <TabsTrigger value="libraries">Install</TabsTrigger> : null}
                </TabsList>
                {hasResponse ? (
                    <TabsContent value="body">
                        <Pre>{body.text + (result.bodyTruncated ? "\n… cut at 1 MB" : "")}</Pre>
                    </TabsContent>
                ) : null}
                {hasResponse ? (
                    <TabsContent value="headers">
                        <Pre>{headerLines}</Pre>
                    </TabsContent>
                ) : null}
                <TabsContent value="logs">
                    <Pre>{(result.logsTruncated ? "… cut at 1 MB\n" : "") + (result.logs || "No logs")}</Pre>
                </TabsContent>
                {result.error ? (
                    <TabsContent value="error">
                        <Pre>{result.error}</Pre>
                    </TabsContent>
                ) : null}
                {result.librariesLog ? (
                    <TabsContent value="libraries">
                        <Pre>{result.librariesLog}</Pre>
                    </TabsContent>
                ) : null}
            </Tabs>
        </div>
    );
}

/**
 * A test run of the code as it is in the editor, not yet saved: a request, the
 * answer, the logs, and the lock file the run made.
 */
export function FunctionTestPanel({ projectId, env, appId, files, readOnly, onAddFiles }: Props) {
    const [method, setMethod] = useState("GET");
    const [path, setPath] = useState("/");
    const [headers, setHeaders] = useState("");
    const [body, setBody] = useState("");
    const { mutate: run, data, error, isPending, reset } = AppFunctionCommands.useTestRun();
    const result = data?.data;
    const hasBody = method !== "GET" && method !== "HEAD";

    return (
        <div className="flex w-full flex-col gap-3 xl:w-[440px] xl:shrink-0">
            <div className="flex gap-2">
                <Select
                    value={method}
                    onValueChange={setMethod}
                >
                    <SelectTrigger className="w-28">
                        <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                        {METHODS.map(item => (
                            <SelectItem
                                key={item}
                                value={item}
                            >
                                {item}
                            </SelectItem>
                        ))}
                    </SelectContent>
                </Select>
                <Input
                    value={path}
                    onChange={event => {
                        setPath(event.target.value);
                    }}
                    placeholder="/hello?name=Ada"
                    className="font-mono"
                    aria-label="Path and query"
                />
                <Button
                    type="button"
                    isLoading={isPending}
                    disabled={readOnly || files.length === 0}
                    onClick={() => {
                        reset();
                        run({
                            projectID: projectId,
                            env,
                            appID: appId,
                            files,
                            request: buildTestRequest(method, path, headers, body),
                        });
                    }}
                >
                    <Play /> Run
                </Button>
            </div>
            <Textarea
                value={headers}
                onChange={event => {
                    setHeaders(event.target.value);
                }}
                placeholder={"Headers, one per line\ncontent-type: application/json"}
                rows={2}
                className="font-mono text-xs"
                aria-label="Headers"
            />
            {hasBody ? (
                <Textarea
                    value={body}
                    onChange={event => {
                        setBody(event.target.value);
                    }}
                    placeholder="Body"
                    rows={5}
                    className="font-mono text-xs"
                    aria-label="Body"
                />
            ) : null}

            {isPending ? (
                <p className="text-sm text-muted-foreground">
                    Running on a build node. A first run after the libraries change installs them, which can take a few
                    minutes.
                </p>
            ) : null}
            {error ? <p className="text-sm text-destructive">{error.message}</p> : null}
            {result ? (
                <TestResult
                    result={result}
                    onAddLockFiles={
                        readOnly
                            ? undefined
                            : () => {
                                  onAddFiles(result.lockFiles);
                              }
                    }
                />
            ) : null}
        </div>
    );
}

interface Props {
    projectId: string;
    env: string;
    appId: string;
    /** The code as the editor has it. */
    files: FunctionFile[];
    readOnly: boolean;
    onAddFiles: (files: FunctionFile[]) => void;
}
