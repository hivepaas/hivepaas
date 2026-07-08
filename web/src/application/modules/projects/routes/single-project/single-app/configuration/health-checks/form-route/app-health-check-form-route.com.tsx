import { useState } from "react";

import { toast } from "sonner";
import type { AppHealthChecks_REST_Payload, AppHealthChecks_Upsert_Payload } from "~/projects/api/services";
import { AppHealthChecksCommands } from "~/projects/data/commands";
import { AppHealthChecksQueries } from "~/projects/data/queries";
import { CreateOrEditAppHealthCheckForm } from "~/projects/dialogs/create-or-edit-app-health-check/form";
import type { CreateOrEditAppHealthCheckFormOutput } from "~/projects/dialogs/create-or-edit-app-health-check/schemas";
import {
    EAppHealthCheckRestMethod,
    EAppHealthCheckReturnBodyMode,
    EAppHealthCheckType,
} from "~/projects/module-shared/enums";

import { AppLoader, RouteFormHeader } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useConditionalModule } from "@application/shared/permissions";

type AppHealthCheckFormRouteMode = "create" | "edit";

const BODY_METHODS = [EAppHealthCheckRestMethod.POST, EAppHealthCheckRestMethod.PUT] as const;

function hasText(value: string): boolean {
    return value.trim().length > 0;
}

function getRestPayload(values: CreateOrEditAppHealthCheckFormOutput): AppHealthChecks_REST_Payload | null {
    if (values.healthcheckType !== EAppHealthCheckType.REST) {
        return null;
    }

    const returnText =
        values.rest.returnBodyMode === EAppHealthCheckReturnBodyMode.Text
            ? {
                  exact: values.rest.textExact,
                  regex: values.rest.textRegex,
              }
            : null;
    const returnJSON =
        values.rest.returnBodyMode === EAppHealthCheckReturnBodyMode.JSON
            ? {
                  exact: values.rest.jsonExact,
                  contain: values.rest.jsonContain,
              }
            : null;

    return {
        url: values.rest.url,
        method: values.rest.method,
        contentType: values.rest.contentType,
        body: BODY_METHODS.includes(values.rest.method as (typeof BODY_METHODS)[number]) ? values.rest.body : "",
        ...(hasText(values.rest.returnCode) ? { returnCode: values.rest.returnCode } : {}),
        returnText,
        returnJSON,
    };
}

function mapFormValuesToPayload(values: CreateOrEditAppHealthCheckFormOutput): AppHealthChecks_Upsert_Payload {
    return {
        availableInProjects: false,
        default: false,
        name: values.name,
        healthcheckType: values.healthcheckType,
        interval: values.interval,
        ...(values.maxRetry !== undefined && Number.isFinite(values.maxRetry) ? { maxRetry: values.maxRetry } : {}),
        ...(hasText(values.retryDelay) ? { retryDelay: values.retryDelay } : {}),
        ...(hasText(values.timeout) ? { timeout: values.timeout } : {}),
        saveResultTasks: true,
        rest: getRestPayload(values),
        grpc: values.healthcheckType === EAppHealthCheckType.GRPC ? values.grpc : null,
        notification: {
            successUseDefault: values.notification.successUseDefault,
            ...(!values.notification.successUseDefault && values.notification.success
                ? { success: values.notification.success }
                : {}),
            failureUseDefault: values.notification.failureUseDefault,
            ...(!values.notification.failureUseDefault && values.notification.failure
                ? { failure: values.notification.failure }
                : {}),
            ...(hasText(values.notification.minSendInterval)
                ? { minSendInterval: values.notification.minSendInterval }
                : {}),
        },
    };
}

export function AppHealthCheckFormRoute({ mode, projectId, appId, healthCheckId }: Props) {
    const [hasChanges, setHasChanges] = useState(false);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });
    const { navigate } = useAppNavigate();
    const isEditMode = mode === "edit";

    function navigateToList() {
        navigate.modules(ROUTE.projects.single.apps.single.configuration.healthChecks.$route(projectId, appId), {
            ignorePrevPath: true,
        });
    }

    const detailQuery = AppHealthChecksQueries.useFindOneById(
        {
            projectID: projectId,
            appID: appId,
            healthCheckID: healthCheckId ?? "",
        },
        {
            enabled: isEditMode && Boolean(healthCheckId),
        },
    );
    const healthCheck = detailQuery.data?.data;

    const { mutate: createHealthCheck, isPending: isCreating } = AppHealthChecksCommands.useCreateOne({
        onSuccess: () => {
            toast.success("App health check created successfully");
            navigateToList();
        },
    });

    const { mutate: updateHealthCheck, isPending: isUpdating } = AppHealthChecksCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("App health check updated successfully");
            navigateToList();
        },
    });

    function onSubmit(values: CreateOrEditAppHealthCheckFormOutput) {
        if (!canWrite) {
            return;
        }

        const payload = mapFormValuesToPayload(values);

        if (isEditMode && healthCheck) {
            updateHealthCheck({
                projectID: projectId,
                appID: appId,
                healthCheckID: healthCheck.id,
                payload: {
                    ...payload,
                    default: healthCheck.default,
                    saveResultTasks: healthCheck.saveResultTasks,
                    updateVer: healthCheck.updateVer,
                },
            });
            return;
        }

        if (!isEditMode) {
            createHealthCheck({
                projectID: projectId,
                appID: appId,
                payload,
            });
        }
    }

    function handleClose(): void {
        if (canWrite && hasChanges) {
            const userConfirmed = window.confirm("Are you sure you want to close without saving changes?");
            if (!userConfirmed) {
                return;
            }
        }

        setHasChanges(false);
        navigateToList();
    }

    const isDetailLoading = isEditMode && detailQuery.isFetching;
    const shouldRenderForm = mode === "create" || Boolean(healthCheck);

    return (
        <div className="flex w-full flex-col">
            <RouteFormHeader title={mode === "create" ? "Create Health Check" : "Edit Health Check"} />

            {isDetailLoading && (
                <div className="flex min-h-[220px] items-center justify-center">
                    <AppLoader />
                </div>
            )}

            {!isDetailLoading && shouldRenderForm && (
                <CreateOrEditAppHealthCheckForm
                    projectId={projectId}
                    isPending={isCreating || isUpdating}
                    onSubmit={onSubmit}
                    onHasChanges={setHasChanges}
                    initialValues={isEditMode ? healthCheck : undefined}
                    readOnly={!canWrite}
                    stickyActions
                    onClose={handleClose}
                />
            )}
        </div>
    );
}

interface Props {
    mode: AppHealthCheckFormRouteMode;
    projectId: string;
    appId: string;
    healthCheckId?: string;
}
