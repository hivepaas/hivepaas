import { useMemo } from "react";

import { Button } from "@components/ui";
import { AlertTriangleIcon, CheckCircle2Icon, XCircleIcon } from "lucide-react";
import type { SpecImportResult } from "~/operations/domain";

import { SpecImportPlanTree } from "./spec-import-plan-tree.com";
import { buildImportTree, leafPaths } from "./spec-import.tree";

/**
 * What the import did. The database part is all or nothing and has happened by
 * now; what follows it - bringing running apps to their new configuration - can
 * fail for one app and not the others, and importing again retries it.
 */
export function SpecImportResultPanel({ result, onImportAgain, onChooseAnother }: Props) {
    const roots = useMemo(() => buildImportTree(result.plan.nodes), [result]);
    const selectedPaths = useMemo(
        () => new Set(result.plan.nodes.filter(node => node.selected).map(node => node.path)),
        [result],
    );
    // The tree draws checkboxes from leaves; here they only mirror what was taken.
    const checked = useMemo(
        () => new Set(leafPaths(roots).filter(path => selectedPaths.has(path))),
        [roots, selectedPaths],
    );

    const counts = { applied: 0, failed: 0, skipped: 0, unchanged: 0 };
    for (const node of result.plan.nodes) {
        if (node.outcome) {
            counts[node.outcome]++;
        }
    }
    const failed = counts.failed > 0;

    return (
        <div className="rounded-lg border bg-background p-4">
            <div className="flex flex-col gap-4">
                <div className="flex items-start gap-2">
                    {failed ? (
                        <XCircleIcon className="mt-0.5 size-5 shrink-0 text-destructive" />
                    ) : (
                        <CheckCircle2Icon className="mt-0.5 size-5 shrink-0 text-emerald-600" />
                    )}
                    <div>
                        <p className="text-sm font-medium text-foreground">
                            {failed
                                ? `Imported, but ${counts.failed} ${counts.failed === 1 ? "app was" : "apps were"} not brought up to date`
                                : "Imported"}
                        </p>
                        <p className="text-xs text-muted-foreground">
                            {counts.applied} applied · {counts.unchanged} unchanged · {counts.skipped} left out
                            {result.deployments.length > 0 &&
                                ` · ${result.deployments.length} ${result.deployments.length === 1 ? "deployment" : "deployments"} queued`}
                        </p>
                        {failed && (
                            <p className="mt-1 max-w-[720px] text-xs text-muted-foreground">
                                Their configuration is saved; their running service was not updated. Importing the same
                                bundle again retries them.
                            </p>
                        )}
                    </div>
                </div>

                {result.warning && (
                    <p className="flex items-start gap-2 text-sm text-amber-700 dark:text-amber-400">
                        <AlertTriangleIcon className="mt-0.5 size-4 shrink-0" />
                        <span className="whitespace-pre-line">{result.warning}</span>
                    </p>
                )}

                <SpecImportPlanTree
                    roots={roots}
                    checked={checked}
                    onChange={() => undefined}
                    readOnly
                />

                <div className="flex gap-2">
                    <Button
                        type="button"
                        variant="outline"
                        onClick={onImportAgain}
                    >
                        Import again
                    </Button>
                    <Button
                        type="button"
                        variant="ghost"
                        onClick={onChooseAnother}
                    >
                        Choose another file
                    </Button>
                </div>
            </div>
        </div>
    );
}

interface Props {
    result: SpecImportResult;
    onImportAgain: () => void;
    onChooseAnother: () => void;
}
