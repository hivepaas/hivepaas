import { useMemo, useState } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { FormProvider, useController, useForm, useFormState } from "react-hook-form";
import { useUpdateEffect } from "react-use";
import { APP_CONFIGURATION_QUERY_OPTIONS, AppStorageSettingsQueries } from "~/projects/data";
import { ProjectBackupRepoQueries } from "~/projects/data/queries";
import type { AppScheduledJob } from "~/projects/domain";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";
import { ESchedJobDataBackupSource } from "~/projects/module-shared/enums";
import { useProjectNotificationSettingsSources } from "~/projects/module-shared/hooks";

import {
    AppLink,
    Combobox,
    ContentBlock,
    FormActionBar,
    InfoBlock,
    LabelWithInfo,
} from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { KeyValueList, NotificationSettings } from "@application/shared/form";

import { Button, Checkbox, Field, FieldError, FieldGroup, Input, Tabs, TabsList, TabsTrigger } from "@/components/ui";

import { CommandConfigSection } from "../command-config-section";
import { JobScheduleFields, PriorityTabsField } from "../job-schedule-fields";
import { JobTriggersField } from "../job-triggers-field";

import { createEmptyDataBackupFormDefaults, mapDataBackupToFormInput } from "./data-backup.form-mappers";
import { type DataBackupFormInput, type DataBackupFormOutput, DataBackupFormSchema } from "./data-backup.form.schema";

const TITLE_WIDTH = 220;

interface RefOption {
    [key: string]: unknown;
    id: string;
    name: string;
}

/** Reading a volume while the app writes to it. */
function VolumeConsistencyNote() {
    return (
        <div className={cn(dashedBorderBox, "flex max-w-[720px] flex-col gap-2 text-sm leading-normal")}>
            <p>
                <span className="font-semibold text-orange-500">Note:</span> a volume the app writes to while it is read
                may be caught mid-write. For a database, back up its dump command instead; otherwise a job sequence
                whose first step stops or flushes the app, then this backup, avoids it.
            </p>
        </div>
    );
}

/**
 * A data backup of an app: a command's output, or a volume the app mounts, into
 * a backup repository the app sees.
 */
