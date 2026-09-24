import { useEffect, useRef, useState } from "react";

import { EyeOffIcon } from "lucide-react";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { AppContainerSettingsCommands, AppContainerSettingsQueries } from "~/projects/data";
import { APP_CONFIGURATION_QUERY_OPTIONS } from "~/projects/data/constants";
import {
    type AppContainerSettings,
    type Healthcheck,
    type LogDriver,
    type Privileges,
    type RestartPolicy,
} from "~/projects/domain";
import { ProjectPermissionSubmitButton } from "~/projects/module-shared/components";
import { ERestartPolicyCondition } from "~/projects/module-shared/enums";
import { ConfirmRevealSecretsDialog, RevealSecretsButton } from "~/settings/module-shared/components";

import { AppLoader, FormActionBar } from "@application/shared/components";
import { CAPABILITY_IDS, MODULE_IDS } from "@application/shared/constants";
import { useProfileContext } from "@application/shared/context";
import { EUserRole } from "@application/shared/enums";
import { useCapability, useConditionalModule } from "@application/shared/permissions";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { Button } from "@/components/ui/button";

import { AppConfigContainerSettingsForm } from "../form";
import { type AppConfigContainerSettingsFormSchemaOutput } from "../schemas";
import { type AppConfigContainerSettingsFormRef } from "../types";

function buildRestartPolicy(values: AppConfigContainerSettingsFormSchemaOutput): RestartPolicy | null {
    const rp = values.restartPolicy;
    const delay = rp.delay.trim() === "" ? null : rp.delay;
    const window = rp.window.trim() === "" ? null : rp.window;
    const maxAttempts = rp.maxAttempts ?? null;

    if (rp.condition === ERestartPolicyCondition.None && delay == null && window == null && maxAttempts == null) {
        return null;
    }

    return {
        condition: rp.condition,
        delay,
        maxAttempts,
        window,
    };
}

function buildPrivileges(values: AppConfigContainerSettingsFormSchemaOutput): Privileges {
    const p = values.privileges;
    return {
        noNewPrivileges: p.noNewPrivileges,
        seLinuxContext: {
            disable: !p.selinuxEnabled,
            user: p.seLinuxUser,
            role: p.seLinuxRole,
            type: p.seLinuxType,
            level: p.seLinuxLevel,
        },
        seccomp: {
            mode: p.seccompMode,
            profile: p.seccompProfile,
        },
        appArmor: {
            mode: p.appArmorMode,
        },
    };
}

function buildLabels(rows: AppConfigContainerSettingsFormSchemaOutput["serviceLabels"]): Record<string, string> {
    const record: Record<string, string> = {};
    for (const row of rows) {
        const key = row.key.trim();
        if (key) {
            record[key] = row.value;
        }
    }
    return record;
}

function buildHealthcheck(values: AppConfigContainerSettingsFormSchemaOutput): Healthcheck | null {
    const { healthcheck } = values;
    if (!healthcheck.enabled) {
        return null;
    }

    return {
        enabled: true,
        mode: healthcheck.mode,
        command: healthcheck.command,
        interval: healthcheck.interval,
        timeout: healthcheck.timeout,
        startPeriod: healthcheck.startPeriod,
        startInterval: healthcheck.startInterval,
        retries: healthcheck.retries ?? 0,
    };
}

function buildLogDriver(values: AppConfigContainerSettingsFormSchemaOutput): LogDriver | null {
    const driver = values.logDriver.driver.trim();
    if (!driver) {
        return null;
    }

    return {
        name: driver,
        options: buildLabels(values.logDriver.options),
    };
}

function mapFormValuesToPayload(
    values: AppConfigContainerSettingsFormSchemaOutput,
    server: AppContainerSettings | undefined,
): AppContainerSettings {
    const g = values.general;
    const groups = g.groups.trim().split(/\s+/).filter(Boolean);

    return {
        serviceLabels: buildLabels(values.serviceLabels),
        containerLabels: buildLabels(values.containerLabels),
        image: g.image,
        command: g.command,
        workingDir: g.workingDir,
        hostname: g.hostname,
        user: g.user,
        groups,
        stopSignal: g.stopSignal,
        init: g.init === "auto" ? null : g.init === "on",
        tty: g.tty,
        openStdin: g.openStdin,
        readOnly: g.readOnly,
        stopGracePeriod: g.stopGracePeriod.trim() === "" ? null : g.stopGracePeriod,
        privileges: buildPrivileges(values),
        healthcheck: buildHealthcheck(values),
        restartPolicy: buildRestartPolicy(values),
        logDriver: buildLogDriver(values),
        updateVer: server?.updateVer ?? 0,
    };
}

