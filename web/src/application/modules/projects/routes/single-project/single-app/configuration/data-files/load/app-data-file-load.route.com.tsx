import { useState } from "react";

import { format } from "date-fns";
import { AlertTriangle } from "lucide-react";
import { FormProvider, useForm } from "react-hook-form";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { ProjectAppsQueries } from "~/projects/data";
import { AppDataFilesCommands } from "~/projects/data/commands";
import { AppDataFilesQueries } from "~/projects/data/queries";
import type { AppDataFile } from "~/projects/domain";
import {
    CommandConfigSection,
    type CommandFormInput,
    emptySourceCommand,
    hasCommand,
    mapSourceCommandToPayload,
} from "~/projects/module-shared/components";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";

import {
    AppLoader,
    ContentBlock,
    FormActionBar,
    InfoBlock,
    LabelWithInfo,
    RouteFormHeader,
} from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { formatDataSizeCompact } from "@application/shared/utils/data-size";

import { Button, Input } from "@/components/ui";

const TITLE_WIDTH = 220;
const CONTROL_CLASS = PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS;
const LOAD_COMMAND_EXAMPLE = "psql -U $POSTGRES_USER $POSTGRES_DB";

/** A file its job saved encrypted, as the job names it. */
function isEncryptedDataFile(name: string): boolean {
    return name.endsWith(".age");
}

/** Loading a data file into a command run in the app, as a page of its own: the command, and a confirmation. */
export function AppDataFileLoadRoute() {
    const {
        id: projectId,
        env,
        appId,
        dataFileId,
    } = useParams<{
        id: string;
        env: string;
        appId: string;
        dataFileId: string;
    }>();
    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");
    invariant(appId, "appId must be defined");
    invariant(dataFileId, "dataFileId must be defined");

    const { data: fileData, isLoading: isLoadingFile } = AppDataFilesQueries.useFindOneById({
        projectID: projectId,
        env,
        appID: appId,
        dataFileID: dataFileId,
    });
    const { data: appData, isLoading: isLoadingApp } = ProjectAppsQueries.useFindOneById({
        projectID: projectId,
        env,
        appID: appId,
    });
    const file = fileData?.data;
    const appName = appData?.data.name ?? "";

    return (
        <div className="flex w-full flex-col">
            <RouteFormHeader title={file ? `Load ${file.name}` : "Load Data File"} />
            {(isLoadingFile || isLoadingApp) && (
                <div className="flex min-h-[220px] items-center justify-center">
                    <AppLoader />
                </div>
            )}
            {!isLoadingFile && !isLoadingApp && (!file || !appName) && (
                <p className="px-4 text-sm text-muted-foreground">The data file, or its app, is not found.</p>
            )}
            {file && appName && (
                <LoadForm
                    key={file.id}
                    projectId={projectId}
                    env={env}
                    appId={appId}
                    appName={appName}
                    file={file}
                />
            )}
        </div>
    );
}

function LoadForm({ projectId, env, appId, appName, file }: LoadFormProps) {
    const { navigate } = useAppNavigate();
    const [passphrase, setPassphrase] = useState("");
    const [typed, setTyped] = useState("");
    const form = useForm<{ command: CommandFormInput }>({ defaultValues: { command: emptySourceCommand() } });
    const encrypted = isEncryptedDataFile(file.name);

    function backToList() {
        navigate.modules(ROUTE.projects.single.apps.single.configuration.dataFiles.$route(projectId, env, appId), {
            ignorePrevPath: true,
        });
    }

    const { mutate: load, isPending } = AppDataFilesCommands.useLoad({
        onSuccess: response => {
            toast.success("Load started");
            navigate.modules(
                ROUTE.projects.single.apps.single.tasks.details.$route(projectId, env, appId, response.data.taskId),
                { ignorePrevPath: true },
            );
        },
    });

    const problems: string[] = [];
    if (!hasCommand(form.watch("command"))) {
        problems.push("Give the command that reads the file.");
    }
    if (encrypted && !passphrase) {
        problems.push("Give the passphrase the file was saved with.");
    }
    const confirmed = typed.trim() === appName;

    function submit() {
        load({
            projectID: projectId,
            env,
            appID: appId,
            dataFileID: file.id,
            payload: {
                command: mapSourceCommandToPayload(form.getValues().command),
                ...(encrypted ? { passphrase } : {}),
            },
        });
    }

    return (
        <fieldset
            disabled={isPending}
            className="flex flex-col gap-6"
        >
            <ContentBlock label="Data File">
                <div className="flex flex-col gap-4 text-sm">
                    <InfoBlock
                        title="Name"
                        titleWidth={TITLE_WIDTH}
                    >
                        <span className="font-mono">{file.name}</span>
                    </InfoBlock>
                    <InfoBlock
                        title="Size"
                        titleWidth={TITLE_WIDTH}
                    >
                        {formatDataSizeCompact(file.sizeBytes)}
                    </InfoBlock>
                    <InfoBlock
                        title="Created"
                        titleWidth={TITLE_WIDTH}
                    >
                        {format(file.createdAt, "yyyy-MM-dd HH:mm:ss")}
                    </InfoBlock>
                </div>
            </ContentBlock>
            <FormProvider {...form}>
                <p className="px-4 text-sm text-muted-foreground">
                    The command runs in a container of the app, and reads the file on its stdin, without a TTY -
                    decompressed and decrypted as its name says. For example:{" "}
                    <span className="font-mono">{LOAD_COMMAND_EXAMPLE}</span>
                </p>
                <CommandConfigSection
                    label="Command"
                    fieldPrefix="command"
                    showLoadTemplate
                    templateProjectId={projectId}
                    templateEnv={env}
                    readOnly={isPending}
                    showArgGroups
                    hideTerminal
                />
            </FormProvider>
            {encrypted && (
                <ContentBlock label="Decryption">
                    <InfoBlock
                        title={
                            <LabelWithInfo
                                label="Passphrase"
                                isRequired
                                content="The file was saved encrypted: the passphrase its job saved it with reads it."
                            />
                        }
                        titleWidth={TITLE_WIDTH}
                    >
                        <Input
                            type="password"
                            value={passphrase}
                            onChange={event => {
                                setPassphrase(event.target.value);
                            }}
                            autoComplete="off"
                            className={CONTROL_CLASS}
                        />
                    </InfoBlock>
                </ContentBlock>
            )}
            <ContentBlock label="Confirm">
                <InfoBlock
                    title="App Name"
                    titleWidth={TITLE_WIDTH}
                >
                    <div className={`flex flex-col gap-2 ${CONTROL_CLASS}`}>
                        <div className="flex items-center gap-2.5 rounded-md border border-destructive bg-destructive/10 px-3.5 py-2.5 text-sm text-destructive font-medium">
                            <AlertTriangle className="size-4 shrink-0" />
                            <span>
                                The command can change {appName}&apos;s data. Type the app&apos;s name to confirm.
                            </span>
                        </div>
                        <Input
                            value={typed}
                            onChange={event => {
                                setTyped(event.target.value);
                            }}
                            placeholder={appName}
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
                    disabled={problems.length > 0 || !confirmed}
                    onClick={submit}
                >
                    Load
                </Button>
            </FormActionBar>
        </fieldset>
    );
}

interface LoadFormProps {
    projectId: string;
    env: string;
    appId: string;
    appName: string;
    file: AppDataFile;
}
