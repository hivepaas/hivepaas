import { useState } from "react";

import { Link } from "react-router";
import { toast } from "sonner";
import { HivePaaSRegistrySettingsCommands } from "~/system-settings/data";
import type { HivePaaSRegistrySettings } from "~/system-settings/domain";
import { SectionHeader } from "~/system-settings/module-shared";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

import { Button } from "@/components/ui";

import { RotateRegistryCredentialDialog } from "./rotate-credential.dialog.com";

const BYTES_PER_UNIT = 1024;
const SIZE_UNITS = ["B", "KB", "MB", "GB", "TB"] as const;

function formatBytes(bytes: number): string {
    let value = bytes;
    let unit = 0;
    while (value >= BYTES_PER_UNIT && unit < SIZE_UNITS.length - 1) {
        value /= BYTES_PER_UNIT;
        unit += 1;
    }
    return `${value.toFixed(unit === 0 ? 0 : 1)} ${SIZE_UNITS[unit]}`;
}

function formatDate(value?: string): string {
    if (!value) {
        return "";
    }
    const parsed = new Date(value);
    return Number.isNaN(parsed.getTime()) ? "" : parsed.toLocaleString();
}

interface Props {
    settings: HivePaaSRegistrySettings;
    canWrite: boolean;
}

/**
 * What is running, and the three things an operator does to a registry that is:
 * check what answers at its address, check that a large push survives whatever
 * is in front of it, and rotate the password.
 */
export function RegistryStatusSection({ settings, canWrite }: Props) {
    const status = settings.registryStatus;
    const [isRotateOpen, setIsRotateOpen] = useState(false);

    const probe = HivePaaSRegistrySettingsCommands.useProbeDomain();
    const pushCheck = HivePaaSRegistrySettingsCommands.useCheckPush();
    const rotate = HivePaaSRegistrySettingsCommands.useRotateCredential({
        onSuccess: response => {
            setIsRotateOpen(false);
            toast.success(`Password rotated. The previous one works until ${formatDate(response.data.graceEndsAt)}.`);
        },
    });

    if (!status?.provisioned) {
        return null;
    }

    const probeResult = probe.data?.data;
    const pushResult = pushCheck.data?.data;
    const graceEndsAt = formatDate(settings.credentialRotation?.graceEndsAt);

    return (
        <>
            <SectionHeader>Status</SectionHeader>
            <div className="flex flex-col gap-6 px-3">
                <InfoBlock
                    titleWidth={220}
                    title={
                        <LabelWithInfo
                            label="Registry"
                            content="Whether the registry answered its own API just now. Unreachable is normal while it restarts after a save."
                        />
                    }
                >
                    {status.reachable ? (
                        <span>Running, and answering at {settings.domain}</span>
                    ) : (
                        <span className="text-amber-700 dark:text-amber-400">
                            Provisioned, but not answering{status.unreachable ? `: ${status.unreachable}` : ""}
                        </span>
                    )}
                </InfoBlock>

                {status.reachable && (
                    <InfoBlock
                        titleWidth={220}
                        title={
                            <LabelWithInfo
                                label="What it holds"
                                content="One repository per app, and the size its images add up to."
                            />
                        }
                    >
                        {status.repositories === 0
                            ? "Nothing yet"
                            : `${status.repositories} ${status.repositories === 1 ? "repository" : "repositories"}, ${formatBytes(status.storedBytes)}`}
                    </InfoBlock>
                )}

                {settings.credential?.id && (
                    <InfoBlock
                        titleWidth={220}
                        title={
                            <LabelWithInfo
                                label="Credential"
                                content="The registry auth every build pushes with. It is an ordinary one, so any app can be pointed at it."
                            />
                        }
                    >
                        <div className="flex flex-col gap-1">
                            <Link
                                to={ROUTE.settings.registryAuth.edit.$route(settings.credential.id)}
                                className="text-primary hover:underline"
                            >
                                System registry
                            </Link>
                            {graceEndsAt && (
                                <span className="text-xs text-muted-foreground">
                                    Rotated. The previous password works until {graceEndsAt} — redeploy your apps before
                                    then.
                                </span>
                            )}
                        </div>
                    </InfoBlock>
                )}

                <InfoBlock
                    titleWidth={220}
                    title={
                        <LabelWithInfo
                            label="Checks"
                            content="The address check reads what answers there. The push check uploads 150 MB and throws it away, which is the only way to find a body limit before a build does."
                        />
                    }
                >
                    <div className="flex flex-col gap-2">
                        <div className="flex flex-wrap gap-2">
                            <Button
                                type="button"
                                variant="outline"
                                size="sm"
                                disabled={!canWrite || probe.isPending}
                                isLoading={probe.isPending}
                                onClick={() => {
                                    probe.mutate({ domain: settings.domain });
                                }}
                            >
                                Check the address
                            </Button>
                            <Button
                                type="button"
                                variant="outline"
                                size="sm"
                                disabled={!canWrite || pushCheck.isPending}
                                isLoading={pushCheck.isPending}
                                onClick={() => {
                                    pushCheck.mutate({});
                                }}
                            >
                                Test a large push
                            </Button>
                            <Button
                                type="button"
                                variant="outline"
                                size="sm"
                                disabled={!canWrite || rotate.isPending}
                                onClick={() => {
                                    setIsRotateOpen(true);
                                }}
                            >
                                Rotate password
                            </Button>
                        </div>

                        {probeResult && (
                            <div className="text-sm">
                                {!probeResult.reached && (
                                    <span className="text-muted-foreground">
                                        Nothing is answering at that address yet.
                                    </span>
                                )}
                                {probeResult.reached && !probeResult.proxied && (
                                    <span className="text-muted-foreground">
                                        The registry answered for itself — nothing is proxying this address.
                                    </span>
                                )}
                                {probeResult.reached && probeResult.proxied && (
                                    <div className="text-amber-700 dark:text-amber-400">
                                        <p className="font-medium">Something is answering instead of the registry.</p>
                                        <ul className="ml-4 list-disc">
                                            {probeResult.evidence.map(line => (
                                                <li key={line}>{line}</li>
                                            ))}
                                        </ul>
                                        <p>Set this record to DNS-only so that layers reach the cluster directly.</p>
                                    </div>
                                )}
                            </div>
                        )}

                        {pushResult && (
                            <p
                                className={
                                    pushResult.ok
                                        ? "text-sm text-muted-foreground"
                                        : "text-sm text-amber-700 dark:text-amber-400"
                                }
                            >
                                {pushResult.detail} ({Math.round(pushResult.elapsedMs / 1000)}s)
                            </p>
                        )}
                    </div>
                </InfoBlock>
            </div>

            <RotateRegistryCredentialDialog
                open={isRotateOpen}
                onOpenChange={setIsRotateOpen}
                onConfirm={() => {
                    rotate.mutate({});
                }}
                isPending={rotate.isPending}
            />
        </>
    );
}
