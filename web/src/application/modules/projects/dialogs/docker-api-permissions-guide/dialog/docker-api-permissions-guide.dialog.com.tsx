import type { ReactNode } from "react";

import { Badge } from "@components/ui/badge";
import {
    Dialog,
    DialogActionFooter,
    DialogBody,
    DialogDescription,
    DialogFixedContent,
    DialogHeader,
    DialogTitle,
} from "@components/ui/dialog";
import { cn } from "@lib/utils";
import { Ban, Check, CircleSlash, ShieldCheck, Users } from "lucide-react";

import { Button } from "@/components/ui/button";

import { useDockerApiPermissionsGuideDialogState } from "../hooks";

import {
    CORE_AREAS,
    type GuideEndpoint,
    type HttpMethod,
    NEVER_ALLOWED,
    PERMISSION_GUIDES,
    type PermissionGuide,
    type RiskLevel,
    TEMPLATE_EXAMPLES,
} from "./docker-api-permissions-guide.content";

const METHOD_STYLES: Record<HttpMethod, string> = {
    GET: "border-emerald-500/40 text-emerald-700 dark:text-emerald-400",
    HEAD: "border-slate-500/40 text-slate-600 dark:text-slate-400",
    POST: "border-sky-500/40 text-sky-700 dark:text-sky-400",
    PUT: "border-amber-500/40 text-amber-700 dark:text-amber-400",
    DELETE: "border-red-500/40 text-red-700 dark:text-red-400",
};

const RISK_STYLES: Record<RiskLevel, { label: string; className: string }> = {
    low: {
        label: "Low risk",
        className: "border-emerald-500/40 bg-emerald-500/10 text-emerald-700 dark:text-emerald-400",
    },
    medium: {
        label: "Medium risk",
        className: "border-amber-500/40 bg-amber-500/10 text-amber-700 dark:text-amber-400",
    },
};

const sectionId = (value: string) => `docker-api-permission-${value}`;

/**
 * What each permission of the Docker API screen lets an app do: the endpoints,
 * the rules the proxy holds them to, what is refused without it, who needs it,
 * and what it risks. Long, so it is a scrolling dialog rather than a tooltip.
 */
export function DockerApiPermissionsGuideDialog() {
    const { state, props: dialogOptions, ...actions } = useDockerApiPermissionsGuideDialogState();

    function handleClose() {
        actions.close();
        dialogOptions?.onClose?.();
    }

    return (
        <Dialog
            open={state.mode === "open"}
            onOpenChange={isOpen => {
                if (!isOpen) {
                    handleClose();
                }
            }}
        >
            <DialogFixedContent className="sm:max-w-[1200px]">
                <DialogHeader>
                    <DialogTitle>Docker API permissions</DialogTitle>
                    <DialogDescription>
                        What an app may do through the Docker API proxy of HivePaaS. Everything is refused unless a rule
                        below lets it through, and everything the app reaches is what it created itself.
                    </DialogDescription>
                </DialogHeader>

                <DialogBody className="flex flex-col gap-6 px-6 pb-6">
                    <nav className="flex flex-wrap items-center gap-2">
                        <span className="text-xs font-medium text-muted-foreground">Jump to:</span>
                        {PERMISSION_GUIDES.map(guide => (
                            <Button
                                key={guide.value}
                                type="button"
                                variant="outline"
                                size="sm"
                                className="h-7 gap-1.5 px-2.5 text-xs"
                                onClick={() => {
                                    document
                                        .getElementById(sectionId(guide.value))
                                        ?.scrollIntoView({ behavior: "smooth", block: "start" });
                                }}
                            >
                                <guide.icon className="size-3.5" />
                                {guide.label}
                            </Button>
                        ))}
                    </nav>

                    <CoreSection />

                    {PERMISSION_GUIDES.map(guide => (
                        <PermissionSection
                            key={guide.value}
                            guide={guide}
                        />
                    ))}

                    <NeverAllowedSection />
                    <ExamplesSection />
                </DialogBody>

                <DialogActionFooter>
                    <Button
                        type="button"
                        onClick={handleClose}
                    >
                        Close
                    </Button>
                </DialogActionFooter>
            </DialogFixedContent>
        </Dialog>
    );
}

