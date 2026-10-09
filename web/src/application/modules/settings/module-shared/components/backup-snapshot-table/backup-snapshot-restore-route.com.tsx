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
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";
import type { BackupSnapshot_Restore_Payload } from "~/settings/api/services";
import { BackupSnapshotCommands } from "~/settings/data/commands";
import { BackupSnapshotQueries } from "~/settings/data/queries";
import {
    BACKUP_RESTORE_MODE,
    type BackupRestoreMode,
    type BackupSnapshot,
    type BackupSnapshotScope,
} from "~/settings/domain";
import { SettingsFormRouteHeader } from "~/settings/module-shared/components/settings-form-route-header";

import {
    AppLoader,
    Combobox,
    ContentBlock,
    FormActionBar,
    InfoBlock,
    LabelWithInfo,
} from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { formatDataSizeCompact } from "@application/shared/utils/data-size";

import { Button, Checkbox, Input, Tabs, TabsList, TabsTrigger } from "@/components/ui";

import { BackupSnapshotPathPicker } from "./backup-snapshot-path-picker.com";
import { BackupSnapshotRestoreTarget, type RestoreTarget } from "./backup-snapshot-restore-target.com";
import { restoreKindNeedsListing, restoreKindOf, snapshotListRoute } from "./backup-snapshot-restore.helpers";

const TITLE_WIDTH = 220;
/** The form controls' width, as the project forms have it. */
const CONTROL_CLASS = PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS;
const RESTORE_COMMAND_EXAMPLE = "psql -U $POSTGRES_USER $POSTGRES_DB";

