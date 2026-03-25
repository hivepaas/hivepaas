import { FieldError, Input } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";

import { InfoBlock } from "@application/shared/components";

import {
    type AppConfigDeploymentSettingsFormSchemaInput,
    type AppConfigDeploymentSettingsFormSchemaOutput,
} from "../schemas";

export function RunConfigurationFields() {
    const { control } = useFormContext<
        AppConfigDeploymentSettingsFormSchemaInput,
        unknown,
        AppConfigDeploymentSettingsFormSchemaOutput
    >();

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
            <InfoBlock title="Command">
                <Input
                    {...command}
                    value={command.value ?? ""}
                    onChange={command.onChange}
                    placeholder='my-app --arg1=123 --arg2="my data"'
                    className="max-w-[400px]"
                />
                <FieldError errors={[commandError]} />
            </InfoBlock>

            <InfoBlock title="Working Directory">
                <Input
                    {...workingDir}
                    value={workingDir.value ?? ""}
                    onChange={workingDir.onChange}
                    placeholder="/path/in/container"
                    className="max-w-[400px]"
                />
                <FieldError errors={[workingDirError]} />
            </InfoBlock>

            <InfoBlock title="Pre-deployment Command">
                <Input
                    {...preDeploymentCommand}
                    value={preDeploymentCommand.value ?? ""}
                    onChange={preDeploymentCommand.onChange}
                    className="max-w-[400px]"
                    placeholder="make prepare-deployment"
                />
                <FieldError errors={[preDeploymentCommandError]} />
            </InfoBlock>

            <InfoBlock title="Post-deployment Command">
                <Input
                    {...postDeploymentCommand}
                    value={postDeploymentCommand.value ?? ""}
                    onChange={postDeploymentCommand.onChange}
                    className="max-w-[400px]"
                    placeholder="make db-migrate-up"
                />
                <FieldError errors={[postDeploymentCommandError]} />
            </InfoBlock>
        </>
    );
}
