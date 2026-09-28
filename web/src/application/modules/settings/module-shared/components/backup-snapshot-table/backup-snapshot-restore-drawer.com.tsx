import { useId, useMemo, useState } from "react";

import { format } from "date-fns";
import { AlertTriangle, Loader2 } from "lucide-react";
import { FormProvider, useForm } from "react-hook-form";
import { toast } from "sonner";
import { APP_CONFIGURATION_QUERY_OPTIONS, AppStorageSettingsQueries } from "~/projects/data";
import {
    CommandConfigSection,
    type CommandFormInput,
    emptySourceCommand,
    hasCommand,
    isRelativeSubpath,
    mapCommandToFormInput,
    mapSourceCommandToPayload,
} from "~/projects/module-shared/components";
import type { BackupSnapshot_Restore_Payload } from "~/settings/api/services";
import { BackupSnapshotCommands } from "~/settings/data/commands";
import { BackupSnapshotQueries } from "~/settings/data/queries";
import {
    BACKUP_RESTORE_MODE,
    type BackupRestoreMode,
    type BackupSnapshot,
    type BackupSnapshotScope,
} from "~/settings/domain";

import { Combobox } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { formatDataSizeCompact } from "@application/shared/utils/data-size";

import { Button, Checkbox, Input, Tabs, TabsList, TabsTrigger } from "@/components/ui";
import { Sheet, SheetContent, SheetDescription, SheetFooter, SheetHeader, SheetTitle } from "@/components/ui/sheet";

import { BackupSnapshotPathPicker } from "./backup-snapshot-path-picker.com";
import { BackupSnapshotRestoreTarget, type RestoreTarget } from "./backup-snapshot-restore-target.com";
import { restoreKindNeedsListing, restoreKindOf } from "./backup-snapshot-restore.helpers";

interface VolumeOption extends Record<string, unknown> {
    id: string;
    name: string;
}

const RESTORE_COMMAND_EXAMPLE = "psql -U $POSTGRES_USER $POSTGRES_DB";

function Row({ label, children }: { label: React.ReactNode; children: React.ReactNode }) {
    return (
        <div className="flex flex-col gap-1.5">
            <span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">{label}</span>
            <div className="text-sm text-foreground">{children}</div>
        </div>
    );
}

/** The snapshot's own app, when the view's viewer can pick it. */
function defaultTarget(scope: BackupSnapshotScope, snapshot: BackupSnapshot): RestoreTarget | null {
    const { app } = snapshot;
    if (scope.type === "app") {
        return {
            id: scope.appId,
            name: app?.id === scope.appId ? app.name : "",
            projectId: scope.projectId,
            env: scope.env,
        };
    }
    if (!app || app.deleted || !app.projectId) {
        return null;
    }
    if (scope.type === "project" && app.projectId !== scope.projectId) {
        return null;
    }
    return { id: app.id, name: app.name, projectId: app.projectId, env: app.env };
}

/** Restoring a snapshot into an app: its target, how, and a confirmation. */
export function BackupSnapshotRestoreDrawer({ scope, snapshot, onOpenChange }: Props) {
    return (
        <Sheet
            open={Boolean(snapshot)}
            onOpenChange={onOpenChange}
        >
            <SheetContent className="sm:max-w-[680px] overflow-y-auto">
                {snapshot && (
                    <RestoreForm
                        key={snapshot.id}
                        scope={scope}
                        snapshot={snapshot}
                        onDone={() => {
                            onOpenChange(false);
                        }}
                    />
                )}
            </SheetContent>
        </Sheet>
    );
}

