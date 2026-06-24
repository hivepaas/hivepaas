import { Checkbox, FieldError, Input } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import {
    type AppConfigContainerSettingsFormSchemaInput,
    type AppConfigContainerSettingsFormSchemaOutput,
} from "../schemas";

const DOCKER_INIT_TOOLTIP =
    "Runs an init process (Tini) as PID 1 to reap zombie processes and forward signals for a graceful shutdown.";
const ALLOCATE_TTY_TOOLTIP =
    "Allocates a pseudo-TTY (virtual terminal) to the container. Allows for terminal-like text styling (colors, formatting) and interactive shell sessions (e.g., bash/sh) when attaching to the container.";
const KEEP_STDIN_OPEN_TOOLTIP =
    "Keeps the container's standard input (stdin) open even if no client is attached. Essential for CLI applications or interactive shells that need to read keyboard inputs.";
const READONLY_ROOT_FILESYSTEM_TOOLTIP =
    "Mounts the container's root filesystem as read-only. Prevents the application from writing to the container's disk, enhancing security and ensuring statelessness. Any temp data must be written to volumes or tmpfs.";

function TooltipText({ text }: { text: string }) {
    return <span className="block max-w-[360px] whitespace-normal">{text}</span>;
}

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
    const { field: dockerInit } = useController({ control, name: "general.init" });
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
        <div className="flex flex-col gap-6">
            <InfoBlock title="Image">
                <Input
                    {...image}
                    value={image.value}
                    onChange={image.onChange}
                    readOnly
                    className={`${PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS} bg-muted/50 text-muted-foreground`}
                />
            </InfoBlock>

            <InfoBlock title="Command">
                <Input
                    {...command}
                    value={command.value}
                    onChange={command.onChange}
                    placeholder='my_app --arg1=123 --arg2="my data"'
                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                />
                <FieldError errors={[commandError]} />
            </InfoBlock>

            <InfoBlock title="Working Directory">
                <Input
                    {...workingDir}
                    value={workingDir.value}
                    onChange={workingDir.onChange}
                    placeholder="/path/in/container"
                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                />
                <FieldError errors={[workingDirError]} />
            </InfoBlock>

            <InfoBlock title="Hostname">
                <Input
                    {...hostname}
                    value={hostname.value}
                    onChange={hostname.onChange}
                    placeholder="hostname"
                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                />
                <FieldError errors={[hostnameError]} />
            </InfoBlock>

            <InfoBlock title="User">
                <Input
                    {...user}
                    value={user.value}
                    onChange={user.onChange}
                    placeholder="user"
                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                />
                <FieldError errors={[userError]} />
            </InfoBlock>

            <InfoBlock title="Groups">
                <Input
                    {...groups}
                    value={groups.value}
                    onChange={groups.onChange}
                    placeholder="space-separated groups"
                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                />
                <FieldError errors={[groupsError]} />
            </InfoBlock>

            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Docker Init"
                        content={<TooltipText text={DOCKER_INIT_TOOLTIP} />}
                    />
                }
            >
                <div className="flex items-center gap-2">
                    <Checkbox
                        checked={dockerInit.value}
                        onCheckedChange={v => {
                            dockerInit.onChange(v === true);
                        }}
                    />
                </div>
            </InfoBlock>

            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Allocate TTY"
                        content={<TooltipText text={ALLOCATE_TTY_TOOLTIP} />}
                    />
                }
            >
                <div className="flex items-center gap-2">
                    <Checkbox
                        checked={tty.value}
                        onCheckedChange={v => {
                            tty.onChange(v === true);
                        }}
                    />
                </div>
            </InfoBlock>

            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Keep Stdin Open"
                        content={<TooltipText text={KEEP_STDIN_OPEN_TOOLTIP} />}
                    />
                }
            >
                <div className="flex items-center gap-2">
                    <Checkbox
                        checked={openStdin.value}
                        onCheckedChange={v => {
                            openStdin.onChange(v === true);
                        }}
                    />
                </div>
            </InfoBlock>

            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Read-only Root Filesystem"
                        content={<TooltipText text={READONLY_ROOT_FILESYSTEM_TOOLTIP} />}
                    />
                }
            >
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
                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
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
    );
}
