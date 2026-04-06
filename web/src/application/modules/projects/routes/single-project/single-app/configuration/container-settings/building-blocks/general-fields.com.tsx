import { Checkbox, FieldError, Input } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";

import { InfoBlock } from "@application/shared/components";

import {
    type AppConfigContainerSettingsFormSchemaInput,
    type AppConfigContainerSettingsFormSchemaOutput,
} from "../schemas";

export function GeneralFields() {
    const { control } = useFormContext<
        AppConfigContainerSettingsFormSchemaInput,
        unknown,
        AppConfigContainerSettingsFormSchemaOutput
    >();

    const { field: image } = useController({ control, name: "general.image" });
    const {
        field: command,
        fieldState: { error: commandError },
    } = useController({
        control,
        name: "general.command",
    });
    const {
        field: workingDir,
        fieldState: { error: workingDirError },
    } = useController({
        control,
        name: "general.workingDir",
    });
    const {
        field: hostname,
        fieldState: { error: hostnameError },
    } = useController({
        control,
        name: "general.hostname",
    });
    const {
        field: user,
        fieldState: { error: userError },
    } = useController({ control, name: "general.user" });
    const {
        field: groups,
        fieldState: { error: groupsError },
    } = useController({
        control,
        name: "general.groups",
    });
    const { field: tty } = useController({ control, name: "general.tty" });
    const { field: openStdin } = useController({ control, name: "general.openStdin" });
    const { field: readOnly } = useController({ control, name: "general.readOnly" });
    const {
        field: stopSignal,
        fieldState: { error: stopSignalError },
    } = useController({
        control,
        name: "general.stopSignal",
    });
    const {
        field: stopGracePeriod,
        fieldState: { error: stopGracePeriodError },
    } = useController({
        control,
        name: "general.stopGracePeriod",
    });

    return (
        <>
            <h3 className="font-medium bg-accent py-2 px-3 rounded-lg">General</h3>
            <div className="flex flex-col gap-6 px-2">
                <InfoBlock title="Image">
                    <Input
                        {...image}
                        value={image.value}
                        onChange={image.onChange}
                        readOnly
                        className="max-w-[400px] bg-muted/50 text-muted-foreground"
                    />
                </InfoBlock>

                <InfoBlock title="Command">
                    <Input
                        {...command}
                        value={command.value}
                        onChange={command.onChange}
                        placeholder='my_app --arg1=123 --arg2="my data"'
                        className="max-w-[400px]"
                    />
                    <FieldError errors={[commandError]} />
                </InfoBlock>

                <InfoBlock title="Working Directory">
                    <Input
                        {...workingDir}
                        value={workingDir.value}
                        onChange={workingDir.onChange}
                        placeholder="/path/in/container"
                        className="max-w-[400px]"
                    />
                    <FieldError errors={[workingDirError]} />
                </InfoBlock>

                <InfoBlock title="Hostname">
                    <Input
                        {...hostname}
                        value={hostname.value}
                        onChange={hostname.onChange}
                        placeholder="hostname"
                        className="max-w-[400px]"
                    />
                    <FieldError errors={[hostnameError]} />
                </InfoBlock>

                <InfoBlock title="User">
                    <Input
                        {...user}
                        value={user.value}
                        onChange={user.onChange}
                        placeholder="user"
                        className="max-w-[400px]"
                    />
                    <FieldError errors={[userError]} />
                </InfoBlock>

                <InfoBlock title="Groups">
                    <Input
                        {...groups}
                        value={groups.value}
                        onChange={groups.onChange}
                        placeholder="space-separated groups"
                        className="max-w-[400px]"
                    />
                    <FieldError errors={[groupsError]} />
                </InfoBlock>

                <InfoBlock title="TTY">
                    <div className="flex items-center gap-2">
                        <Checkbox
                            checked={tty.value}
                            onCheckedChange={v => {
                                tty.onChange(v === true);
                            }}
                        />
                    </div>
                </InfoBlock>

                <InfoBlock title="Open Stdin">
                    <div className="flex items-center gap-2">
                        <Checkbox
                            checked={openStdin.value}
                            onCheckedChange={v => {
                                openStdin.onChange(v === true);
                            }}
                        />
                    </div>
                </InfoBlock>

                <InfoBlock title="Read-only root filesystem">
                    <div className="flex items-center gap-2">
                        <Checkbox
                            checked={readOnly.value}
                            onCheckedChange={v => {
                                readOnly.onChange(v === true);
                            }}
                        />
                    </div>
                </InfoBlock>

                <InfoBlock title="Stop Signal">
                    <Input
                        {...stopSignal}
                        value={stopSignal.value}
                        onChange={stopSignal.onChange}
                        placeholder="SIGTERM"
                        className="max-w-[400px]"
                    />
                    <FieldError errors={[stopSignalError]} />
                </InfoBlock>

                <InfoBlock title="Stop Grace Period">
                    <Input
                        {...stopGracePeriod}
                        value={stopGracePeriod.value}
                        onChange={stopGracePeriod.onChange}
                        placeholder="5s"
                        className="max-w-[400px]"
                    />
                    <FieldError errors={[stopGracePeriodError]} />
                </InfoBlock>
            </div>
        </>
    );
}
