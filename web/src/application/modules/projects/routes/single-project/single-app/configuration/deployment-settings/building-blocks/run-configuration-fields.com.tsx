import { FieldError, Input } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import {
    type AppConfigDeploymentSettingsFormSchemaInput,
    type AppConfigDeploymentSettingsFormSchemaOutput,
} from "../schemas";

const ENTRYPOINT_TOOLTIP =
    "Runs instead of the image's entrypoint, with the command as its arguments. Leave it empty to keep the image's.";
const COMMAND_TOOLTIP =
    "The arguments of the entrypoint - the image's, or the one above. Leave it empty to keep the image's command.";

function TooltipText({ text }: { text: string }) {
    return <span className="block max-w-[360px] whitespace-normal">{text}</span>;
}

export function RunConfigurationFields() {
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
