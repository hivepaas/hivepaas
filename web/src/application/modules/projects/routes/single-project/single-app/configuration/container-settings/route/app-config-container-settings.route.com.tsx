import { useRef } from "react";

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

import { AppLoader } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { useConditionalModule } from "@application/shared/permissions";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

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
        init: g.init,
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

export function AppConfigContainerSettingsRoute() {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();
    const formRef = useRef<AppConfigContainerSettingsFormRef>(null);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    const { data, isLoading } = AppContainerSettingsQueries.useFindOne(
        {
            projectID: projectId,
            appID: appId,
        },
        APP_CONFIGURATION_QUERY_OPTIONS,
    );

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
                readOnly={!canWrite}
            >
                <div className="flex justify-end mt-4">
                    <ProjectPermissionSubmitButton isPending={isPending} />
                </div>
            </AppConfigContainerSettingsForm>
        </div>
    );
}