export function DataBackupForm({
    projectId,
    env,
    appId,
    isPending,
    onSubmit,
    initialValues,
    onHasChanges,
    readOnly = false,
    stickyActions = false,
    onClose,
}: Props) {
    const defaultValues = useMemo(
        () => (initialValues ? mapDataBackupToFormInput(initialValues) : createEmptyDataBackupFormDefaults()),
        [initialValues],
    );

    const methods = useForm<DataBackupFormInput, unknown, DataBackupFormOutput>({
        defaultValues,
        resolver: zodResolver(DataBackupFormSchema),
        mode: "onSubmit",
    });
    const {
        control,
        handleSubmit,
        formState: { errors },
    } = methods;
    const { isDirty } = useFormState({ control });

    useUpdateEffect(() => {
        methods.reset(defaultValues);
    }, [defaultValues]);

    useUpdateEffect(() => {
        onHasChanges?.(readOnly ? false : isDirty);
    }, [isDirty, readOnly]);

    const { sources: notificationSources, manageLink: notificationManageLink } = useProjectNotificationSettingsSources(
        projectId,
        env,
    );

    const { field: name, fieldState: nameState } = useController({ control, name: "name" });
    const { field: source } = useController({ control, name: "source" });
    const { field: fileName, fieldState: fileNameState } = useController({ control, name: "sourceFileName" });
    const { field: volume, fieldState: volumeState } = useController({ control, name: "sourceVolume" });
    const { field: subpath, fieldState: subpathState } = useController({ control, name: "sourceVolumeSubpath" });
    const { field: repository, fieldState: repositoryState } = useController({ control, name: "targetRepository" });
    const { field: timeout, fieldState: timeoutState } = useController({ control, name: "timeout" });
    const { field: maxRetry, fieldState: maxRetryState } = useController({ control, name: "maxRetry" });
    const { field: retryDelay, fieldState: retryDelayState } = useController({ control, name: "retryDelay" });
    const { field: priority } = useController({ control, name: "priority" });
    const { field: controlEnabled } = useController({ control, name: "controlEnabled" });

    const isCommand = source.value === ESchedJobDataBackupSource.Command;

    // The volumes the app mounts as its own directory: what a volume source reads.
    const { data: storageData, isFetching: isStorageFetching } = AppStorageSettingsQueries.useFindOne(
        { projectID: projectId, env, appID: appId },
        { ...APP_CONFIGURATION_QUERY_OPTIONS, enabled: !isCommand },
    );
    const volumeOptions = useMemo(() => {
        const seen = new Set<string>();
        return (storageData?.data.mounts ?? []).flatMap(mount => {
            if (!mount.volumeId || seen.has(mount.volumeId)) {
                return [];
            }
            seen.add(mount.volumeId);
            const label = mount.target ?? mount.volumeId;
            return [{ value: { id: mount.volumeId, name: label } satisfies RefOption, label }];
        });
    }, [storageData]);

    const [repositorySearch, setRepositorySearch] = useState("");
    const {
        data: repositoryData,
        isFetching: isRepositoryFetching,
        isRefetching: isRepositoryRefetching,
        refetch: refetchRepositories,
    } = ProjectBackupRepoQueries.useFindManyPaginated({ projectID: projectId, env, search: repositorySearch });
    const repositoryOptions = (repositoryData?.data ?? []).map(repo => ({
        value: { id: repo.id, name: repo.name } satisfies RefOption,
        label: repo.name,
    }));

    const configureTemplatesLink = ROUTE.projects.single.providerConfiguration.commandTemplates.$route(projectId);
    const backupReposLink = ROUTE.projects.single.providerConfiguration.backupRepos.$route(projectId);

    return (
        <FormProvider {...methods}>
            <form
                onSubmit={event => {
                    event.preventDefault();
                    if (readOnly) {
                        return;
                    }

                    void handleSubmit(values => {
                        onSubmit(values);
                    })(event);
                }}
                className="min-h-0 flex flex-1 flex-col"
            >
                <fieldset className="contents">
                    <FieldGroup className="gap-6">
                        <InfoBlock
                            titleWidth={TITLE_WIDTH}
                            title={
                                <LabelWithInfo
                                    label="Name"
                                    isRequired
                                />
                            }
                        >
                            <Field>
                                <Input
                                    {...name}
                                    placeholder="data backup name"
                                    aria-invalid={nameState.invalid}
                                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                    disabled={readOnly}
                                />
                                <FieldError errors={[errors.name]} />
                            </Field>
                        </InfoBlock>

                        <ContentBlock label="Source">
                            <div className="flex flex-col gap-6">
                                <InfoBlock
                                    title="Back Up"
                                    titleWidth={TITLE_WIDTH}
                                >
                                    <div className="flex flex-col gap-2">
                                        <Tabs
                                            value={source.value}
                                            onValueChange={source.onChange}
                                        >
                                            <TabsList>
                                                <TabsTrigger
                                                    value={ESchedJobDataBackupSource.Command}
                                                    disabled={readOnly}
                                                >
                                                    Command
                                                </TabsTrigger>
                                                <TabsTrigger
                                                    value={ESchedJobDataBackupSource.Volume}
                                                    disabled={readOnly}
                                                >
                                                    Volume
                                                </TabsTrigger>
                                            </TabsList>
                                        </Tabs>
                                        <p className="text-sm text-muted-foreground">
                                            {isCommand
                                                ? "A command runs in the app, without a TTY; what it prints is the backup, such as pg_dump's output."
                                                : "A directory of a volume the app mounts, read on the volume's node."}
                                        </p>
                                    </div>
                                </InfoBlock>

                                {isCommand ? (
                                    <InfoBlock
                                        title={
                                            <LabelWithInfo
                                                label="File Name"
                                                isRequired
                                                content="The name of the command's output in the snapshot."
                                            />
                                        }
                                        titleWidth={TITLE_WIDTH}
                                    >
                                        <Field>
                                            <Input
                                                {...fileName}
                                                placeholder="db.sql"
                                                className="max-w-[400px]"
                                                aria-invalid={fileNameState.invalid}
                                                disabled={readOnly}
                                            />
                                            <FieldError errors={[errors.sourceFileName]} />
                                        </Field>
                                    </InfoBlock>
                                ) : (
                                    <>
                                        <InfoBlock
                                            title={
                                                <LabelWithInfo
                                                    label="Volume"
                                                    isRequired
                                                    content="The app's own directory of a volume, where the app mounts it."
                                                />
                                            }
                                            titleWidth={TITLE_WIDTH}
                                        >
                                            <Field>
                                                <Combobox<RefOption>
                                                    options={volumeOptions}
                                                    value={(volume.value as RefOption | null)?.id ?? null}
                                                    onChange={(_, option) => {
                                                        volume.onChange(option ?? null);
                                                    }}
                                                    placeholder={volume.value?.name ?? "Select a mounted volume"}
                                                    emptyText="The app mounts no volume of its own"
                                                    className="max-w-[400px]"
                                                    valueKey="id"
                                                    closeOnSelect
                                                    loading={isStorageFetching}
                                                    aria-invalid={volumeState.invalid}
                                                    disabled={readOnly}
                                                />
                                                <FieldError errors={[errors.sourceVolume]} />
                                            </Field>
                                        </InfoBlock>

                                        <InfoBlock
                                            title={
                                                <LabelWithInfo
                                                    label="Path"
                                                    content="A path inside the volume, as the app sees it; empty for all of it."
                                                />
                                            }
                                            titleWidth={TITLE_WIDTH}
                                        >
                                            <Field>
                                                <Input
                                                    {...subpath}
                                                    placeholder="uploads"
                                                    className="max-w-[400px]"
                                                    aria-invalid={subpathState.invalid}
                                                    disabled={readOnly}
                                                />
                                                <FieldError errors={[errors.sourceVolumeSubpath]} />
                                            </Field>
                                        </InfoBlock>

                                        <VolumeConsistencyNote />
                                    </>
                                )}
                            </div>
                        </ContentBlock>

                        {isCommand && (
                            <CommandConfigSection
                                label="Command"
                                fieldPrefix="sourceCommand"
                                showLoadTemplate
                                templateProjectId={projectId}
                                templateEnv={env}
                                configureTemplatesLink={configureTemplatesLink}
                                readOnly={readOnly}
                                showArgGroups
                            />
                        )}

                        {isCommand && (
                            <CommandConfigSection
                                label={
                                    <LabelWithInfo
                                        label="Restore Command"
                                        content="Optional. Loads a backup it reads on its stdin, such as psql -U $POSTGRES_USER $POSTGRES_DB. A restore of this job's snapshots offers it, and runs it in the app without a TTY."
                                    />
                                }
                                fieldPrefix="restoreCommand"
                                showLoadTemplate
                                templateProjectId={projectId}
                                templateEnv={env}
                                readOnly={readOnly}
                                showArgGroups
                            />
                        )}

                        <ContentBlock label="Repository">
                            <div className="flex flex-col gap-6">
                                <InfoBlock
                                    title={
                                        <LabelWithInfo
                                            label="Backup Repository"
                                            isRequired
                                        />
                                    }
                                    titleWidth={TITLE_WIDTH}
                                >
                                    <Field>
                                        <Combobox<RefOption>
                                            options={repositoryOptions}
                                            value={(repository.value as RefOption | null)?.id ?? null}
                                            onChange={(_, option) => {
                                                repository.onChange(option ?? null);
                                            }}
                                            onSearch={setRepositorySearch}
                                            placeholder={repository.value?.name ?? "Select a backup repository"}
                                            emptyText="No backup repositories available"
                                            className="max-w-[400px]"
                                            valueKey="id"
                                            searchable
                                            closeOnSelect
                                            loading={isRepositoryFetching}
                                            onRefresh={() => void refetchRepositories()}
                                            isRefreshing={isRepositoryRefetching}
                                            aria-invalid={repositoryState.invalid}
                                            disabled={readOnly}
                                        />
                                        <AppLink.Modules
                                            to={backupReposLink}
                                            className="text-xs text-link hover:underline"
                                        >
                                            Configure Backup Repositories
                                        </AppLink.Modules>
                                        <FieldError errors={[errors.targetRepository]} />
                                    </Field>
                                </InfoBlock>

                                <InfoBlock
                                    title={
                                        <LabelWithInfo
                                            label="Tags"
                                            content="Put on every snapshot the job takes, beside the job's and the app's."
                                        />
                                    }
                                    titleWidth={TITLE_WIDTH}
                                >
                                    <FieldGroup>
                                        <KeyValueList
                                            name={"tags" as never}
                                            keyLabel="Key"
                                            valueLabel="Value"
                                            keyPlaceholder="env"
                                            valuePlaceholder="prod"
                                            className="max-w-[600px]"
                                            checkDuplicates
                                            disabled={readOnly}
                                            enableValueEditing
                                        />
                                        <FieldError errors={[errors.tags?.root ?? errors.tags]} />
                                    </FieldGroup>
                                </InfoBlock>
                            </div>
                        </ContentBlock>

                        <ContentBlock label="Scheduling">
                            <div className="flex flex-col gap-6">
                                <InfoBlock
                                    title="Priority"
                                    titleWidth={TITLE_WIDTH}
                                >
                                    <PriorityTabsField
                                        value={priority.value}
                                        onChange={priority.onChange}
                                        readOnly={readOnly}
                                    />
                                </InfoBlock>

                                <JobScheduleFields
                                    titleWidth={TITLE_WIDTH}
                                    nextRuns={initialValues?.nextRuns ?? []}
                                    readOnly={readOnly}
                                />

                                <InfoBlock
                                    title="Timeout"
                                    titleWidth={TITLE_WIDTH}
                                >
                                    <Field>
                                        <Input
                                            {...timeout}
                                            placeholder="30m, 1h30m"
                                            className="max-w-[400px]"
                                            aria-invalid={timeoutState.invalid}
                                            disabled={readOnly}
                                        />
                                        <FieldError errors={[errors.timeout]} />
                                    </Field>
                                </InfoBlock>

                                <InfoBlock
                                    title="Retry"
                                    titleWidth={TITLE_WIDTH}
                                >
                                    <div className="flex flex-wrap items-start gap-4">
                                        <Field className="w-auto">
                                            <Input
                                                ref={maxRetry.ref}
                                                name={maxRetry.name}
                                                type="number"
                                                inputMode="numeric"
                                                min={0}
                                                value={maxRetry.value ?? ""}
                                                onBlur={maxRetry.onBlur}
                                                onChange={event => {
                                                    const next = Number(event.target.value);
                                                    maxRetry.onChange(
                                                        event.target.value === "" || !Number.isFinite(next)
                                                            ? undefined
                                                            : next,
                                                    );
                                                }}
                                                placeholder="max, 0"
                                                className="w-24"
                                                aria-invalid={maxRetryState.invalid}
                                                disabled={readOnly}
                                            />
                                            <FieldError errors={[errors.maxRetry]} />
                                        </Field>
                                        <Field className="w-auto">
                                            <Input
                                                {...retryDelay}
                                                placeholder="delay, 1m"
                                                className="w-24"
                                                aria-invalid={retryDelayState.invalid}
                                                disabled={readOnly}
                                            />
                                            <FieldError errors={[errors.retryDelay]} />
                                        </Field>
                                    </div>
                                </InfoBlock>

                                <InfoBlock
                                    title="Allow Canceling"
                                    titleWidth={TITLE_WIDTH}
                                >
                                    <Checkbox
                                        checked={controlEnabled.value}
                                        onCheckedChange={checked => {
                                            controlEnabled.onChange(checked === true);
                                        }}
                                        disabled={readOnly}
                                    />
                                </InfoBlock>
                            </div>
                        </ContentBlock>

                        <ContentBlock label="Triggers">
                            <JobTriggersField readOnly={readOnly} />
                        </ContentBlock>

                        <ContentBlock label="Notification Configuration">
                            <NotificationSettings<DataBackupFormInput>
                                names={{
                                    successUseDefault: "notification.successUseDefault",
                                    success: "notification.success",
                                    failureUseDefault: "notification.failureUseDefault",
                                    failure: "notification.failure",
                                }}
                                sources={notificationSources}
                                manageLink={notificationManageLink}
                                readOnly={readOnly}
                                titleWidth={TITLE_WIDTH}
                            />
                        </ContentBlock>
                    </FieldGroup>

                    <FormActionBar sticky={stickyActions}>
                        {readOnly ? (
                            <Button
                                type="button"
                                onClick={onClose}
                                className="min-w-[100px]"
                            >
                                Close
                            </Button>
                        ) : (
                            <>
                                <Button
                                    type="button"
                                    variant="outline"
                                    className="min-w-[100px]"
                                    disabled={isPending}
                                    onClick={onClose}
                                >
                                    Cancel
                                </Button>
                                <Button
                                    type="submit"
                                    isLoading={isPending}
                                    className="min-w-[100px]"
                                >
                                    Save
                                </Button>
                            </>
                        )}
                    </FormActionBar>
                </fieldset>
            </form>
        </FormProvider>
    );
}

interface Props {
    projectId: string;
    env: string;
    appId: string;
    isPending: boolean;
    onSubmit: (values: DataBackupFormOutput) => void;
    initialValues?: AppScheduledJob;
    onHasChanges?: (dirty: boolean) => void;
    readOnly?: boolean;
    stickyActions?: boolean;
    onClose?: () => void;
}
