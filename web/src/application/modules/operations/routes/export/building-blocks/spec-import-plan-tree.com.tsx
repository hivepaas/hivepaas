import { cn } from "@/lib/utils";
import { Checkbox } from "@components/ui";
import { Badge } from "@components/ui/badge";
import { AlertTriangleIcon, InfoIcon, OctagonXIcon, SkipForwardIcon, WrenchIcon } from "lucide-react";
import type {
    SpecImportIssue,
    SpecImportNode,
    SpecImportNodeAction,
    SpecImportNodeKind,
    SpecImportNodeOutcome,
    SpecImportSeverity,
} from "~/operations/domain";

import { type SpecImportTreeNode, checkStateOf, flattenImportTree, toggleNode } from "./spec-import.tree";

const KIND_LABELS: Record<SpecImportNodeKind, string> = {
    global: "Global settings",
    project: "Project",
    env: "Environment",
    settings: "Settings",
    app: "App",
};

const ACTION_STYLES: Record<SpecImportNodeAction, { label: string; className: string }> = {
    create: {
        label: "create",
        className: "border-emerald-500/40 bg-emerald-500/10 text-emerald-700 dark:text-emerald-400",
    },
    update: { label: "update", className: "border-sky-500/40 bg-sky-500/10 text-sky-700 dark:text-sky-400" },
    unchanged: { label: "unchanged", className: "text-muted-foreground" },
    keep: { label: "kept as is", className: "text-muted-foreground" },
    skip: { label: "left out", className: "border-amber-500/40 bg-amber-500/10 text-amber-700 dark:text-amber-400" },
};

const OUTCOME_STYLES: Record<SpecImportNodeOutcome, { label: string; className: string }> = {
    applied: {
        label: "applied",
        className: "border-emerald-500/40 bg-emerald-500/10 text-emerald-700 dark:text-emerald-400",
    },
    unchanged: { label: "unchanged", className: "text-muted-foreground" },
    skipped: { label: "left out", className: "border-amber-500/40 bg-amber-500/10 text-amber-700 dark:text-amber-400" },
    failed: { label: "failed", className: "border-destructive/40 bg-destructive/10 text-destructive" },
};

const SEVERITY_STYLES: Record<SpecImportSeverity, { icon: typeof InfoIcon; className: string }> = {
    blocked: { icon: OctagonXIcon, className: "text-destructive" },
    skipped: { icon: SkipForwardIcon, className: "text-amber-600 dark:text-amber-400" },
    fixable: { icon: WrenchIcon, className: "text-amber-600 dark:text-amber-400" },
    warning: { icon: AlertTriangleIcon, className: "text-amber-600 dark:text-amber-400" },
};

/** The detail an issue carries, as `key: value` pairs - never a secret's value, which the server never sends. */
function describeDetail(detail: Record<string, unknown> | undefined): string {
    if (!detail) {
        return "";
    }

    return Object.entries(detail)
        .map(([key, value]) => {
            const text = typeof value === "object" && value !== null ? JSON.stringify(value) : String(value);

            return `${key}: ${text}`;
        })
        .join(" · ");
}

function IssueLine({ issue }: { issue: SpecImportIssue }) {
    const style = issue.severity ? SEVERITY_STYLES[issue.severity] : undefined;
    const Icon = style?.icon ?? InfoIcon;
    const detail = describeDetail(issue.detail);

    return (
        <div className="flex items-start gap-2 text-xs">
            <Icon className={cn("mt-0.5 size-3.5 shrink-0", style?.className ?? "text-muted-foreground")} />
            <div className="min-w-0">
                <span className="text-foreground">{issue.action ?? issue.code}</span>{" "}
                <code className="font-mono text-[11px] text-muted-foreground">{issue.code}</code>
                {detail && <div className="break-all text-muted-foreground">{detail}</div>}
                {issue.availableIn && (
                    <div className="text-muted-foreground">
                        In the bundle at <code className="font-mono">{issue.availableIn}</code>
                    </div>
                )}
            </div>
        </div>
    );
}

function nodeTitle(node: SpecImportNode): string {
    if (node.kind === "global") {
        return KIND_LABELS.global;
    }
    if (node.kind === "settings") {
        return "Own settings";
    }

    return node.name && node.name !== node.key ? `${node.name} (${node.key})` : (node.key ?? node.path);
}