interface VolumeOption extends Record<string, unknown> {
    id: string;
    name: string;
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

/** Restoring a snapshot into an app, as a page of its own: its target, how, and a confirmation. */
export function BackupSnapshotRestoreRoute({ scope, snapshotId }: Props) {
    const { data, isLoading } = BackupSnapshotQueries.useFindOneById({ scope, id: snapshotId });
    const snapshot = data?.data;

    return (
        <div className="flex w-full flex-col">
            <SettingsFormRouteHeader title={snapshot ? `Restore Snapshot ${snapshot.shortId}` : "Restore Snapshot"} />
            {isLoading && (
                <div className="flex min-h-[220px] items-center justify-center">
                    <AppLoader />
                </div>
            )}
            {snapshot && (
                <RestoreForm
                    key={snapshot.id}
                    scope={scope}
                    snapshot={snapshot}
                />
            )}
        </div>
    );
}

function RestoreForm({ scope, snapshot }: { scope: BackupSnapshotScope; snapshot: BackupSnapshot }) {
    const { navigate } = useAppNavigate();
    const { job } = snapshot;
    const stopAppId = useId();

    function backToList() {
        navigate.modules(snapshotListRoute(scope), { ignorePrevPath: true });
    }

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
            if (target) {
                navigate.modules(
                    ROUTE.projects.single.apps.single.tasks.details.$route(
                        target.projectId,
                        target.env,
                        target.id,
                        response.data.taskId,
                    ),
                    { ignorePrevPath: true },
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
    // An app whose name is not known cannot be confirmed by typing nothing.
    const confirmed = Boolean(target?.name) && typed.trim() === target?.name;

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
        <fieldset
            disabled={isPending}
            className="flex flex-col gap-6"
        >
            <ContentBlock label="Snapshot">
                <div className="flex flex-col gap-4 text-sm">
                    <InfoBlock
                        title="Taken"
                        titleWidth={TITLE_WIDTH}
                    >
                        {format(snapshot.time, "yyyy-MM-dd HH:mm:ss")}{" "}
                        <span className="font-mono text-muted-foreground">{snapshot.shortId}</span>
                    </InfoBlock>
                    <InfoBlock
                        title="Repository"
                        titleWidth={TITLE_WIDTH}
                    >
                        {snapshot.repo.name || snapshot.repo.id}
                    </InfoBlock>
                    <InfoBlock
                        title="Size"
                        titleWidth={TITLE_WIDTH}
                    >
                        {formatDataSizeCompact(snapshot.sizeBytes)}
                    </InfoBlock>
                    {snapshot.app && (
                        <InfoBlock
                            title="App"
                            titleWidth={TITLE_WIDTH}
                        >
                            {snapshot.app.deleted
                                ? `Deleted app ${snapshot.app.id}`
                                : `${snapshot.app.name} (${snapshot.app.env})`}
                        </InfoBlock>
                    )}
                    {job && (
                        <InfoBlock
                            title="Job"
                            titleWidth={TITLE_WIDTH}
                        >
                            {job.deleted ? `Deleted job ${job.id}` : job.name}
                        </InfoBlock>
                    )}
                </div>
            </ContentBlock>

            <ContentBlock label="Target">
                <div className="flex flex-col gap-4">
                    <InfoBlock
                        title={
                            <LabelWithInfo
                                label="Restore Into"
                                isRequired
                                content="The app the data goes into: the snapshot's own, or another, such as a staging copy to try the backup on."
                            />
                        }
                        titleWidth={TITLE_WIDTH}
                    >
                        <div className={CONTROL_CLASS}>
                            <BackupSnapshotRestoreTarget
                                scope={scope}
                                value={target}
                                onChange={next => {
                                    setTarget(next);
                                    setTyped("");
                                    const own = Boolean(next && next.id === snapshot.app?.id);
                                    setVolumeId(own ? (job?.sourceVolumeId ?? "") : "");
                                    setSubpath(own ? (job?.sourceVolumeSubpath ?? "") : "");
                                }}
                                disabled={isPending}
                            />
                            {target && snapshot.app && target.id !== snapshot.app.id && (
                                <p className="mt-1.5 text-xs text-amber-700 dark:text-amber-400">
                                    The data of {snapshot.app.deleted ? "a deleted app" : snapshot.app.name} is loaded
                                    into {target.name}.
                                </p>
                            )}
                        </div>
                    </InfoBlock>

                    {kind === undefined && (
                        <p className="flex items-center gap-2 text-sm text-muted-foreground">
                            {isListing && <Loader2 className="size-4 animate-spin" />}
                            Reading what the snapshot holds...
                        </p>
                    )}

                    {isVolume && (
                        <>
                            <InfoBlock
                                title={
                                    <LabelWithInfo
                                        label="Volume"
                                        isRequired
                                        content="A volume the app mounts as its own directory."
                                    />
                                }
                                titleWidth={TITLE_WIDTH}
                            >
                                <Combobox<VolumeOption>
                                    options={volumeOptions}
                                    value={volumeId || null}
                                    onChange={(_, option) => {
                                        setVolumeId(option?.id ?? "");
                                    }}
                                    placeholder="Select a mounted volume"
                                    emptyText={target ? "The app mounts no volume of its own" : "Pick an app first"}
                                    className={CONTROL_CLASS}
                                    closeOnSelect
                                    loading={isStorageFetching}
                                    disabled={isPending || !target}
                                />
                            </InfoBlock>
                            <InfoBlock
                                title={
                                    <LabelWithInfo
                                        label="Path"
                                        content="Where the snapshot's root goes: a path inside the volume as the app sees it; empty for all of it."
                                    />
                                }
                                titleWidth={TITLE_WIDTH}
                            >
                                <Input
                                    value={subpath}
                                    onChange={event => {
                                        setSubpath(event.target.value);
                                    }}
                                    placeholder="uploads"
                                    className={`${CONTROL_CLASS} font-mono`}
                                />
                            </InfoBlock>
                        </>
                    )}
                </div>
            </ContentBlock>

            {kind?.command && (
                <FormProvider {...form}>
                    <p className="px-4 text-sm text-muted-foreground">
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
                <ContentBlock label="Restore">
                    <div className="flex flex-col gap-6">
                        <InfoBlock
                            title={
                                <LabelWithInfo
                                    label="What to Restore"
                                    content="All of the snapshot, or one of its directories: it goes to the same place under the path."
                                />
                            }
                            titleWidth={TITLE_WIDTH}
                        >
                            <div className={CONTROL_CLASS}>
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
                            </div>
                        </InfoBlock>
                        <InfoBlock
                            title="How"
                            titleWidth={TITLE_WIDTH}
                        >
                            <div className={`flex flex-col gap-2 ${CONTROL_CLASS}`}>
                                <Tabs
                                    value={mode}
                                    onValueChange={value => {
                                        setMode(value as BackupRestoreMode);
                                    }}
                                >
                                    <TabsList>
                                        <TabsTrigger value={BACKUP_RESTORE_MODE.Replace}>Replace</TabsTrigger>
                                        <TabsTrigger value={BACKUP_RESTORE_MODE.Overwrite}>Overwrite</TabsTrigger>
                                    </TabsList>
                                </Tabs>
                                <p className="text-sm text-muted-foreground">
                                    {replace
                                        ? "The directory as it is now is moved aside, and the snapshot is restored into an empty one: the state at backup time. The old directory is kept as ….before-restore-…, for going back; delete it once you are sure."
                                        : "The snapshot's files are written over what is there. Files made after the backup stay, and nothing is kept aside."}
                                </p>
                                {replace && (
                                    <p className="text-sm text-muted-foreground">
                                        This needs {formatDataSizeCompact(pathSize)} of free space on the volume.
                                    </p>
                                )}
                            </div>
                        </InfoBlock>
                        <InfoBlock
                            title="The App"
                            titleWidth={TITLE_WIDTH}
                        >
                            <div className={`flex flex-col gap-2 ${CONTROL_CLASS}`}>
                                <div className="flex items-center gap-2 text-sm">
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
                                <p className="text-sm text-muted-foreground">
                                    {replace
                                        ? "Replace always stops the app: its containers would go on writing to the directory moved aside."
                                        : "A running app can write while its files are restored, and a database's files can end up broken. Leave it running only for static files, such as uploads."}
                                </p>
                                {borrowers.length > 0 && (
                                    <p className="text-sm text-amber-700 dark:text-amber-400">
                                        These apps were given a directory of {target?.name}&apos;s storage, and are not
                                        stopped: {borrowers.map(borrower => borrower.name || borrower.appId).join(", ")}
                                        .
                                    </p>
                                )}
                            </div>
                        </InfoBlock>
                    </div>
                </ContentBlock>
            )}

            {kind && (
                <ContentBlock label="Confirm">
                    <InfoBlock
                        title="App Name"
                        titleWidth={TITLE_WIDTH}
                    >
                        <div className={`flex flex-col gap-2 ${CONTROL_CLASS}`}>
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
                                disabled={isPending || !target}
                            />
                            {problems.length > 0 && (
                                <ul className="list-disc pl-5 text-sm text-muted-foreground">
                                    {problems.map(problem => (
                                        <li key={problem}>{problem}</li>
                                    ))}
                                </ul>
                            )}
                        </div>
                    </InfoBlock>
                </ContentBlock>
            )}

            <FormActionBar>
                <Button
                    type="button"
                    variant="outline"
                    className="min-w-[100px]"
                    disabled={isPending}
                    onClick={backToList}
                >
                    Cancel
                </Button>
                <Button
                    type="button"
                    variant="destructive"
                    className="min-w-[100px]"
                    isLoading={isPending}
                    disabled={!kind || problems.length > 0 || !confirmed}
                    onClick={submit}
                >
                    Restore
                </Button>
            </FormActionBar>
        </fieldset>
    );
}

interface Props {
    scope: BackupSnapshotScope;
    /** The snapshot's record. */
    snapshotId: string;
}
