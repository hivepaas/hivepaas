import { FileCode2, Loader2 } from "lucide-react";
import { toast } from "sonner";
import { AppManagedCommands, AppManagedQueries } from "~/projects/data";
import type { AppManagedScriptChange, ManagedScriptsState } from "~/projects/domain";

import { PopConfirm } from "@application/shared/components";

import { Badge, type StatusMeaning, statusClassName } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";

interface Props {
    projectId: string;
    env: string;
    appId: string;
    /** What writes now: this database, or the member promoted - restarted last. */
    writerName: string;
    canWrite: boolean;
}

const STATES: Record<ManagedScriptsState, { label: string; meaning: StatusMeaning }> = {
    "up-to-date": { label: "Up to date", meaning: "success" },
    "update-available": { label: "Update available", meaning: "attention" },
    "incompatible": { label: "Incompatible", meaning: "failure" },
    "unavailable": { label: "Unavailable", meaning: "inactive" },
};

/** A revision as a person reads it: its first twelve characters, or "unknown". */
function short(revision: string | undefined): string {
    return revision ? revision.slice(0, 12) : "unknown";
}

function changeText(change: AppManagedScriptChange): string {
    if (change.kind === "artifact") {
        switch (change.change) {
            case "added":
                return `${change.name} ${change.to} added`;
            case "removed":
                return `${change.name} ${change.from} removed`;
            default:
                return change.from === change.to
                    ? `${change.name} ${change.to} rebuilt`
                    : `${change.name} ${change.from} → ${change.to}`;
        }
    }
    return `${change.name} ${change.change}`;
}

/**
 * The scripts HivePaaS runs beside a managed database - its archiving, base backups, restores and replicas - as
 * its template ships them, against what the installation's templates now offer; and their update, which restarts
 * each member once, the one that writes last.
 */
export function ManagedScripts({ projectId, env, appId, writerName, canWrite }: Props) {
    const scope = { projectID: projectId, env, appID: appId };
    const { data, isLoading } = AppManagedQueries.useFindScripts(scope);
    const { mutate: upgrade, isPending } = AppManagedCommands.useUpgradeScripts({
        onSuccess: response => {
            const count = response.data.deployments.length;
            toast.success(`Scripts updated: ${count} ${count === 1 ? "app restarts" : "apps restart"} onto them`);
        },
    });
    const scripts = data?.data;
    if (isLoading || !scripts) {
        return null;
    }
    const state = STATES[scripts.state];
    const canUpgrade =
        scripts.available !== undefined &&
        (scripts.state === "update-available" || (scripts.state === "up-to-date" && scripts.outOfStep.length > 0));

    return (
        <section
            aria-label="Scripts"
            className="flex flex-col gap-3"
        >
            <div className="flex flex-col">
                <h2 className="text-sm font-semibold">Scripts</h2>
                <p className="text-xs text-muted-foreground">
                    What HivePaaS runs beside the database - archiving, base backups, restores, replicas - as its
                    template ships them. A newer revision of the templates may fix or add to them; it is applied only
                    when you update.
                </p>
            </div>
            <div className="flex flex-col divide-y rounded-md border">
                <div className="flex flex-wrap items-center justify-between gap-3 px-3 py-2.5 text-xs">
                    <div className="flex flex-col gap-0.5">
                        <div className="flex items-center gap-2">
                            <Badge className={statusClassName(state.meaning)}>{state.label}</Badge>
                            <span className="text-muted-foreground">
                                Runs revision {short(scripts.current.revision)}
                                {scripts.available && scripts.state !== "up-to-date"
                                    ? `; the templates offer ${short(scripts.available.revision)}`
                                    : ""}
                            </span>
                        </div>
                        {scripts.reason && <span className="text-muted-foreground">{scripts.reason}</span>}
                    </div>
                    <PopConfirm
                        title="Update the database's scripts?"
                        description={
                            "Each member restarts once onto them, " +
                            `${writerName} last: it does not serve while it restarts.`
                        }
                        confirmText="Update"
                        onConfirm={() => {
                            if (scripts.available) {
                                upgrade({ ...scope, payload: { fingerprint: scripts.available.fingerprint } });
                            }
                        }}
                    >
                        <Button
                            variant="outline"
                            size="sm"
                            disabled={!canWrite || !canUpgrade || isPending}
                        >
                            {isPending ? (
                                <Loader2 className="size-3.5 animate-spin" />
                            ) : (
                                <FileCode2 className="size-3.5" />
                            )}
                            Update scripts
                        </Button>
                    </PopConfirm>
                </div>
                {scripts.changes.length > 0 && (
                    <ul
                        aria-label="Changes"
                        className="flex flex-col gap-0.5 px-3 py-2.5 font-mono text-xs"
                    >
                        {scripts.changes.map(change => (
                            <li key={`${change.kind}:${change.name}`}>{changeText(change)}</li>
                        ))}
                    </ul>
                )}
                {scripts.outOfStep.length > 0 && (
                    <p className="px-3 py-2.5 text-xs text-muted-foreground">
                        {scripts.outOfStep.map(member => member.key).join(", ")}{" "}
                        {scripts.outOfStep.length === 1 ? "runs" : "run"} other scripts than the database: updating
                        brings {scripts.outOfStep.length === 1 ? "it" : "them"} back.
                    </p>
                )}
            </div>
        </section>
    );
}
