import { useMemo } from "react";

import { Button } from "@components/ui";
import { AlertTriangleIcon, CheckCircle2Icon, XCircleIcon } from "lucide-react";
import { Link } from "react-router";
import type { ComposeImportResult } from "~/operations/domain";
import { SpecImportPlanTree, buildImportTree, leafPaths } from "~/operations/routes/export";

import { ROUTE } from "@application/shared/constants";

/**
 * What creating the project did: the project, its apps, the deployments queued
 * and what failed after it was saved.
 */
export function ComposeResult({ result, onStartOver }: Props) {
    const roots = useMemo(() => buildImportTree(result.plan.nodes), [result]);
    const checked = useMemo(() => {
        const selected = new Set(result.plan.nodes.filter(node => node.selected).map(node => node.path));

        return new Set(leafPaths(roots).filter(path => selected.has(path)));
    }, [roots, result]);
    const failed = result.plan.nodes.filter(node => node.outcome === "failed").length;

    return (
        <div className="rounded-lg border bg-background p-4">
            <div className="flex flex-col gap-4">
                <div className="flex items-start gap-2">
                    {failed > 0 ? (
                        <XCircleIcon className="mt-0.5 size-5 shrink-0 text-destructive" />
                    ) : (
                        <CheckCircle2Icon className="mt-0.5 size-5 shrink-0 text-emerald-600" />
                    )}
                    <div>
                        <p className="text-sm font-medium text-foreground">
                            {failed > 0
                                ? `The project is created, but ${failed} ${failed === 1 ? "app was" : "apps were"} not brought up`
                                : "The project is created"}
                        </p>
                        <p className="text-xs text-muted-foreground">
                            {result.deployments.length > 0
                                ? `${result.deployments.length} ${result.deployments.length === 1 ? "deployment" : "deployments"} queued: the apps start as their images are pulled.`
                                : "No deployment was queued: the apps run a placeholder until they are deployed."}
                        </p>
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
                    {result.projectId && (
                        <Button asChild>
                            <Link to={ROUTE.projects.single.apps.$route(result.projectId)}>Open the project</Link>
                        </Button>
                    )}
                    <Button
                        type="button"
                        variant="ghost"
                        onClick={onStartOver}
                    >
                        Create another
                    </Button>
                </div>
            </div>
        </div>
    );
}

interface Props {
    result: ComposeImportResult;
    onStartOver: () => void;
}
