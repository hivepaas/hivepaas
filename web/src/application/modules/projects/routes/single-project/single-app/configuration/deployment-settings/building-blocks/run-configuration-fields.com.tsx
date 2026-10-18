import {
    FieldDescription,
    FieldError,
    Input,
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@components/ui";
import { useController, useFormContext } from "react-hook-form";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";
import type { ManagedDatabaseOption } from "~/projects/module-shared/hooks";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import {
    type AppConfigDeploymentSettingsFormSchemaInput,
    type AppConfigDeploymentSettingsFormSchemaOutput,
} from "../schemas";

const ENTRYPOINT_TOOLTIP =
    "Runs instead of the image's entrypoint, with the command as its arguments. Leave it empty to keep the image's.";
const COMMAND_TOOLTIP =
    "The arguments of the entrypoint - the image's, or the one above. Leave it empty to keep the image's command.";
const RESTORE_POINT_TOOLTIP =
    "A managed database of this environment that each deployment marks a restore point on, before its " +
    "pre-deployment command: a migration that breaks data then has its way back already marked.";
// The select's value for no database: an empty value is the placeholder's.
const NO_DATABASE = "__none__";

function selectValueOf(database: string | undefined): string {
    return database === undefined || database === "" ? NO_DATABASE : database;
}

function TooltipText({ text }: { text: string }) {
    return <span className="block max-w-[360px] whitespace-normal">{text}</span>;
}

interface Props {
    /** A managed database's entrypoint is its engine's. */
    entrypointLocked?: boolean;
    /** The environment's managed databases a deployment can mark a restore point on. */
    managedDatabases?: ManagedDatabaseOption[];
}

export function RunConfigurationFields({ entrypointLocked = false, managedDatabases = [] }: Props) {
    const { control } = useFormContext<
        AppConfigDeploymentSettingsFormSchemaInput,
        unknown,
        AppConfigDeploymentSettingsFormSchemaOutput
    >();

    const {
        field: entrypoint,
        fieldState: { error: entrypointError },
    } = useController({ control, name: "entrypoint" });

    const {
        field: command,
        fieldState: { error: commandError },
    } = useController({ control, name: "command" });

    const {
        field: workingDir,
        fieldState: { error: workingDirError },
    } = useController({ control, name: "workingDir" });

    const {
        field: preDeploymentCommand,
        fieldState: { error: preDeploymentCommandError },
    } = useController({ control, name: "preDeploymentCommand" });

    const {
        field: postDeploymentCommand,
        fieldState: { error: postDeploymentCommandError },
    } = useController({ control, name: "postDeploymentCommand" });

    const {
        field: restorePoint,
        fieldState: { error: restorePointError },
    } = useController({ control, name: "restorePointBeforeDeploy" });
    // One saved that the list does not have - deleted, or no longer managed - is
    // shown as it is, so that saving does not drop it unseen.
    const savedElsewhere = restorePoint.value && !managedDatabases.some(database => database.id === restorePoint.value);

    return (
        <>
            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Entrypoint"
                        content={<TooltipText text={ENTRYPOINT_TOOLTIP} />}
                    />
                }
            >
                <Input
                    {...entrypoint}
                    value={entrypoint.value ?? ""}
                    onChange={entrypoint.onChange}
                    placeholder="/docker-entrypoint.sh"
                    disabled={entrypointLocked}
                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                />
                <FieldError errors={[entrypointError]} />
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Command"
                        content={<TooltipText text={COMMAND_TOOLTIP} />}
                    />
                }
            >
                <Input
                    {...command}
                    value={command.value ?? ""}
                    onChange={command.onChange}
                    placeholder='my-app --arg1=123 --arg2="my data"'
                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                />
                <FieldError errors={[commandError]} />
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title="Working Directory"
            >
                <Input
                    {...workingDir}
                    value={workingDir.value ?? ""}
                    onChange={workingDir.onChange}
                    placeholder="/path/in/container"
                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                />
                <FieldError errors={[workingDirError]} />
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title="Pre-deployment Command"
            >
                <Input
                    {...preDeploymentCommand}
                    value={preDeploymentCommand.value ?? ""}
                    onChange={preDeploymentCommand.onChange}
                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                    placeholder="make prepare-deployment"
                />
                <FieldError errors={[preDeploymentCommandError]} />
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Restore point before deploy"
                        content={<TooltipText text={RESTORE_POINT_TOOLTIP} />}
                    />
                }
            >
                <Select
                    value={selectValueOf(restorePoint.value)}
                    onValueChange={value => {
                        restorePoint.onChange(value === NO_DATABASE ? "" : value);
                    }}
                >
                    <SelectTrigger
                        className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                        aria-label="Restore point before deploy"
                    >
                        <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value={NO_DATABASE}>None</SelectItem>
                        {managedDatabases.map(database => (
                            <SelectItem
                                key={database.id}
                                value={database.id}
                            >
                                {database.name}
                            </SelectItem>
                        ))}
                        {savedElsewhere && (
                            <SelectItem value={restorePoint.value ?? ""}>
                                {restorePoint.value} (not a managed database of this environment)
                            </SelectItem>
                        )}
                    </SelectContent>
                </Select>
                <FieldDescription>
                    A point that cannot be made stops the deployment before anything changes. The points deployments
                    make are removed once no base backup reaches them.
                </FieldDescription>
                <FieldError errors={[restorePointError]} />
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title="Post-deployment Command"
            >
                <Input
                    {...postDeploymentCommand}
                    value={postDeploymentCommand.value ?? ""}
                    onChange={postDeploymentCommand.onChange}
                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                    placeholder="make db-migrate-up"
                />
                <FieldError errors={[postDeploymentCommandError]} />
            </InfoBlock>
        </>
    );
}