/**
 * Reveals or hides the labels HivePaaS and Docker manage. Revealing them takes
 * what revealing secrets takes, so it is offered to whoever could reveal those.
 */
function SystemLabelsRevealButton({
    isRevealed,
    onReveal,
    onHide,
    isLoading,
}: {
    isRevealed: boolean;
    onReveal: () => void;
    onHide: () => void;
    isLoading: boolean;
}) {
    if (isRevealed) {
        return (
            <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={onHide}
                className="gap-1.5 shrink-0 px-2.5 sm:px-3 text-xs sm:text-sm h-8 sm:h-9"
            >
                <EyeOffIcon className="size-3.5 sm:size-4" />
                <span>Hide System Labels</span>
            </Button>
        );
    }

    return (
        <RevealSecretsButton
            onClick={onReveal}
            isLoading={isLoading}
            label="Reveal System Labels"
        />
    );
}

export function AppConfigContainerSettingsRoute() {
    const { id: projectId, env, appId } = useParams<{ id: string; env: string; appId: string }>();
    const formRef = useRef<AppConfigContainerSettingsFormRef>(null);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });
    const [isSystemLabelsRevealed, setIsSystemLabelsRevealed] = useState(false);
    const [isRevealDialogOpen, setIsRevealDialogOpen] = useState(false);

    const profile = useProfileContext(state => state.profile);
    const isAdmin = profile?.role === EUserRole.Admin;
    const { hasCapability: canRevealSecrets } = useCapability(CAPABILITY_IDS.SecretReveal);
    const canRevealSystemLabels = isAdmin || canRevealSecrets;

    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");
    invariant(appId, "appId must be defined");

    const { data, isLoading, isFetching, isError, error } = AppContainerSettingsQueries.useFindOne(
        {
            projectID: projectId,
            env,
            appID: appId,
            revealSystemLabels: isSystemLabelsRevealed,
        },
        {
            ...APP_CONFIGURATION_QUERY_OPTIONS,
            // A refused reveal is not retried: the answer will not change, and
            // every attempt is recorded.
            ...(isSystemLabelsRevealed ? { retry: false } : {}),
        },
    );

    // A refused reveal leaves the labels hidden, and says why.
    useEffect(() => {
        if (isError && isSystemLabelsRevealed) {
            setIsSystemLabelsRevealed(false);
            toast.error(error.message || "Failed to reveal system labels");
        }
    }, [isError, error, isSystemLabelsRevealed]);

    const { mutate: update, isPending } = AppContainerSettingsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Container settings updated");
        },
        onError: err => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            } else if (err instanceof Error) {
                toast.error(err.message);
            } else {
                toast.error("Failed to update container settings");
            }
        },
    });

    function handleSubmit(values: AppConfigContainerSettingsFormSchemaOutput) {
        if (!canWrite) {
            return;
        }

        invariant(projectId, "projectId must be defined");
        invariant(env, "env must be defined");
        invariant(appId, "appId must be defined");
        update({
            projectID: projectId,
            env,
            appID: appId,
            payload: mapFormValuesToPayload(values, data?.data),
        });
    }

    if (isLoading) {
        return <AppLoader />;
    }

    return (
        <div className="flex flex-col gap-4">
            <AppConfigContainerSettingsForm
                ref={formRef}
                defaultValues={data?.data}
                onSubmit={handleSubmit}
                readOnly={!canWrite}
                labelsToolbar={
                    canRevealSystemLabels ? (
                        <SystemLabelsRevealButton
                            isRevealed={isSystemLabelsRevealed}
                            onReveal={() => {
                                setIsRevealDialogOpen(true);
                            }}
                            onHide={() => {
                                setIsSystemLabelsRevealed(false);
                            }}
                            isLoading={isFetching && isSystemLabelsRevealed}
                        />
                    ) : undefined
                }
            >
                <FormActionBar>
                    <ProjectPermissionSubmitButton isPending={isPending} />
                </FormActionBar>
            </AppConfigContainerSettingsForm>

            <ConfirmRevealSecretsDialog
                open={isRevealDialogOpen}
                onOpenChange={setIsRevealDialogOpen}
                onConfirm={() => {
                    setIsSystemLabelsRevealed(true);
                    setIsRevealDialogOpen(false);
                }}
                isPending={isFetching && isSystemLabelsRevealed}
                title="Reveal System Labels"
                actionLabel="Reveal the system labels"
                note={
                    <p>
                        System labels are the ones HivePaaS and Docker manage: <code>hivepaas.*</code>,{" "}
                        <code>traefik.*</code> and <code>com.docker.stack.*</code>. They can carry credentials, which is
                        why they are revealed like secrets. Changes to them are not saved: saving keeps them as they
                        are.
                    </p>
                }
            />
        </div>
    );
}