function NodeRow({ treeNode, checked, onChange, readOnly }: RowProps) {
    const { node } = treeNode;
    const state = checkStateOf(treeNode, checked);
    const action = ACTION_STYLES[node.action];
    const outcome = node.outcome ? OUTCOME_STYLES[node.outcome] : undefined;
    const pulledIn = node.selectedBy === "dependency";

    return (
        <div
            className={cn("flex flex-col gap-1 border-b py-2 last:border-b-0", !node.selected && "opacity-60")}
            style={{ paddingLeft: treeNode.depth * 20 }}
        >
            <div className="flex flex-wrap items-center gap-2">
                <Checkbox
                    checked={state === "indeterminate" ? "indeterminate" : state === "checked"}
                    disabled={readOnly}
                    className="data-[state=indeterminate]:border-primary data-[state=indeterminate]:bg-primary/40"
                    onCheckedChange={value => {
                        onChange(toggleNode(treeNode, checked, value === true));
                    }}
                    aria-label={`Import ${nodeTitle(node)}`}
                />
                <span className="text-sm font-medium text-foreground">{nodeTitle(node)}</span>
                {node.kind !== "settings" && node.kind !== "global" && (
                    <span className="text-xs text-muted-foreground">{KIND_LABELS[node.kind]}</span>
                )}
                {outcome ? (
                    <Badge
                        variant="outline"
                        className={outcome.className}
                    >
                        {outcome.label}
                    </Badge>
                ) : (
                    node.selected && (
                        <Badge
                            variant="outline"
                            className={action.className}
                        >
                            {action.label}
                        </Badge>
                    )
                )}
                {pulledIn && (
                    <Badge
                        variant="outline"
                        title="Not chosen, but something chosen needs it and this installation lacks it"
                    >
                        pulled in
                    </Badge>
                )}
                {node.selected && node.restart && (
                    <Badge
                        variant="outline"
                        className="border-amber-500/40 text-amber-700 dark:text-amber-400"
                    >
                        restarts
                    </Badge>
                )}
                {node.selected && node.deploy && <Badge variant="outline">deploys</Badge>}
            </div>

            {node.selected && node.changes.length > 0 && node.action !== "skip" && (
                <p
                    className="pl-6 text-xs text-muted-foreground"
                    title={node.changes.join("\n")}
                >
                    {node.action === "create" ? "Creates " : "Changes "}
                    {node.changes.length > 4
                        ? `${node.changes.slice(0, 4).join(", ")} and ${node.changes.length - 4} more`
                        : node.changes.join(", ")}
                </p>
            )}

            {node.error && <p className="pl-6 text-xs text-destructive">{node.error}</p>}

            {node.selected && (node.issues.length > 0 || node.notes.length > 0) && (
                <div className="flex flex-col gap-1 pl-6">
                    {node.issues.map((issue, index) => (
                        <IssueLine
                            key={`${issue.code}-${index}`}
                            issue={issue}
                        />
                    ))}
                    {node.notes.map((note, index) => (
                        <IssueLine
                            key={`note-${note.code}-${index}`}
                            issue={note}
                        />
                    ))}
                </div>
            )}
        </div>
    );
}

interface RowProps {
    treeNode: SpecImportTreeNode;
    checked: ReadonlySet<string>;
    onChange: (checked: Set<string>) => void;
    readOnly?: boolean;
}

/**
 * The plan as a tree: each node with its checkbox, what import does to it, and
 * what it finds on the way. A node left unchecked is dimmed; one the import pulls
 * in anyway says so.
 */
export function SpecImportPlanTree({ roots, checked, onChange, readOnly }: Props) {
    return (
        <div className="flex flex-col rounded-md border px-3">
            {flattenImportTree(roots).map(treeNode => (
                <NodeRow
                    key={treeNode.node.path}
                    treeNode={treeNode}
                    checked={checked}
                    onChange={onChange}
                    readOnly={readOnly}
                />
            ))}
        </div>
    );
}

interface Props {
    roots: SpecImportTreeNode[];
    checked: ReadonlySet<string>;
    onChange: (checked: Set<string>) => void;
    readOnly?: boolean;
}
