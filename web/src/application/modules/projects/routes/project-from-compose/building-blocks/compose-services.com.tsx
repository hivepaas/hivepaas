import { Input, Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui";
import { Badge } from "@components/ui/badge";
import type {
    ComposePortAs,
    ComposePortInput,
    ComposePortView,
    ComposeServiceInput,
    ComposeServiceView,
    ComposeVolumeKind,
    ComposeVolumeView,
} from "~/operations/domain";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

const PORT_AS_LABELS: Record<ComposePortAs, string> = {
    domain: "A domain",
    node: "A port on the nodes",
    none: "None",
};

const VOLUME_KIND_LABELS: Record<ComposeVolumeKind, string> = {
    volume: "a directory of the project's volume",
    shared: "another app's directory",
    file: "a file of the env's",
    files: "the files given, read only",
    host: "a directory of the host's",
    tmpfs: "memory",
    dropped: "not mounted",
};

function samePort(input: ComposePortInput, port: ComposePortView): boolean {
    return (
        input.published === port.published &&
        input.target === port.target &&
        input.protocol === port.protocol &&
        (input.source ?? "") === port.source
    );
}

/**
 * Each service as the app it becomes: its image, how many run, what each
 * published port becomes, where its volumes land, the names it is reached by,
 * and what of it is left out.
 */
export function ComposeServices({ services, inputs, onChange }: Props) {
    if (services.length === 0) {
        return null;
    }

    const setImage = (name: string, image: string) => {
        const current = inputs[name] ?? { ports: [] };
        onChange({ ...inputs, [name]: { ...current, image: image === "" ? undefined : image } });
    };
    const setChoice = (name: string, choice: Pick<ComposeServiceInput, "app" | "useExisting">) => {
        const current = inputs[name] ?? { ports: [] };
        onChange({ ...inputs, [name]: { ...current, app: choice.app, useExisting: choice.useExisting } });
    };
    const setPort = (name: string, port: ComposePortView, patch: Partial<ComposePortInput>) => {
        const current = inputs[name] ?? { ports: [] };
        const existing = current.ports.find(input => samePort(input, port)) ?? {
            published: port.published,
            target: port.target,
            protocol: port.protocol,
            source: port.source || undefined,
            as: port.as,
            domain: port.domain,
        };
        const others = current.ports.filter(input => !samePort(input, port));
        onChange({ ...inputs, [name]: { ...current, ports: [...others, { ...existing, ...patch }] } });
    };

    return (
        <InfoBlock
            titleWidth={220}
            title={
                <LabelWithInfo
                    label="Services"
                    content="An app per service, named after it, reached by the same name as in the file. Change what a published port becomes; give an image to a service the file only builds."
                />
            }
        >
            <div className="flex w-full max-w-[900px] flex-col gap-3">
                {services.map(service => (
                    <ServiceCard
                        key={service.name}
                        service={service}
                        image={inputs[service.name]?.image}
                        input={inputs[service.name]}
                        onImageChange={image => {
                            setImage(service.name, image);
                        }}
                        onChoice={choice => {
                            setChoice(service.name, choice);
                        }}
                        onPortChange={(port, patch) => {
                            setPort(service.name, port, patch);
                        }}
                    />
                ))}
            </div>
        </InfoBlock>
    );
}

function ServiceCard({ service, image, input, onImageChange, onChoice, onPortChange }: CardProps) {
    if (service.useExisting) {
        return (
            <div className="flex flex-col gap-3 rounded-md border px-3 py-3">
                <div className="flex flex-wrap items-center gap-2">
                    <span className="text-sm font-medium text-foreground">{service.name}</span>
                    <Badge variant="outline">the env&apos;s app</Badge>
                </div>
                <ExistingChoice
                    service={service}
                    input={input}
                    onChoice={onChoice}
                />
            </div>
        );
    }

    return (
        <div className="flex flex-col gap-3 rounded-md border px-3 py-3">
            <div className="flex flex-wrap items-center gap-2">
                <span className="text-sm font-medium text-foreground">{service.name}</span>
                {service.app !== service.name && (
                    <span className="text-xs text-muted-foreground">
                        app <code className="font-mono">{service.app}</code>
                    </span>
                )}
                {service.skipped ? (
                    <Badge
                        variant="outline"
                        className="border-amber-500/40 text-amber-700 dark:text-amber-400"
                    >
                        not created
                    </Badge>
                ) : (
                    <Badge variant="outline">
                        {service.mode === "replicated" ? `${service.replicas} replica(s)` : service.mode}
                    </Badge>
                )}
            </div>

            {(service.existing !== "" || input?.app !== undefined) && (
                <ExistingChoice
                    service={service}
                    input={input}
                    onChoice={onChoice}
                />
            )}

            {service.skipped || image !== undefined ? (
                <Input
                    value={image ?? ""}
                    onChange={event => {
                        onImageChange(event.target.value.trim());
                    }}
                    placeholder="registry/image:tag - the file builds it, and HivePaaS builds from a repository"
                    className="font-mono text-xs"
                />
            ) : (
                <code className="font-mono text-xs text-muted-foreground">{service.image}</code>
            )}
            {service.build && !service.skipped && image === undefined && (
                <p className="text-xs text-muted-foreground">
                    The file builds this image too: the image is deployed, and the build left out.
                </p>
            )}

            {service.ports.length > 0 && (
                <div className="flex flex-col gap-2">
                    {service.ports.map(port => (
                        <PortRow
                            key={`${port.published}-${port.target}-${port.protocol}-${port.source}`}
                            port={port}
                            onChange={patch => {
                                onPortChange(port, patch);
                            }}
                        />
                    ))}
                </div>
            )}

            {service.volumes.length > 0 && (
                <ul className="flex flex-col gap-1 text-xs text-muted-foreground">
                    {service.volumes
                        .filter(volume => !(service.dockerSocket && volume.target === service.dockerSocket))
                        .map(volume => (
                            <li key={volume.target}>
                                <code className="font-mono text-foreground">{volume.target}</code>{" "}
                                {describeVolume(volume)}
                            </li>
                        ))}
                </ul>
            )}
            {service.dockerSocket && (
                <p className="text-xs text-amber-700 dark:text-amber-400">
                    The Docker socket, at <code className="font-mono">{service.dockerSocket}</code>, is not mounted.
                    Once the app is created, give it the Docker API in its Docker API settings: through the proxy, as
                    configured there - its socket is at <code className="font-mono">$DOCKER_HOST</code> - or the
                    node&apos;s own socket, at <code className="font-mono">/var/run/docker.sock</code>, which takes an
                    administrator and privileged apps.
                </p>
            )}

            {service.aliases.length > 0 && (
                <p className="text-xs text-muted-foreground">
                    Reached as {service.aliases.map(alias => `"${alias}"`).join(", ")} too.
                </p>
            )}
            {service.secrets.length > 0 && (
                <p className="text-xs text-muted-foreground">
                    Kept as secrets of the app: {service.secrets.join(", ")}.
                </p>
            )}
            {service.dropped.length > 0 && (
                <p className="text-xs text-amber-700 dark:text-amber-400">Left out: {service.dropped.join(", ")}.</p>
            )}
        </div>
    );
}

type ExistingAction = "use" | "rename";

/**
 * What to do with a service whose name or key an app of the env answers to:
 * use that app, or create the service under another key. Until one is chosen,
 * nothing is created.
 */
function ExistingChoice({ service, input, onChoice }: ExistingProps) {
    const action: ExistingAction | undefined = input?.useExisting
        ? "use"
        : input?.app !== undefined
          ? "rename"
          : undefined;

    return (
        <div className="flex flex-col gap-2 rounded-md border border-amber-500/40 bg-amber-500/5 px-3 py-2">
            <p className="text-xs text-amber-700 dark:text-amber-400">
                {service.existing ? (
                    <>
                        The env has an app, <code className="font-mono">{service.existing}</code>, reached by this
                        service&apos;s name.
                    </>
                ) : (
                    <>The env has an app reached by this service&apos;s name.</>
                )}
            </p>
            <Select
                value={action ?? ""}
                onValueChange={value => {
                    onChoice(
                        value === "use"
                            ? { useExisting: true }
                            : { app: input?.app ?? `${service.app || service.name}-2` },
                    );
                }}
            >
                <SelectTrigger className="w-[320px]">
                    <SelectValue placeholder="Choose what to do with it" />
                </SelectTrigger>
                <SelectContent>
                    <SelectItem value="use">Use the env&apos;s app, as it is</SelectItem>
                    <SelectItem value="rename">Create it under another key</SelectItem>
                </SelectContent>
            </Select>
            {action === "use" && (
                <p className="text-xs text-muted-foreground">
                    Not created: the other services reach the env&apos;s app by this name.
                </p>
            )}
            {action === "rename" && (
                <>
                    <Input
                        value={input?.app ?? ""}
                        onChange={event => {
                            onChoice({ app: event.target.value.trim().toLowerCase() });
                        }}
                        placeholder="app key"
                        className="w-[320px] font-mono text-xs"
                    />
                    <p className="text-xs text-muted-foreground">
                        The other services still reach the env&apos;s app by this name: change the file where they refer
                        to it.
                    </p>
                </>
            )}
        </div>
    );
}

function describeVolume(volume: ComposeVolumeView): string {
    const parts = [`← ${volume.source || "anonymous"}:`, VOLUME_KIND_LABELS[volume.kind]];
    if (volume.owner) {
        parts.push(`(${volume.owner}'s)`);
    }
    if (volume.files > 0 && volume.kind !== "files") {
        parts.push(`with ${volume.files} ${volume.files === 1 ? "file" : "files"} given, read only,`);
    }
    if (volume.readOnly) {
        parts.push("read only");
    }

    return parts.join(" ");
}

function PortRow({ port, onChange }: PortProps) {
    const label = `${port.published ? `${port.published} → ` : ""}${port.target}/${port.protocol}`;
    const fromLabels = port.source === "labels";

    return (
        <div className="flex flex-wrap items-center gap-2">
            <code className="w-[140px] font-mono text-xs text-foreground">{label}</code>
            <Select
                value={port.as}
                onValueChange={value => {
                    const as = value as ComposePortAs;
                    onChange(as === "domain" ? { as, domain: port.domain || port.suggested } : { as });
                }}
            >
                <SelectTrigger className="w-[200px]">
                    <SelectValue />
                </SelectTrigger>
                <SelectContent>
                    {(Object.keys(PORT_AS_LABELS) as ComposePortAs[]).map(as => (
                        <SelectItem
                            key={as}
                            value={as}
                            disabled={
                                (as === "domain" && port.protocol !== "tcp") || (as === "node" && !port.published)
                            }
                        >
                            {PORT_AS_LABELS[as]}
                            {as === port.default ? " (default)" : ""}
                        </SelectItem>
                    ))}
                </SelectContent>
            </Select>
            {port.as === "domain" && (
                <Input
                    value={port.domain}
                    onChange={event => {
                        onChange({ as: "domain", domain: event.target.value.trim().toLowerCase() });
                    }}
                    placeholder={port.suggested || "app.example.com"}
                    className="min-w-[240px] flex-1 font-mono text-xs"
                />
            )}
            {fromLabels && (
                <p className="w-full text-xs text-muted-foreground">
                    From the service&apos;s Traefik labels
                    {port.also.length > 0 && port.as === "domain"
                        ? `: ${port.also.join(", ")} ${port.also.length === 1 ? "is a domain" : "are domains"} of the app too`
                        : ""}
                    .
                </p>
            )}
        </div>
    );
}

interface Props {
    services: ComposeServiceView[];
    inputs: Record<string, ComposeServiceInput>;
    onChange: (inputs: Record<string, ComposeServiceInput>) => void;
}

interface CardProps {
    service: ComposeServiceView;
    image: string | undefined;
    input: ComposeServiceInput | undefined;
    onImageChange: (image: string) => void;
    onChoice: (choice: Pick<ComposeServiceInput, "app" | "useExisting">) => void;
    onPortChange: (port: ComposePortView, patch: Partial<ComposePortInput>) => void;
}

interface ExistingProps {
    service: ComposeServiceView;
    input: ComposeServiceInput | undefined;
    onChoice: (choice: Pick<ComposeServiceInput, "app" | "useExisting">) => void;
}

interface PortProps {
    port: ComposePortView;
    onChange: (patch: Partial<ComposePortInput>) => void;
}
