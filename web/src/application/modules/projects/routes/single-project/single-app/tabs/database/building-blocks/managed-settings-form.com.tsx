import { useEffect } from "react";

import { Tabs, TabsList, TabsTrigger } from "@components/ui/tabs";
import { Textarea } from "@components/ui/textarea";
import { zodResolver } from "@hookform/resolvers/zod";
import { AlertTriangle } from "lucide-react";
import { useController, useForm, useFormState, useWatch } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import type { AppManagedSettings_UpdatePayload } from "~/projects/api/services";
import {
    AppManagedCommands,
    ProjectBackupRepoQueries,
    ProjectClusterVolumesQueries,
    ProjectKeyAuthQueries,
} from "~/projects/data";
import { type AppManagedSettings, MANAGED_RESTORE_TEST_MODE } from "~/projects/domain";

import { CronTimezoneNote, InfoBlock } from "@application/shared/components";

import { Button } from "@/components/ui/button";
import { Field, FieldDescription, FieldError, FieldGroup } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

const TITLE_WIDTH = 200;
const LIST_ALL_PAGE = { page: 1, size: 1000 };
const MAX_RETAIN_FULL = 365;
const QUERY_MAX_LENGTH = 2000;
/** The verifier's key: none chosen is the database's own. */
const DATABASE_KEY = "__database__";
/** The dump's repository: none, dumps off. */
const NO_DUMP = "__no-dump__";
/** The base backups' host: the database itself. */
const THIS_DATABASE = "__this-database__";
/** The engine's dump schedule, which a database made before dumps were offered has not kept. */
const DEFAULT_DUMP_SCHEDULE = "0 3 * * *";

const SettingsFormSchema = z
    .object({
        retainFull: z.string().trim(),
        baseBackupSchedule: z.string().trim().min(1, "A cron expression"),
        restoreTestSchedule: z.string().trim().min(1, "A cron expression"),
        restoreTestQuery: z.string().trim().max(QUERY_MAX_LENGTH, `At most ${QUERY_MAX_LENGTH} characters`),
        restoreTestMode: z.nativeEnum(MANAGED_RESTORE_TEST_MODE),
        dumpRepo: z.string(),
        backupFrom: z.string(),
        dumpSchedule: z.string().trim().min(1, "A cron expression"),
        // Only for a move into a verifier: the volume it restores into, and its key.
        verifierVolume: z.string(),
        verifierKey: z.string(),
        movingToVerifier: z.boolean(),
    })
    .superRefine((value, ctx) => {
        const retainFull = Number(value.retainFull);
        if (!/^\d+$/.test(value.retainFull) || retainFull < 1 || retainFull > MAX_RETAIN_FULL) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                message: `A whole number, 1 to ${MAX_RETAIN_FULL}`,
                path: ["retainFull"],
            });
        }
        if (value.movingToVerifier && value.verifierVolume === "") {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                message: "Choose the volume the verifier restores into",
                path: ["verifierVolume"],
            });
        }
    });

type FormInput = z.input<typeof SettingsFormSchema>;
type FormOutput = z.output<typeof SettingsFormSchema>;

function initialValuesOf(settings: AppManagedSettings): FormInput {
    return {
        retainFull: String(settings.retainFull),
        baseBackupSchedule: settings.baseBackupSchedule,
        restoreTestSchedule: settings.restoreTestSchedule,
        restoreTestQuery: settings.restoreTest?.checkQuery ?? "",
        restoreTestMode: settings.restoreTest?.mode ?? MANAGED_RESTORE_TEST_MODE.container,
        dumpRepo: settings.dump.repo?.id ?? NO_DUMP,
        backupFrom: settings.backupFrom?.id ?? THIS_DATABASE,
        dumpSchedule: dumpScheduleOf(settings),
        verifierVolume: "",
        verifierKey: DATABASE_KEY,
        movingToVerifier: false,
    };
}

function dumpScheduleOf(settings: AppManagedSettings): string {
    return settings.dump.schedule === "" ? DEFAULT_DUMP_SCHEDULE : settings.dump.schedule;
}