function RestoreForm({
    scope,
    snapshot,
    onDone,
}: {
    scope: BackupSnapshotScope;
    snapshot: BackupSnapshot;
    onDone: () => void;
}) {
    const { navigate } = useAppNavigate();
    const { job } = snapshot;

    const needsListing = restoreKindNeedsListing(snapshot);
    const { data: rootData, isFetching: isListing } = BackupSnapshotQueries.useFindEntries(
        { scope, id: snapshot.id, path: "" },
        { enabled: needsListing },
    );
    const kind = restoreKindOf(snapshot, needsListing ? rootData?.data : []);

    const [target, setTarget] = useState<RestoreTarget | null>(() => defaultTarget(scope, snapshot));
    const isSnapshotsApp = Boolean(target && snapshot.app && !snapshot.app.deleted && target.id === snapshot.app.id);
    const [volumeId, setVolumeId] = useState(isSnapshotsApp ? (job?.sourceVolumeId ?? "") : "");
    const [subpath, setSubpath] = useState(isSnapshotsApp ? (job?.sourceVolumeSubpath ?? "") : "");
    const [snapshotPath, setSnapshotPath] = useState("");
    const [pathSize, setPathSize] = useState(snapshot.sizeBytes);
    const [mode, setMode] = useState<BackupRestoreMode>(BACKUP_RESTORE_MODE.Replace);
    const [stopApp, setStopApp] = useState(true);
    const [typed, setTyped] = useState("");
    const stopAppId = useId();

    const form = useForm<{ command: CommandFormInput }>({
        defaultValues: {
            command: job?.restoreCommand ? mapCommandToFormInput(job.restoreCommand) : emptySourceCommand(),
        },
    });

    const isVolume = kind !== undefined && !kind.command;
    const { data: storageData, isFetching: isStorageFetching } = AppStorageSettingsQueries.useFindOne(
        { projectID: target?.projectId ?? "", env: target?.env ?? "", appID: target?.id ?? "" },
        { ...APP_CONFIGURATION_QUERY_OPTIONS, enabled: isVolume && Boolean(target) },
    );
    // The volumes the app mounts as its own directory: where a volume's snapshot can go.
    const volumeOptions = useMemo(() => {
        const seen = new Set<string>();
        return (storageData?.data.mounts ?? []).flatMap(mount => {
            if (!mount.volumeId || mount.sourceApp?.appId || seen.has(mount.volumeId)) {
                return [];
            }
            seen.add(mount.volumeId);
            const label = mount.target ?? mount.volumeId;
            return [{ value: { id: mount.volumeId, name: label } satisfies VolumeOption, label }];
        });
    }, [storageData]);
    const borrowers = storageData?.data.borrowedBy ?? [];

    const { mutate: restore, isPending } = BackupSnapshotCommands.useRestore({
        onSuccess: response => {
            toast.success("Restore started");
            onDone();
            if (target) {
                navigate.modules(
                    ROUTE.projects.single.apps.single.tasks.details.$route(
                        target.projectId,
                        target.env,
                        target.id,
                        response.data.taskId,
                    ),
                );
            }
        },
    });

    const replace = mode === BACKUP_RESTORE_MODE.Replace;
    const commandValues = form.watch("command");
    const problems: string[] = [];
    if (!target) {
        problems.push("Pick the app to restore into.");
    }
    if (kind?.command && !hasCommand(commandValues)) {
        problems.push("Give the command that loads the file.");
    }
    if (isVolume && !volumeId) {
        problems.push("Pick a volume the app mounts.");
    }
    if (isVolume && !isRelativeSubpath(subpath.trim())) {
        problems.push("The path is inside the volume: no leading '/', never above it with '..'.");
    }
    const confirmed = Boolean(target) && typed.trim() === target?.name;

    function submit() {
        if (!target || !kind) {
            return;
        }
        const payload: BackupSnapshot_Restore_Payload = kind.command
            ? { targetApp: { id: target.id }, command: mapSourceCommandToPayload(form.getValues().command) }
            : {
                  targetApp: { id: target.id },
                  volume: { id: volumeId },
                  subpath: subpath.trim(),
                  snapshotPath,
                  stopApp: replace || stopApp,
                  mode,
              };
        restore({ scope, id: snapshot.id, payload });
    }

    return (
        <>
            <SheetHeader>
                <SheetTitle>Restore Snapshot {snapshot.shortId}</SheetTitle>
                <SheetDescription>
                    {format(snapshot.time, "yyyy-MM-dd HH:mm:ss")} · {snapshot.repo.name || snapshot.repo.id} ·{" "}
                    {formatDataSizeCompact(snapshot.sizeBytes)}
                    {snapshot.app && ` · ${snapshot.app.deleted ? "deleted app" : snapshot.app.name}`}
                    {job && ` · ${job.deleted ? "deleted job" : job.name}`}
                </SheetDescription>
            </SheetHeader>

            <div className="flex flex-col gap-5 px-4 pb-4">
                <Row label="Restore into">
                    <BackupSnapshotRestoreTarget
                        scope={scope}
                        value={target}
                        onChange={next => {
                            setTarget(next);
                            setTyped("");
                            const own = Boolean(next && snapshot.app && next.id === snapshot.app.id);
                            setVolumeId(own ? (job?.sourceVolumeId ?? "") : "");
                            setSubpath(own ? (job?.sourceVolumeSubpath ?? "") : "");
                        }}
                        disabled={isPending}
                    />
                    {target && snapshot.app && target.id !== snapshot.app.id && (
                        <p className="mt-1.5 text-xs text-amber-700 dark:text-amber-400">
                            The data of {snapshot.app.deleted ? "a deleted app" : snapshot.app.name} is loaded into{" "}
                            {target.name}.
                        </p>
                    )}
                </Row>

                {kind === undefined && (
                    <p className="flex items-center gap-2 text-sm text-muted-foreground">
                        {isListing && <Loader2 className="size-4 animate-spin" />}
                        Reading what the snapshot holds...
                    </p>
                )}

                {kind?.command && (
                    <FormProvider {...form}>
                        <p className="text-sm text-muted-foreground">
                            The command runs in a container of the app, and reads{" "}
                            <span className="font-mono">{kind.fileName}</span> on its stdin, without a TTY. For example:{" "}
                            <span className="font-mono">{RESTORE_COMMAND_EXAMPLE}</span>
                        </p>
                        <CommandConfigSection
                            label="Restore Command"
                            fieldPrefix="command"
                            showLoadTemplate={Boolean(target)}
                            templateProjectId={target?.projectId ?? ""}
                            templateEnv={target?.env}
                            readOnly={isPending}
                            showArgGroups
                        />
                    </FormProvider>
                )}

                {isVolume && (
                    <>
                        <Row label="Volume">
                            <Combobox<VolumeOption>
                                options={volumeOptions}
                                value={volumeId || null}
                                onChange={(_, option) => {
                                    setVolumeId(option?.id ?? "");
                                }}
                                placeholder="Select a mounted volume"
                                emptyText={target ? "The app mounts no volume of its own" : "Pick an app first"}
                                closeOnSelect
                                loading={isStorageFetching}
                                disabled={isPending || !target}
                            />
                        </Row>
                        <Row label="Path in the volume">
                            <Input
                                value={subpath}
                                onChange={event => {
                                    setSubpath(event.target.value);
                                }}
                                placeholder="Empty for all of it, as the app sees it"
                                className="font-mono"
                                disabled={isPending}
                            />
                        </Row>
                        <Row label="What to restore">
                            <BackupSnapshotPathPicker
                                scope={scope}
                                snapshotRecordId={snapshot.id}
                                snapshotSize={snapshot.sizeBytes}
                                value={snapshotPath}
                                onChange={(path, size) => {
                                    setSnapshotPath(path);
                                    setPathSize(size);
                                }}
                                disabled={isPending}
                            />
                        </Row>
                        <Row label="How">
                            <Tabs
                                value={mode}
                                onValueChange={value => {
                                    setMode(value as BackupRestoreMode);
                                }}
                            >
                                <TabsList>
                                    <TabsTrigger
                                        value={BACKUP_RESTORE_MODE.Replace}
                                        disabled={isPending}
                                    >
                                        Replace
                                    </TabsTrigger>
                                    <TabsTrigger
                                        value={BACKUP_RESTORE_MODE.Overwrite}
                                        disabled={isPending}
                                    >
                                        Overwrite
                                    </TabsTrigger>
                                </TabsList>
                            </Tabs>
                            <p className="mt-1.5 text-xs text-muted-foreground">
                                {replace
                                    ? "The directory as it is now is moved aside, and the snapshot is restored into an empty one: the state at backup time. The old directory is kept as ….before-restore-…, for going back; delete it once you are sure."
                                    : "The snapshot's files are written over what is there. Files made after the backup stay, and nothing is kept aside."}
                            </p>
                            {replace && (
                                <p className="mt-1 text-xs text-muted-foreground">
                                    This needs {formatDataSizeCompact(pathSize)} of free space on the volume.
                                </p>
                            )}
                        </Row>
                        <Row label="The app">
                            <div className="flex items-center gap-2">
                                <Checkbox
                                    id={stopAppId}
                                    checked={replace || stopApp}
                                    onCheckedChange={checked => {
                                        setStopApp(checked === true);
                                    }}
                                    disabled={isPending || replace}
                                />
                                <label htmlFor={stopAppId}>Stop the app while restoring</label>
                            </div>
                            <p className="mt-1.5 text-xs text-muted-foreground">
                                {replace
                                    ? "Replace always stops the app: its containers would go on writing to the directory moved aside."
                                    : "A running app can write while its files are restored, and a database's files can end up broken. Leave it running only for static files, such as uploads."}
                            </p>
                            {borrowers.length > 0 && (
                                <p className="mt-1 text-xs text-amber-700 dark:text-amber-400">
                                    These apps were given a directory of {target?.name}&apos;s storage, and are not
                                    stopped: {borrowers.map(borrower => borrower.name || borrower.appId).join(", ")}.
                                </p>
                            )}
                        </Row>
                    </>
                )}

                {kind && (
                    <Row label="Confirm">
                        <div className="flex items-center gap-2.5 rounded-md border border-destructive bg-destructive/10 px-3.5 py-2.5 text-sm text-destructive font-medium">
                            <AlertTriangle className="size-4 shrink-0" />
                            <span>
                                The restore changes {target?.name ? `${target.name}'s` : "the app's"} data. Type the
                                app&apos;s name to confirm.
                            </span>
                        </div>
                        <Input
                            value={typed}
                            onChange={event => {
                                setTyped(event.target.value);
                            }}
                            placeholder={target?.name ?? ""}
                            className="mt-2"
                            disabled={isPending || !target}
                        />
                        {problems.length > 0 && (
                            <ul className="mt-2 list-disc pl-5 text-xs text-muted-foreground">
                                {problems.map(problem => (
                                    <li key={problem}>{problem}</li>
                                ))}
                            </ul>
                        )}
                    </Row>
                )}
            </div>

            <SheetFooter className="flex-row justify-end gap-2">
                <Button
                    variant="outline"
                    disabled={isPending}
                    onClick={onDone}
                >
                    Cancel
                </Button>
                <Button
                    variant="destructive"
                    isLoading={isPending}
                    disabled={!kind || problems.length > 0 || !confirmed}
                    onClick={submit}
                >
                    Restore
                </Button>
            </SheetFooter>
        </>
    );
}

interface Props {
    scope: BackupSnapshotScope;
    snapshot: BackupSnapshot | null;
    onOpenChange: (open: boolean) => void;
}
