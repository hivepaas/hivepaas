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
    host: "a directory of the host's",
    tmpfs: "memory",
    dropped: "not mounted",
};

function samePort(input: ComposePortInput, port: ComposePortView): boolean {
    return input.published === port.published && input.target === port.target && input.protocol === port.protocol;
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
    const setPort = (name: string, port: ComposePortView, patch: Partial<ComposePortInput>) => {
        const current = inputs[name] ?? { ports: [] };
        const existing = current.ports.find(input => samePort(input, port)) ?? {
            published: port.published,
            target: port.target,
            protocol: port.protocol,
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
                        onImageChange={image => {
                            setImage(service.name, image);
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

function ServiceCard({ service, image, onImageChange, onPortChange }: CardProps) {
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
                            key={`${port.published}-${port.target}-${port.protocol}`}
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
                    {service.volumes.map(volume => (
                        <li key={volume.target}>
                            <code className="font-mono text-foreground">{volume.target}</code> {describeVolume(volume)}
                        </li>
                    ))}
                </ul>
            )}

            {service.aliases.length > 0 && (
                <p className="text-xs text-muted-foreground">
                    Reached as {service.aliases.map(alias => `"${alias}"`).join(", ")} too.
                </p>
            )}
            {service.dropped.length > 0 && (
                <p className="text-xs text-amber-700 dark:text-amber-400">Left out: {service.dropped.join(", ")}.</p>
            )}
        </div>
    );
}

function describeVolume(volume: ComposeVolumeView): string {
    const parts = [`← ${volume.source || "anonymous"}:`, VOLUME_KIND_LABELS[volume.kind]];
    if (volume.owner) {
        parts.push(`(${volume.owner}'s)`);
    }
    if (volume.readOnly) {
        parts.push("read only");
    }

    return parts.join(" ");
}

function PortRow({ port, onChange }: PortProps) {
    const label = `${port.published ? `${port.published} → ` : ""}${port.target}/${port.protocol}`;

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
    onImageChange: (image: string) => void;
    onPortChange: (port: ComposePortView, patch: Partial<ComposePortInput>) => void;
}

interface PortProps {
    port: ComposePortView;
    onChange: (patch: Partial<ComposePortInput>) => void;
}