function CoreSection() {
    return (
        <section className="rounded-xl border bg-muted/30 p-4">
            <div className="mb-3 flex items-center gap-2">
                <ShieldCheck className="size-5 text-emerald-600 dark:text-emerald-400" />
                <h3 className="text-base font-semibold">Always allowed</h3>
                <span className="text-sm text-muted-foreground">
                    - the core, which every app with Docker API access has, whatever is ticked
                </span>
            </div>
            <div className="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-3">
                {CORE_AREAS.map(area => (
                    <div
                        key={area.title}
                        className="flex flex-col gap-2 rounded-lg border bg-background p-3"
                    >
                        <div className="flex items-center gap-2 text-sm font-semibold">
                            <area.icon className="size-4 text-muted-foreground" />
                            {area.title}
                        </div>
                        <ul className="flex flex-col gap-1 text-xs leading-relaxed text-muted-foreground">
                            {area.items.map(item => (
                                <li
                                    key={item}
                                    className="flex gap-1.5"
                                >
                                    <Check className="mt-0.5 size-3 shrink-0 text-emerald-600" />
                                    <span>{item}</span>
                                </li>
                            ))}
                        </ul>
                        {area.endpoints.length > 0 && <EndpointList endpoints={area.endpoints} />}
                    </div>
                ))}
            </div>
        </section>
    );
}

function PermissionSection({ guide }: { guide: PermissionGuide }) {
    const risk = RISK_STYLES[guide.risk.level];
    return (
        <section
            id={sectionId(guide.value)}
            className="scroll-mt-2 rounded-xl border p-4"
        >
            <div className="flex flex-wrap items-start justify-between gap-3">
                <div className="flex items-start gap-3">
                    <div className="rounded-lg bg-sky-500/10 p-2 text-sky-700 dark:text-sky-400">
                        <guide.icon className="size-5" />
                    </div>
                    <div className="flex flex-col gap-0.5">
                        <div className="flex items-center gap-2">
                            <h3 className="text-base font-semibold">{guide.label}</h3>
                            <code className="rounded bg-muted px-1.5 py-0.5 text-[11px] text-muted-foreground">
                                {guide.value}
                            </code>
                        </div>
                        <p className="text-sm text-muted-foreground">{guide.summary}</p>
                    </div>
                </div>
                <Badge
                    variant="outline"
                    className={cn("h-6 px-2 text-xs", risk.className)}
                >
                    {risk.label}
                </Badge>
            </div>

            <div className="mt-4 grid grid-cols-1 gap-4 lg:grid-cols-3">
                <div className="flex flex-col gap-4">
                    <Block title="What it allows">
                        <BulletList
                            items={guide.allows}
                            icon="check"
                        />
                    </Block>
                    <Block title="Without it">
                        <p className="flex gap-1.5 text-xs leading-relaxed text-muted-foreground">
                            <CircleSlash className="mt-0.5 size-3 shrink-0" />
                            <span>{guide.withoutIt}</span>
                        </p>
                    </Block>
                </div>

                <Block title="Endpoints">
                    {guide.endpoints.length > 0 ? (
                        <EndpointList endpoints={guide.endpoints} />
                    ) : (
                        <p className="text-xs leading-relaxed text-muted-foreground">{guide.endpointsNote}</p>
                    )}
                </Block>

                <Block title="Rules the proxy holds it to">
                    <BulletList
                        items={guide.rules}
                        icon="dot"
                    />
                </Block>
            </div>

            <div className="mt-4 grid grid-cols-1 gap-3 lg:grid-cols-2">
                <div className="flex items-start gap-2 rounded-lg bg-muted/40 p-3 text-xs">
                    <Users className="mt-0.5 size-3.5 shrink-0 text-muted-foreground" />
                    <div className="flex flex-col gap-1">
                        <span className="font-medium">Who needs it</span>
                        {guide.usedBy.map(user => (
                            <span
                                key={user}
                                className="text-muted-foreground"
                            >
                                {user}
                            </span>
                        ))}
                    </div>
                </div>
                <div className={cn("flex flex-col gap-1 rounded-lg border p-3 text-xs", risk.className)}>
                    <span className="font-medium">{risk.label}</span>
                    <span>{guide.risk.text}</span>
                </div>
            </div>
        </section>
    );
}

