import { useRef } from "react";

import { Button } from "@components/ui";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { AppContainerSettingsCommands, AppContainerSettingsQueries } from "~/projects/data";
import { type AppContainerSettings, type Privileges, type RestartPolicy } from "~/projects/domain";
import { ERestartPolicyCondition } from "~/projects/module-shared/enums";

import { AppLoader } from "@application/shared/components";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { AppConfigContainerSettingsForm } from "../form";
import { type AppConfigContainerSettingsFormSchemaOutput } from "../schemas";
import { type AppConfigContainerSettingsFormRef } from "../types";

function buildRestartPolicy(values: AppConfigContainerSettingsFormSchemaOutput): RestartPolicy | null {
    const rp = values.restartPolicy;
    const delay = rp.delay.trim() === "" ? null : rp.delay;
    const window = rp.window.trim() === "" ? null : rp.window;
    const maxRaw = rp.maxAttempts.trim();
    const maxAttempts = maxRaw === "" ? null : Number.isNaN(Number(maxRaw)) ? null : Number(maxRaw);

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

function mapFormValuesToPayload(
    values: AppConfigContainerSettingsFormSchemaOutput,
    server: AppContainerSettings | undefined,
): AppContainerSettings {
    const labels: Record<string, string> = {};
    for (const row of values.labels) {
        const k = row.key.trim();
        if (k) labels[k] = row.value;
    }

    const g = values.general;
    const groups = g.groups.trim().split(/\s+/).filter(Boolean);

    return {
        labels,
        image: g.image,
        command: g.command,
        workingDir: g.workingDir,
        hostname: g.hostname,
        user: g.user,
        groups,
        stopSignal: g.stopSignal,
        tty: g.tty,
        openStdin: g.openStdin,
        readOnly: g.readOnly,
        stopGracePeriod: g.stopGracePeriod.trim() === "" ? null : g.stopGracePeriod,
        privileges: buildPrivileges(values),
        healthcheck: server?.healthcheck ?? null,
        restartPolicy: buildRestartPolicy(values),
        updateVer: server?.updateVer ?? 0,
    };
}

export function AppConfigContainerSettingsRoute() {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();
    const formRef = useRef<AppConfigContainerSettingsFormRef>(null);

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    const { data, isLoading } = AppContainerSettingsQueries.useFindOne({
        projectID: projectId,
        appID: appId,
    });

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
        invariant(projectId, "projectId must be defined");
        invariant(appId, "appId must be defined");
        update({
            projectID: projectId,
            appID: appId,
            payload: mapFormValuesToPayload(values, data?.data),
        });
    }

    if (isLoading) {
        return <AppLoader />;
    }

    return (
        <div className="flex flex-col gap-4">
            <div className="rounded-md border border-dashed border-primary bg-accent p-2 text-sm text-muted-foreground">
                For configuration details, see{" "}
                <a
                    href="https://docs.docker.com/reference/api/engine/version/v1.52/#tag/Service/operation/ServiceUpdate"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-500 underline underline-offset-2"
                >
                    docs
                </a>
            </div>

            <AppConfigContainerSettingsForm
                ref={formRef}
                defaultValues={data?.data}
                onSubmit={handleSubmit}
            >
                <div className="flex justify-end mt-4">
                    <Button
                        type="submit"
                        className="min-w-[100px]"
                        disabled={isPending}
                        isLoading={isPending}
                    >
                        Save
                    </Button>
                </div>
            </AppConfigContainerSettingsForm>
        </div>
    );
}
