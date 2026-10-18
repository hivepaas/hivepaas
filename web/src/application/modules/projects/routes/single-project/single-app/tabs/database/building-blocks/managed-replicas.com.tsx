import { useState } from "react";

import { CopyPlus, Loader2, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { AppManagedCommands, ProjectClusterVolumesQueries } from "~/projects/data";
import type { AppManagedMemberStatus, ManagedPlacement } from "~/projects/domain";
import { formatManagedTime } from "~/projects/module-shared/utils";

import { AppLink, PopConfirm } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { formatDataSizeCompact } from "@application/shared/utils/data-size";

import { Badge, type StatusMeaning, statusClassName } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

const LIST_ALL_PAGE = { page: 1, size: 1000 };

interface Props {
    projectId: string;
    env: string;
    appId: string;
    members: AppManagedMemberStatus[];
    placement: ManagedPlacement;
    canWrite: boolean;
}

/**
 * A managed database's read replicas: each with its node and how far behind it
 * is, as the database's last health check saw it; adding one on a pinned
 * volume, and removing one.
 */
export function ManagedReplicas({ projectId, env, appId, members, placement, canWrite }: Props) {
    const scope = { projectID: projectId, env, appID: appId };
    const [volume, setVolume] = useState("");
    const offSite = placement === "off-site";
    const { data: volumesData } = ProjectClusterVolumesQueries.useFindManyPaginated(
        { projectID: projectId, env, pagination: LIST_ALL_PAGE },
        { enabled: canWrite && offSite },
    );
    const volumes = (volumesData?.data ?? []).filter(one => Boolean(one.nodeId) || Boolean(one.nodeLabel));

    const { mutate: addReplica, isPending: isAdding } = AppManagedCommands.useAddReplica({
        onSuccess: () => {
            toast.success("Replica added: it is seeded from the archive, then streams");
            setVolume("");
        },
    });
    const { mutate: removeReplica, isPending: isRemoving } = AppManagedCommands.useRemoveReplica({
        onSuccess: () => {
            toast.success("Replica removed");
        },
    });

    return (
        <section
            aria-label="Replicas"
            className="flex flex-col gap-3"
        >
            <div className="flex flex-col">
                <h2 className="text-sm font-semibold">Replicas</h2>
                <p className="text-xs text-muted-foreground">
                    Read-only copies that follow this database, each on a volume of its own: seeded from the archive,
                    then streaming from it. Only this database writes, and archives.
                </p>
            </div>
            <div className="flex flex-col divide-y rounded-md border">
                {members.length === 0 && <p className="px-3 py-3 text-xs text-muted-foreground">No replica yet.</p>}
                {members.map(member => {
                    const memberId = member.app?.id ?? "";
                    return (
                        <div
                            key={member.key}
                            role="row"
                            aria-label={member.key}
                            className="flex flex-wrap items-center justify-between gap-3 px-3 py-2.5 text-xs"
                        >
                            <div className="flex flex-col gap-0.5">
                                {memberId ? (
                                    <AppLink.Basic
                                        to={ROUTE.projects.single.apps.single.database.$route(projectId, env, memberId)}
                                        className="text-sm font-medium underline"
                                    >
                                        {member.key}
                                    </AppLink.Basic>
                                ) : (
                                    <span className="text-sm font-medium">{member.key}</span>
                                )}
                                <span className="text-muted-foreground">
                                    On node {member.node || "unknown"}
                                    {member.checkedAt ? `, as of ${formatManagedTime(member.checkedAt)}` : ""}
                                </span>
                            </div>
                            <div className="flex items-center gap-3">
                                <Badge className={statusClassName(stateMeaning(member.state))}>{member.state}</Badge>
                                <span className="tabular-nums text-muted-foreground">{lagText(member)}</span>
                                <PopConfirm
                                    title={`Remove ${member.key}?`}
                                    description="Its app is deleted, with its copy of the data."
                                    confirmText="Remove"
                                    onConfirm={() => {
                                        removeReplica({ ...scope, memberID: memberId });
                                    }}
                                >
                                    <Button
                                        variant="ghost"
                                        size="icon-sm"
                                        aria-label={`Remove ${member.key}`}
                                        disabled={!canWrite || isRemoving || memberId === ""}
                                    >
                                        <Trash2 className="size-3.5" />
                                    </Button>
                                </PopConfirm>
                            </div>
                        </div>
                    );
                })}
                <div className="flex flex-wrap items-center gap-2 px-3 py-2.5">
                    {offSite ? (
                        <>
                            <Select
                                value={volume}
                                onValueChange={setVolume}
                                disabled={!canWrite || isAdding}
                            >
                                <SelectTrigger
                                    className="h-8 w-full max-w-[320px] text-xs"
                                    aria-label="Replica volume"
                                >
                                    <SelectValue placeholder="A volume pinned to a node" />
                                </SelectTrigger>
                                <SelectContent>
                                    {volumes.map(one => (
                                        <SelectItem
                                            key={one.id}
                                            value={one.id}
                                            className="text-xs"
                                        >
                                            {one.name}
                                        </SelectItem>
                                    ))}
                                    {volumes.length === 0 && (
                                        <div className="px-2 py-1.5 text-xs text-muted-foreground">
                                            No cluster volume pinned to a node in this environment.
                                        </div>
                                    )}
                                </SelectContent>
                            </Select>
                            <Button
                                size="sm"
                                variant="outline"
                                disabled={!canWrite || isAdding || volume === ""}
                                onClick={() => {
                                    addReplica({ ...scope, payload: { dataVolume: { id: volume } } });
                                }}
                            >
                                {isAdding ? (
                                    <Loader2 className="size-3.5 animate-spin" />
                                ) : (
                                    <CopyPlus className="size-3.5" />
                                )}
                                Add replica
                            </Button>
                            <span className="text-xs text-muted-foreground">
                                Preferably on another node than this database&apos;s.
                            </span>
                        </>
                    ) : (
                        <span className="text-xs text-muted-foreground">
                            A replica is seeded from an archive in object storage, and this database archives to a
                            volume.
                        </span>
                    )}
                </div>
            </div>
        </section>
    );
}

function stateMeaning(state: string): StatusMeaning {
    switch (state) {
        case "streaming":
            return "success";
        case "disconnected":
            return "failure";
        case "unknown":
            return "inactive";
        default:
            return "attention";
    }
}

function lagText(member: AppManagedMemberStatus): string {
    if (member.state === "unknown" || member.state === "disconnected") {
        return "";
    }
    if (member.lagBytes <= 0) {
        return "caught up";
    }
    const behind = `${formatDataSizeCompact(member.lagBytes)} behind`;
    return member.lagSeconds === undefined ? behind : `${behind}, ${member.lagSeconds}s`;
}

/** On a replica: its primary, which its archive, backups and settings are. */
export function ManagedMemberBanner({
    projectId,
    env,
    primaryId,
}: {
    projectId: string;
    env: string;
    primaryId?: string;
}) {
    return (
        <div
            role="status"
            className="flex items-start gap-2 rounded-md border border-sky-500/30 bg-sky-500/10 px-3 py-2.5 text-xs text-sky-800 dark:text-sky-300"
        >
            <CopyPlus className="mt-0.5 size-3.5 shrink-0" />
            <span>
                A read replica of{" "}
                {primaryId ? (
                    <AppLink.Basic
                        to={ROUTE.projects.single.apps.single.database.$route(projectId, env, primaryId)}
                        className="underline"
                    >
                        its primary
                    </AppLink.Basic>
                ) : (
                    "its primary"
                )}
                , which it follows. Its archive, backups, restores and settings are the primary&apos;s, and so is
                removing it.
            </span>
        </div>
    );
}