interface Props {
    projectId: string;
    env: string;
    appId: string;
    settings: AppManagedSettings;
    readOnly: boolean;
}

/**
 * Keyed by the settings' updateVer where it is used: a save, or anyone else's,
 * gives a fresh form.
 *
 * What a person may change of a managed database's engine: how many base
 * backups it keeps, when its jobs run, what the restore test checks and where
 * it runs. The rest is the engine's.
 */
export function ManagedSettingsForm({ projectId, env, appId, settings, readOnly }: Props) {
    const {
        handleSubmit,
        control,
        setValue,
        formState: { errors },
    } = useForm<FormInput, unknown, FormOutput>({
        defaultValues: initialValuesOf(settings),
        resolver: zodResolver(SettingsFormSchema),
        mode: "onSubmit",
    });
    const { isDirty } = useFormState({ control });

    const currentMode = settings.restoreTest?.mode ?? MANAGED_RESTORE_TEST_MODE.container;
    const mode = useWatch({ control, name: "restoreTestMode" });
    const movingToVerifier = currentMode !== MANAGED_RESTORE_TEST_MODE.verifier && mode === "verifier";
    const movingToContainer = currentMode === MANAGED_RESTORE_TEST_MODE.verifier && mode === "container";
    useEffect(() => {
        setValue("movingToVerifier", movingToVerifier);
    }, [movingToVerifier, setValue]);

    const { data: volumesData } = ProjectClusterVolumesQueries.useFindManyPaginated(
        { projectID: projectId, env, pagination: LIST_ALL_PAGE },
        { enabled: movingToVerifier },
    );
    const { data: keyAuthsData } = ProjectKeyAuthQueries.useFindManyPaginated(
        { projectID: projectId, pagination: LIST_ALL_PAGE },
        { enabled: movingToVerifier },
    );
    const volumes = (volumesData?.data ?? []).filter(volume => Boolean(volume.nodeId) || Boolean(volume.nodeLabel));
    const keyAuths = keyAuthsData?.data ?? [];
    // The environment's: a database dumps into what its environment sees.
    const { data: reposData } = ProjectBackupRepoQueries.useFindManyPaginated({
        projectID: projectId,
        env,
        pagination: LIST_ALL_PAGE,
    });
    const repos = reposData?.data ?? [];
    const dumpRepoId = settings.dump.repo?.id;
    const dumpRepoListed = dumpRepoId === undefined || repos.some(repo => repo.id === dumpRepoId);

    const { field: retainFull } = useController({ name: "retainFull", control });
    const { field: baseBackupSchedule } = useController({ name: "baseBackupSchedule", control });
    const { field: restoreTestSchedule } = useController({ name: "restoreTestSchedule", control });
    const { field: restoreTestQuery } = useController({ name: "restoreTestQuery", control });
    const { field: restoreTestMode } = useController({ name: "restoreTestMode", control });
    const { field: dumpRepo } = useController({ name: "dumpRepo", control });
    const { field: backupFrom } = useController({ name: "backupFrom", control });
    const { field: dumpSchedule } = useController({ name: "dumpSchedule", control });
    const dumpScheduleValue = useWatch({ control, name: "dumpSchedule" });
    const baseBackupScheduleValue = useWatch({ control, name: "baseBackupSchedule" });
    const { field: verifierVolume } = useController({ name: "verifierVolume", control });
    const { field: verifierKey } = useController({ name: "verifierKey", control });

    const { mutate: updateSettings, isPending } = AppManagedCommands.useUpdateSettings({
        onSuccess: () => {
            toast.success("Database settings saved");
        },
    });

    function onValid(values: FormOutput) {
        if (readOnly) {
            return;
        }

        const payload: AppManagedSettings_UpdatePayload = { updateVer: settings.updateVer };
        const retain = Number(values.retainFull);
        if (retain !== settings.retainFull) {
            payload.retainFull = retain;
        }
        if (values.baseBackupSchedule !== settings.baseBackupSchedule) {
            payload.baseBackupSchedule = values.baseBackupSchedule;
        }
        if (values.restoreTestSchedule !== settings.restoreTestSchedule) {
            payload.restoreTestSchedule = values.restoreTestSchedule;
        }
        if (values.restoreTestQuery !== (settings.restoreTest?.checkQuery ?? "")) {
            payload.restoreTestQuery = values.restoreTestQuery;
        }
        if (values.backupFrom !== (settings.backupFrom?.id ?? THIS_DATABASE)) {
            payload.backupFrom = { id: values.backupFrom === THIS_DATABASE ? "" : values.backupFrom };
        }
        if (values.dumpRepo !== (settings.dump.repo?.id ?? NO_DUMP)) {
            payload.dumpRepo = { id: values.dumpRepo === NO_DUMP ? "" : values.dumpRepo };
        }
        if (values.dumpSchedule !== dumpScheduleOf(settings)) {
            payload.dumpSchedule = values.dumpSchedule;
        }
        if (values.restoreTestMode !== currentMode) {
            payload.restoreTest =
                values.restoreTestMode === MANAGED_RESTORE_TEST_MODE.verifier
                    ? {
                          mode: values.restoreTestMode,
                          volume: { id: values.verifierVolume },
                          ...(values.verifierKey !== DATABASE_KEY ? { keyAuth: { id: values.verifierKey } } : {}),
                      }
                    : { mode: values.restoreTestMode };
        }
        updateSettings({ projectID: projectId, env, appID: appId, payload });
    }

    return (
        <section
            aria-label="Settings"
            className="flex flex-col gap-3"
        >
            <div className="flex flex-col">
                <h2 className="text-sm font-semibold">Settings</h2>
                <p className="text-xs text-muted-foreground">
                    The archive, its prefix and its timeout stay as the database was made with.
                </p>
            </div>
            <form
                onSubmit={event => {
                    event.preventDefault();
                    if (readOnly) {
                        return;
                    }

                    void handleSubmit(onValid)(event);
                }}
                className="rounded-md border px-3 py-3"
            >
                <fieldset
                    disabled={readOnly || isPending}
                    className="contents"
                >
                    <FieldGroup>
                        <Field>
                            <InfoBlock
                                title="Full backups kept"
                                titleWidth={TITLE_WIDTH}
                            >
                                <Input
                                    {...retainFull}
                                    aria-label="Full backups kept"
                                    inputMode="numeric"
                                    className="max-w-[160px]"
                                    aria-invalid={Boolean(errors.retainFull)}
                                />
                                <FieldDescription>
                                    How far back a restore can go, in base backups: about a day each. A kept restore
                                    point holds more back.
                                </FieldDescription>
                                <FieldError errors={[errors.retainFull]} />
                            </InfoBlock>
                        </Field>

                        <Field>
                            <InfoBlock
                                title="Base backup schedule"
                                titleWidth={TITLE_WIDTH}
                            >
                                <Input
                                    {...baseBackupSchedule}
                                    aria-label="Base backup schedule"
                                    className="max-w-[240px] font-mono"
                                    aria-invalid={Boolean(errors.baseBackupSchedule)}
                                />
                                <FieldDescription>When the base backup and retention run.</FieldDescription>
                                <CronTimezoneNote />
                                <FieldError errors={[errors.baseBackupSchedule]} />
                            </InfoBlock>
                        </Field>

                        {settings.members.length > 0 && (
                            <Field>
                                <InfoBlock
                                    title="Base backups taken on"
                                    titleWidth={TITLE_WIDTH}
                                >
                                    <Select
                                        value={backupFrom.value}
                                        onValueChange={backupFrom.onChange}
                                    >
                                        <SelectTrigger
                                            className="w-full max-w-[360px]"
                                            aria-label="Base backups taken on"
                                        >
                                            <SelectValue />
                                        </SelectTrigger>
                                        <SelectContent>
                                            <SelectItem value={THIS_DATABASE}>This database</SelectItem>
                                            {settings.members.map(member => (
                                                <SelectItem
                                                    key={member.key}
                                                    value={member.app?.id ?? member.key}
                                                >
                                                    Replica {member.key}
                                                </SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>
                                    <FieldDescription>
                                        A replica takes the base backup, sparing this database its read; retention runs
                                        here either way. Removing the replica takes them back here.
                                    </FieldDescription>
                                </InfoBlock>
                            </Field>
                        )}

                        <Field>
                            <InfoBlock
                                title="Restore test schedule"
                                titleWidth={TITLE_WIDTH}
                            >
                                <Input
                                    {...restoreTestSchedule}
                                    aria-label="Restore test schedule"
                                    className="max-w-[240px] font-mono"
                                    aria-invalid={Boolean(errors.restoreTestSchedule)}
                                />
                                <FieldDescription>When the archive is restored, to prove it restores.</FieldDescription>
                                <CronTimezoneNote />
                                <FieldError errors={[errors.restoreTestSchedule]} />
                            </InfoBlock>
                        </Field>

                        <Field>
                            <InfoBlock
                                title="Restore test check"
                                titleWidth={TITLE_WIDTH}
                            >
                                <Textarea
                                    {...restoreTestQuery}
                                    aria-label="Restore test check"
                                    rows={2}
                                    className="font-mono text-xs"
                                    placeholder="SELECT count(*) FROM orders"
                                    aria-invalid={Boolean(errors.restoreTestQuery)}
                                />
                                <FieldDescription>
                                    A query the restored copy must answer, besides listing its databases. Empty for
                                    none.
                                </FieldDescription>
                                <FieldError errors={[errors.restoreTestQuery]} />
                            </InfoBlock>
                        </Field>

                        <Field>
                            <InfoBlock
                                title="Dump into"
                                titleWidth={TITLE_WIDTH}
                            >
                                <Select
                                    value={dumpRepo.value}
                                    onValueChange={dumpRepo.onChange}
                                >
                                    <SelectTrigger
                                        className="w-full max-w-[360px]"
                                        aria-label="Dump into"
                                    >
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value={NO_DUMP}>No dumps</SelectItem>
                                        {repos.map(repo => (
                                            <SelectItem
                                                key={repo.id}
                                                value={repo.id}
                                            >
                                                {repo.name}
                                            </SelectItem>
                                        ))}
                                        {!dumpRepoListed && (
                                            <SelectItem value={dumpRepoId}>
                                                A repository this environment no longer sees
                                            </SelectItem>
                                        )}
                                    </SelectContent>
                                </Select>
                                <FieldDescription>
                                    A pg_dump into a backup repository, beside the archive: a copy any PostgreSQL can
                                    load. Its snapshots are the app&apos;s backup snapshots, restored from there.
                                </FieldDescription>
                            </InfoBlock>
                        </Field>

                        <Field>
                            <InfoBlock
                                title="Dump schedule"
                                titleWidth={TITLE_WIDTH}
                            >
                                <Input
                                    {...dumpSchedule}
                                    aria-label="Dump schedule"
                                    className="max-w-[240px] font-mono"
                                    aria-invalid={Boolean(errors.dumpSchedule)}
                                />
                                <FieldDescription>
                                    When the dump runs, at any frequency: 0 */6 * * * is every 6 hours. Kept while dumps
                                    are off.
                                </FieldDescription>
                                {dumpScheduleValue.trim() !== "" &&
                                    dumpScheduleValue.trim() === baseBackupScheduleValue.trim() && (
                                        <FieldDescription className="text-amber-700 dark:text-amber-300">
                                            The same as the base backup&apos;s: both read the whole database at once.
                                        </FieldDescription>
                                    )}
                                <CronTimezoneNote />
                                <FieldError errors={[errors.dumpSchedule]} />
                            </InfoBlock>
                        </Field>

                        <Field>
                            <InfoBlock
                                title="Restore test runs in"
                                titleWidth={TITLE_WIDTH}
                            >
                                <Tabs
                                    value={restoreTestMode.value}
                                    onValueChange={value => {
                                        restoreTestMode.onChange(value);
                                    }}
                                >
                                    <TabsList>
                                        <TabsTrigger value={MANAGED_RESTORE_TEST_MODE.container}>
                                            The database&apos;s container
                                        </TabsTrigger>
                                        <TabsTrigger value={MANAGED_RESTORE_TEST_MODE.verifier}>
                                            A verifier app
                                        </TabsTrigger>
                                    </TabsList>
                                </Tabs>
                                <FieldDescription>
                                    {mode === MANAGED_RESTORE_TEST_MODE.verifier
                                        ? "An app of its own, ideally on another node, restores the archive with a " +
                                          "key that need not write: the database's disk and load are left alone."
                                        : "The database's container restores the archive beside the database, on " +
                                          "its disk."}
                                </FieldDescription>
                                {currentMode === MANAGED_RESTORE_TEST_MODE.verifier && settings.restoreTest?.nodeId && (
                                    <FieldDescription>
                                        The verifier runs on node {settings.restoreTest.nodeId}.
                                    </FieldDescription>
                                )}
                            </InfoBlock>
                        </Field>

                        {movingToVerifier && (
                            <>
                                <Field>
                                    <InfoBlock
                                        title="Verifier volume"
                                        titleWidth={TITLE_WIDTH}
                                    >
                                        <Select
                                            value={verifierVolume.value}
                                            onValueChange={verifierVolume.onChange}
                                        >
                                            <SelectTrigger
                                                className="w-full max-w-[360px]"
                                                aria-label="Verifier volume"
                                            >
                                                <SelectValue placeholder="Choose a cluster volume" />
                                            </SelectTrigger>
                                            <SelectContent>
                                                {volumes.map(volume => (
                                                    <SelectItem
                                                        key={volume.id}
                                                        value={volume.id}
                                                    >
                                                        {volume.name}
                                                    </SelectItem>
                                                ))}
                                                {volumes.length === 0 && (
                                                    <div className="px-2 py-1.5 text-xs text-muted-foreground">
                                                        No cluster volume pinned to a node in this project.
                                                    </div>
                                                )}
                                            </SelectContent>
                                        </Select>
                                        <FieldDescription>
                                            Pinned to the node the verifier runs on: preferably not the database&apos;s.
                                        </FieldDescription>
                                        <FieldError errors={[errors.verifierVolume]} />
                                    </InfoBlock>
                                </Field>
                                <Field>
                                    <InfoBlock
                                        title="Verifier key"
                                        titleWidth={TITLE_WIDTH}
                                    >
                                        <Select
                                            value={verifierKey.value}
                                            onValueChange={verifierKey.onChange}
                                        >
                                            <SelectTrigger
                                                className="w-full max-w-[360px]"
                                                aria-label="Verifier key"
                                            >
                                                <SelectValue />
                                            </SelectTrigger>
                                            <SelectContent>
                                                <SelectItem value={DATABASE_KEY}>The database&apos;s</SelectItem>
                                                {keyAuths.map(key => (
                                                    <SelectItem
                                                        key={key.id}
                                                        value={key.id}
                                                    >
                                                        {key.name}
                                                    </SelectItem>
                                                ))}
                                            </SelectContent>
                                        </Select>
                                        <FieldDescription>
                                            A key that only reads the archive keeps a verifier from changing it.
                                        </FieldDescription>
                                    </InfoBlock>
                                </Field>
                            </>
                        )}

                        {movingToContainer && (
                            <div
                                role="note"
                                className="flex items-start gap-2 rounded-md border border-amber-500/30 bg-amber-500/10 px-3 py-2 text-xs text-amber-800 dark:text-amber-300"
                            >
                                <AlertTriangle className="mt-0.5 size-3.5 shrink-0" />
                                <span>Saving deletes the verifier app: the restore test moves back in.</span>
                            </div>
                        )}
                    </FieldGroup>
                    <div className="mt-4 flex justify-end">
                        <Button
                            type="submit"
                            isLoading={isPending}
                            disabled={readOnly || !isDirty}
                            className="min-w-[100px]"
                        >
                            Save
                        </Button>
                    </div>
                </fieldset>
            </form>
        </section>
    );
}