function NeverAllowedSection() {
    return (
        <section className="rounded-xl border border-destructive/40 bg-destructive/5 p-4">
            <div className="mb-3 flex items-center gap-2">
                <Ban className="size-5 text-destructive" />
                <h3 className="text-base font-semibold">Never allowed</h3>
                <span className="text-sm text-muted-foreground">- whatever is ticked</span>
            </div>
            <div className="flex flex-wrap gap-2">
                {NEVER_ALLOWED.map(item => (
                    <Badge
                        key={item}
                        variant="outline"
                        className="h-6 border-destructive/40 px-2 text-xs text-destructive"
                    >
                        {item}
                    </Badge>
                ))}
            </div>
            <p className="mt-3 text-xs text-muted-foreground">
                An app that needs any of these cannot be served by the proxy. An administrator can give it the
                node&apos;s own Docker socket instead - host mode - with the Privileged Apps switch on.
            </p>
        </section>
    );
}

function ExamplesSection() {
    return (
        <section className="rounded-xl border p-4">
            <h3 className="mb-3 text-base font-semibold">What the templates ask for</h3>
            <div className="overflow-x-auto">
                <table className="w-full text-sm">
                    <thead>
                        <tr className="border-b text-left text-xs text-muted-foreground">
                            <th className="py-2 pr-4 font-medium">Template</th>
                            {PERMISSION_GUIDES.map(guide => (
                                <th
                                    key={guide.value}
                                    className="px-2 py-2 text-center font-medium"
                                >
                                    {guide.label}
                                </th>
                            ))}
                            <th className="py-2 pl-4 font-medium">Why</th>
                        </tr>
                    </thead>
                    <tbody>
                        {TEMPLATE_EXAMPLES.map(example => (
                            <tr
                                key={example.template}
                                className="border-b last:border-0 align-top"
                            >
                                <td className="py-2 pr-4 font-medium">{example.template}</td>
                                {PERMISSION_GUIDES.map(guide => (
                                    <td
                                        key={guide.value}
                                        className="px-2 py-2 text-center"
                                    >
                                        {example.allow.includes(guide.value) ? (
                                            <Check className="mx-auto size-4 text-emerald-600" />
                                        ) : (
                                            <span className="text-muted-foreground">-</span>
                                        )}
                                    </td>
                                ))}
                                <td className="py-2 pl-4 text-xs leading-relaxed text-muted-foreground">
                                    {example.why}
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </section>
    );
}

function Block({ title, children }: { title: string; children: ReactNode }) {
    return (
        <div className="flex flex-col gap-2">
            <h4 className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">{title}</h4>
            {children}
        </div>
    );
}

function BulletList({ items, icon }: { items: string[]; icon: "check" | "dot" }) {
    return (
        <ul className="flex flex-col gap-1.5 text-xs leading-relaxed">
            {items.map(item => (
                <li
                    key={item}
                    className="flex gap-1.5"
                >
                    {icon === "check" ? (
                        <Check className="mt-0.5 size-3 shrink-0 text-emerald-600" />
                    ) : (
                        <span className="mt-1.5 size-1 shrink-0 rounded-full bg-muted-foreground" />
                    )}
                    <span>{item}</span>
                </li>
            ))}
        </ul>
    );
}

function EndpointList({ endpoints }: { endpoints: GuideEndpoint[] }) {
    return (
        <ul className="flex flex-col gap-1.5">
            {endpoints.map(endpoint => (
                <li
                    key={endpoint.methods.join() + endpoint.path}
                    className="flex flex-col gap-0.5 rounded-md border bg-muted/20 px-2 py-1.5"
                >
                    <div className="flex flex-wrap items-center gap-1.5">
                        {endpoint.methods.map(method => (
                            <Badge
                                key={method}
                                variant="outline"
                                className={cn("h-4 px-1 font-mono text-[10px]", METHOD_STYLES[method])}
                            >
                                {method}
                            </Badge>
                        ))}
                        <code className="break-all font-mono text-[11px]">{endpoint.path}</code>
                    </div>
                    <span className="text-[11px] text-muted-foreground">{endpoint.note}</span>
                </li>
            ))}
        </ul>
    );
}
