import { useEffect, useMemo, useState } from "react";

import { Badge } from "@components/ui/badge";
import { AlertTriangle } from "lucide-react";
import { useParams } from "react-router";
import invariant from "tiny-invariant";
import type { EnvLinkTarget } from "~/projects/api/services";
import { ProjectAppEnvVarsQueries } from "~/projects/data/queries";

import { AppLoader, Combobox } from "@application/shared/components";

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
    Tabs,
    TabsList,
    TabsTrigger,
} from "@/components/ui";

import { type LinkRow, type LinkSection, finalKey, rowIssues, rowsFromGroups } from "./link-app.utils";

const ISSUE_TEXT = {
    invalid: "Not a variable name",
    duplicate: "Another selected row has this name",
    exists: "The form already has this variable",
} as const;

export function LinkAppDialog({ open, initialSection, onOpenChange, existingKeys, onAdd }: Props) {
    const { id: projectId, env, appId } = useParams<{ id: string; env: string; appId: string }>();
    invariant(projectId && env && appId, "the app's route params must be defined");

    const [targetID, setTargetID] = useState<string | null>(null);
    const [prefix, setPrefix] = useState("");
    const [section, setSection] = useState<LinkSection>(initialSection);
    const [rows, setRows] = useState<LinkRow[]>([]);

    useEffect(() => {
        if (open) {
            setSection(initialSection);
        } else {
            setTargetID(null);
            setPrefix("");
            setRows([]);
        }
    }, [open, initialSection]);

    const targetsQuery = ProjectAppEnvVarsQueries.useFindLinkTargets(
        { projectID: projectId, env, appID: appId },
        { enabled: open },
    );
    const suggestionsQuery = ProjectAppEnvVarsQueries.useFindLinkSuggestions(
        { projectID: projectId, env, appID: appId, targetAppID: targetID ?? "" },
        { enabled: open && Boolean(targetID) },
    );
    const groups = useMemo(() => suggestionsQuery.data?.data.groups ?? [], [suggestionsQuery.data]);

    useEffect(() => {
        setRows(rowsFromGroups(groups));
    }, [groups]);

    const existing = existingKeys(section);
    const issues = rowIssues(rows, prefix, existing);
    const selected = rows.filter(row => row.selected);
    const blocked = selected.some(row => (issues[row.id] ?? "") !== "");

    function updateRow(id: string, change: Partial<LinkRow>) {
        setRows(current => current.map(row => (row.id === id ? { ...row, ...change } : row)));
    }

    const targets = targetsQuery.data?.data ?? [];
    const targetOptions = targets.map(target => ({ value: { id: target.id, name: target.name }, label: target.name }));

    return (
        <Dialog
            open={open}
            onOpenChange={onOpenChange}
        >
            <DialogFixedContent className="flex h-[80vh] w-[1000px] max-w-[calc(100vw-1rem)] flex-col">
                <DialogHeader>
                    <DialogTitle>Link to another app</DialogTitle>
                </DialogHeader>
                <div className="px-4">
                    <Separator className="opacity-50" />
                </div>

                <DialogBody className="flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto">
                    <div className="grid grid-cols-1 gap-3 sm:grid-cols-[1fr_200px_auto]">
                        <Combobox
                            options={targetOptions}
                            value={targetID}
                            onChange={(_, option) => {
                                setTargetID(option?.id ?? null);
                            }}
                            placeholder="Choose an app of this env"
                            emptyText="No other app in this env"
                            valueKey="id"
                            loading={targetsQuery.isFetching}
                            renderOption={option => (
                                <TargetOption target={targets.find(target => target.id === option.value.id)} />
                            )}
                        />
                        <Input
                            aria-label="Prefix"
                            placeholder="Prefix, e.g. ANALYTICS_"
                            value={prefix}
                            onChange={event => {
                                setPrefix(event.target.value);
                            }}
                        />
                        <Tabs
                            value={section}
                            onValueChange={value => {
                                setSection(value as LinkSection);
                            }}
                        >
                            <TabsList>
                                <TabsTrigger value="runtime">Runtime</TabsTrigger>
                                <TabsTrigger value="buildtime">Build-time</TabsTrigger>
                            </TabsList>
                        </Tabs>
                    </div>

                    {!targetID && (
                        <p className="text-sm text-muted-foreground">
                            Choose an app to see the variables that connect to it. Each is a reference such as{" "}
                            <code>{"${db.HIVEPAAS_HOST}"}</code>, resolved when this app is built.
                        </p>
                    )}

                    {targetID && suggestionsQuery.isFetching && (
                        <div className="flex flex-1 items-center justify-center">
                            <AppLoader />
                        </div>
                    )}

                    {targetID &&
                        !suggestionsQuery.isFetching &&
                        groups.map(group => {
                            const groupRows = rows.filter(row => row.groupId === group.id);
                            const allSelected = groupRows.length > 0 && groupRows.every(row => row.selected);
                            return (
                                <div
                                    key={group.id}
                                    className="flex flex-col gap-2 rounded-md border p-3"
                                >
                                    <div className="flex items-center gap-2">
                                        <Checkbox
                                            checked={allSelected}
                                            onCheckedChange={checked => {
                                                setRows(current =>
                                                    current.map(row =>
                                                        row.groupId === group.id
                                                            ? { ...row, selected: checked === true }
                                                            : row,
                                                    ),
                                                );
                                            }}
                                        />
                                        <span className="font-medium">{group.title}</span>
                                        {group.recommended && <Badge variant="outline">Recommended</Badge>}
                                    </div>
                                    {group.description && (
                                        <p className="text-sm text-muted-foreground">{group.description}</p>
                                    )}
                                    {group.warnings.map(warning => (
                                        <p
                                            key={warning}
                                            className="flex items-center gap-1.5 text-xs text-amber-600"
                                        >
                                            <AlertTriangle className="size-3.5 shrink-0" />
                                            {warning}
                                        </p>
                                    ))}
                                    {groupRows.map(row => {
                                        const issue = issues[row.id] ?? "";
                                        return (
                                            <div
                                                key={row.id}
                                                className="grid grid-cols-[auto_240px_1fr] items-start gap-2"
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
                                                    {prefix.trim() && (
                                                        <span className="text-xs text-muted-foreground">
                                                            {finalKey(prefix, row.key)}
                                                        </span>
                                                    )}
                                                    {issue !== "" && (
                                                        <span className="text-xs text-destructive">
                                                            {ISSUE_TEXT[issue]}
                                                        </span>
                                                    )}
                                                    {(issue === "exists" || row.replace) && row.selected && (
                                                        <div className="flex items-center gap-1.5 text-xs">
                                                            <Checkbox
                                                                id={`link-replace-${row.id}`}
                                                                checked={row.replace}
                                                                onCheckedChange={checked => {
                                                                    updateRow(row.id, { replace: checked === true });
                                                                }}
                                                            />
                                                            <label htmlFor={`link-replace-${row.id}`}>Replace it</label>
                                                        </div>
                                                    )}
                                                </div>
                                                <div className="flex flex-col gap-1 pt-2">
                                                    <code className="break-all text-xs">{row.value}</code>
                                                    {row.description && (
                                                        <span className="text-xs text-muted-foreground">
                                                            {row.description}
                                                        </span>
                                                    )}
                                                </div>
                                            </div>
                                        );
                                    })}
                                </div>
                            );
                        })}
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
                            onAdd(
                                section,
                                selected.map(row => ({ key: finalKey(prefix, row.key), value: row.value })),
                            );
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

/** A target in the list: its name, its key, and its kind when it has one. */
function TargetOption({ target }: { target?: EnvLinkTarget }) {
    if (!target) {
        return null;
    }
    return (
        <span className="flex items-center gap-2">
            <span>{target.name}</span>
            <span className="text-xs text-muted-foreground">{target.key}</span>
            {target.category && (
                <Badge variant="outline">
                    {target.engine ? `${target.category} · ${target.engine}` : target.category}
                </Badge>
            )}
        </span>
    );
}

interface Props {
    open: boolean;
    initialSection: LinkSection;
    onOpenChange: (open: boolean) => void;
    /** The keys the form holds in a section. */
    existingKeys: (section: LinkSection) => Set<string>;
    /** Adds the rows; a key the section has is replaced in place. */
    onAdd: (section: LinkSection, vars: { key: string; value: string }[]) => void;
}
