import { Checkbox } from "@components/ui";
import { cn } from "@lib/utils";
import { Container, Lock, ServerCog } from "lucide-react";
import { useController, useFormContext } from "react-hook-form";
import { type AppDockerApiHostModeBlocker, type AppDockerApiMode } from "~/projects/domain";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { type AppConfigDockerApiFormSchemaInput, type AppConfigDockerApiFormSchemaOutput } from "../schemas";

const TITLE_WIDTH = 220;

/** Why host mode cannot be chosen, in the words of the person who can change it. */
const HOST_MODE_BLOCKED: Record<Exclude<AppDockerApiHostModeBlocker, "">, string> = {
    switch: "The Privileged Apps switch is off. An administrator turns it on in System → HivePaaS → Security.",
    admin: "Only an administrator can give an app the node's Docker socket.",
};

export function DockerApiEnabledField() {
    const { control } = useFormContext<
        AppConfigDockerApiFormSchemaInput,
        unknown,
        AppConfigDockerApiFormSchemaOutput
    >();
    const { field } = useController({ control, name: "enabled" });

    return (
        <InfoBlock
            titleWidth={TITLE_WIDTH}
            title={
                <LabelWithInfo
                    label="Enabled"
                    content="Turned off, the app loses the Docker API and its containers are removed. What access allowed is kept, for turning it on again."
                />
            }
        >
            <Checkbox
                checked={field.value}
                onCheckedChange={value => {
                    field.onChange(value === true);
                }}
            />
        </InfoBlock>
    );
}

/**
 * The choice between the proxy HivePaaS runs and the node's own socket. Host mode
 * is always shown, so that nobody wonders whether it exists, and locked with the
 * reason for someone who may not choose it - unless the app already has it,
 * which anybody may leave.
 */
export function DockerApiModeField({ hostModeBlockedBy, hasHostMode }: ModeProps) {
    const { control } = useFormContext<
        AppConfigDockerApiFormSchemaInput,
        unknown,
        AppConfigDockerApiFormSchemaOutput
    >();
    const { field } = useController({ control, name: "mode" });
    const hostLocked = hasHostMode || hostModeBlockedBy === "" ? undefined : HOST_MODE_BLOCKED[hostModeBlockedBy];

    const options: {
        value: AppDockerApiMode;
        title: string;
        description: string;
        icon: typeof Container;
        locked?: string;
    }[] = [
        {
            value: "proxy",
            title: "Through the HivePaaS proxy",
            description:
                "Recommended. The app starts containers of the images listed below, on its own network, within the limits. It never gets the Docker socket.",
            icon: Container,
        },
        {
            value: "host",
            title: "The node's Docker socket",
            description:
                "For an app the proxy cannot serve, such as a Docker manager. The app gets /var/run/docker.sock of the node it runs on.",
            icon: ServerCog,
            locked: hostLocked,
        },
    ];

    return (
        <InfoBlock
            titleWidth={TITLE_WIDTH}
            title={
                <LabelWithInfo
                    label="Mode"
                    content="How the app reaches the Docker API."
                />
            }
        >
            <div className="grid max-w-[720px] grid-cols-1 gap-3 md:grid-cols-2">
                {options.map(option => {
                    const selected = field.value === option.value;
                    const Icon = option.locked ? Lock : option.icon;
                    return (
                        <button
                            key={option.value}
                            type="button"
                            aria-pressed={selected}
                            disabled={option.locked !== undefined}
                            onClick={() => {
                                field.onChange(option.value);
                            }}
                            className={cn(
                                "flex flex-col gap-1.5 rounded-lg border p-3 text-left transition-colors",
                                selected ? "border-primary bg-primary/5" : "hover:bg-muted/50",
                                option.locked && "cursor-not-allowed opacity-60",
                            )}
                        >
                            <span className="flex items-center gap-2 text-sm font-medium">
                                <Icon className="size-4 shrink-0" />
                                {option.title}
                            </span>
                            <span className="text-xs leading-relaxed text-muted-foreground">{option.description}</span>
                            {option.locked && (
                                <span className="text-xs font-medium text-destructive">{option.locked}</span>
                            )}
                        </button>
                    );
                })}
            </div>
        </InfoBlock>
    );
}

type ModeProps = {
    hostModeBlockedBy: AppDockerApiHostModeBlocker;
    /** The app is in host mode now: leaving it takes nothing, and it stays selectable. */
    hasHostMode: boolean;
};

/**
 * What host mode gives, said before it is saved. The placement note is there
 * because which node's socket the app gets is where it happens to run.
 */
export function DockerApiHostModeWarning() {
    return (
        <div className="space-y-2 rounded-lg border border-destructive/50 bg-destructive/5 p-4 text-sm">
            <p className="font-semibold text-destructive">The app will control the node it runs on.</p>
            <p className="text-muted-foreground">
                With the node&apos;s Docker socket, the app can start any container, with any privilege and any
                directory of the node. On a manager node that is the whole cluster: HivePaaS itself and the data of
                every other app. Give it only to an app you trust as much as an administrator.
            </p>
            <p className="text-muted-foreground">
                The app gets the socket of the node it runs on. Its placement settings decide which node that is.
            </p>
        </div>
    );
}
