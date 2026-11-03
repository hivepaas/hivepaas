import { useEffect, useMemo, useState } from "react";

import { AlertTriangle, Copy, Info } from "lucide-react";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import type { EnvLinkGroup } from "~/projects/api/services";
import { ProjectAppEnvVarsQueries } from "~/projects/data/queries";

import { AppLink, AppLoader } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

import {
    Button,
    Checkbox,
    Dialog,
    DialogActionFooter,
    DialogBody,
    DialogFixedContent,
    DialogHeader,
    DialogTitle,
    Input,
    Separator,
} from "@/components/ui";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

import { type LinkRow, rowIssues, rowsFromGroups } from "../link-app";

const ISSUE_TEXT = {
    invalid: "Not a variable name",
    duplicate: "Another selected row has this name",
    exists: "The form already has this variable",
} as const;

const SELF_GROUP_ID = "self";

/**
 * The variables an engine's official image reads to set itself up - its user,
 * its password, its first database - each a reference to the credentials this
 * app's App Kind publishes, so that they are kept in one place.
 */
export function SuggestEnvDialog({ open, onOpenChange, existingKeys, onAdd }: Props) {
    const { id: projectId, env, appId } = useParams<{ id: string; env: string; appId: string }>();
    invariant(projectId && env && appId, "the app's route params must be defined");

    // Empty asks for App Kind's engine.
    const [engine, setEngine] = useState("");
    const [rows, setRows] = useState<LinkRow[]>([]);

    useEffect(() => {
        if (!open) {
            setEngine("");
            setRows([]);
        }
    }, [open]);

    const query = ProjectAppEnvVarsQueries.useFindSelfSuggestions(
        { projectID: projectId, env, appID: appId, engine },
        { enabled: open },
    );
    const data = query.data?.data;
    const existing = existingKeys();

    useEffect(() => {
        const group: EnvLinkGroup = {
            id: SELF_GROUP_ID,
            title: "",
            description: "",
            recommended: true,
            warnings: [],
            vars: data?.vars ?? [],
        };
        // A variable the form has already starts unselected: replacing it is a choice.
        const keys = existingKeys();
        setRows(rowsFromGroups([group]).map(row => ({ ...row, selected: !keys.has(row.key) })));
        // existingKeys reads the form when the suggestions arrive, not on each render.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [data]);

    const issues = rowIssues(rows, "", existing);
    const selected = rows.filter(row => row.selected);
    const blocked = selected.some(row => (issues[row.id] ?? "") !== "");

    function updateRow(id: string, change: Partial<LinkRow>) {
        setRows(current => current.map(row => (row.id === id ? { ...row, ...change } : row)));
    }

    const engines = useMemo(() => data?.engines ?? [], [data]);
    const chosen = data?.engine ?? null;

    function copyCommand(command: string) {
        void navigator.clipboard
            .writeText(command)
            .then(() => {
                toast.success("Command copied to clipboard");
            })
            .catch(() => {
                toast.error("Failed to copy the command");
            });
    }

    return (
        <Dialog
            open={open}
            onOpenChange={onOpenChange}
        >
            <DialogFixedContent className="flex h-[80vh] w-[900px] max-w-[calc(100vw-1rem)] flex-col">
                <DialogHeader>
                    <DialogTitle>Suggest env vars for an engine</DialogTitle>
                </DialogHeader>
                <div className="px-4">
                    <Separator className="opacity-50" />
                </div>

                <DialogBody className="flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto">
                    <p className="text-sm text-muted-foreground">
                        For an app that runs a database or cache image: the variables its official image reads to set
                        itself up. Each refers to what this app&apos;s{" "}
                        <AppLink.Modules
                            to={ROUTE.projects.single.apps.single.configuration.kindSettings.$route(
                                projectId,
                                env,
                                appId,
                            )}
                            className="font-semibold text-link hover:underline"
                        >
                            App Kind
                        </AppLink.Modules>{" "}
                        publishes, such as <code>{"${HIVEPAAS_PASSWORD}"}</code>, so the credentials stay there.
                    </p>

                    <Select
                        value={chosen?.id ?? ""}
                        onValueChange={setEngine}
                    >
                        <SelectTrigger className="w-full sm:w-[320px]">
                            <SelectValue placeholder="Choose an engine" />
                        </SelectTrigger>
                        <SelectContent>
                            {engines.map(item => (
                                <SelectItem
                                    key={item.id}
                                    value={item.id}
                                >
                                    <span className="flex items-center gap-2">
                                        <span>{item.title}</span>
                                        <span className="text-xs text-muted-foreground">{item.image}</span>
                                        {item.id === data?.appEngine && (
                                            <span className="text-xs text-muted-foreground">· App Kind&apos;s</span>
                                        )}
                                    </span>
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>

                    {query.isFetching && (
                        <div className="flex flex-1 items-center justify-center">
                            <AppLoader />
                        </div>
                    )}

                    {!query.isFetching && !chosen && (
                        <p className="text-sm text-muted-foreground">
                            App Kind names no engine this knows: choose the one the app&apos;s image runs.
                        </p>
                    )}

                    {!query.isFetching && chosen && (
                        <div className="flex flex-col gap-3">
                            {data?.warnings.map(warning => (
                                <p
                                    key={warning}
                                    className="flex items-start gap-1.5 text-xs text-amber-600"
                                >
                                    <AlertTriangle className="mt-0.5 size-3.5 shrink-0" />
                                    {warning}
                                </p>
                            ))}

                            {chosen.initOnly && (
                                <p className="flex items-start gap-1.5 text-xs text-muted-foreground">
                                    <Info className="mt-0.5 size-3.5 shrink-0" />
                                    {chosen.title} reads these only when it first creates its data. On a database that
                                    has data already, they change nothing: its users and passwords are changed inside
                                    it.
                                </p>
                            )}

                            {data?.command && (
                                <div className="flex flex-col gap-2 rounded-md border p-3">
                                    <p className="text-sm">
                                        {chosen.title} reads no variable for its password. Run the image with this
                                        command, in{" "}
                                        <AppLink.Modules
                                            to={ROUTE.projects.single.apps.single.configuration.deploymentSettings.$route(
                                                projectId,
                                                env,
                                                appId,
                                            )}
                                            className="font-semibold text-link hover:underline"
                                        >
                                            Deployment Settings
                                        </AppLink.Modules>
                                        : it starts the server with App Kind&apos;s password, memory, eviction and
                                        persistence.
                                    </p>
                                    <div className="flex items-start gap-2">
                                        <code className="flex-1 break-all rounded bg-muted/60 p-2 text-xs">
                                            {data.command}
                                        </code>
                                        <Button
                                            type="button"
                                            variant="outline"
                                            size="sm"
                                            onClick={() => {
                                                copyCommand(data.command);
                                            }}
                                        >
                                            <Copy className="size-3.5" />
                                            Copy
                                        </Button>
                                    </div>
                                </div>
                            )}

                            {rows.map(row => {
                                const issue = issues[row.id] ?? "";
                                return (
                                    <div
                                        key={row.id}
                                        className="grid grid-cols-[auto_260px_1fr] items-start gap-2"
                                    >
                                        <Checkbox
                                            className="mt-2.5"
                                            checked={row.selected}
                                            onCheckedChange={checked => {
                                                updateRow(row.id, { selected: checked === true });
                                            }}
                                        />
                                        <div className="flex flex-col gap-1">
                                            <Input
                                                aria-label={`${row.key} name`}
                                                value={row.key}
                                                aria-invalid={issue !== ""}
                                                onChange={event => {
                                                    updateRow(row.id, { key: event.target.value });
                                                }}
                                            />
                                            {issue !== "" && (
                                                <span className="text-xs text-destructive">{ISSUE_TEXT[issue]}</span>
                                            )}
                                            {(issue === "exists" || row.replace) && row.selected && (
                                                <div className="flex items-center gap-1.5 text-xs">
                                                    <Checkbox
                                                        id={`self-replace-${row.id}`}
                                                        checked={row.replace}
                                                        onCheckedChange={checked => {
                                                            updateRow(row.id, { replace: checked === true });
                                                        }}
                                                    />
                                                    <label htmlFor={`self-replace-${row.id}`}>Replace it</label>
                                                </div>
                                            )}
                                            {issue === "" && !row.selected && existing.has(row.key) && (
                                                <span className="text-xs text-muted-foreground">
                                                    The form has it already
                                                </span>
                                            )}
                                        </div>
                                        <div className="flex flex-col gap-1 pt-2">
                                            <code className="break-all text-xs">{row.value}</code>
                                            {row.description && (
                                                <span className="text-xs text-muted-foreground">{row.description}</span>
                                            )}
                                        </div>
                                    </div>
                                );
                            })}
                        </div>
                    )}
                </DialogBody>

                <DialogActionFooter>
                    <Button
                        type="button"
                        variant="outline"
                        onClick={() => {
                            onOpenChange(false);
                        }}
                    >
                        Cancel
                    </Button>
                    <Button
                        type="button"
                        disabled={selected.length === 0 || blocked}
                        onClick={() => {
                            onAdd(selected.map(row => ({ key: row.key.trim(), value: row.value })));
                            onOpenChange(false);
                        }}
                    >
                        Add {selected.length} variable{selected.length === 1 ? "" : "s"}
                    </Button>
                </DialogActionFooter>
            </DialogFixedContent>
        </Dialog>
    );
}

interface Props {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    /** The keys the form holds in the runtime section. */
    existingKeys: () => Set<string>;
    /** Adds the rows to the runtime section; a key it has is replaced in place. */
    onAdd: (vars: { key: string; value: string }[]) => void;
}
