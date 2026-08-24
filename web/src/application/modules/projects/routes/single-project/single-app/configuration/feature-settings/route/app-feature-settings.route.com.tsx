import { useRef } from "react";

import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import type { AppFeatureSettings_UpdatePayload } from "~/projects/api/services";
import { AppFeatureSettingsCommands, AppFeatureSettingsQueries, ProjectAppsQueries } from "~/projects/data";
import { APP_CONFIGURATION_QUERY_OPTIONS } from "~/projects/data/constants";
import type { AppFeatureSettings } from "~/projects/domain";
import { ProjectPermissionSubmitButton } from "~/projects/module-shared/components";

import { AppLoader, FormActionBar } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PageError } from "@application/shared/pages";
import { useConditionalModule } from "@application/shared/permissions";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { AppFeatureSettingsForm } from "../form";
import { type AppFeatureSettingsFormSchemaOutput, DEFAULT_PREVIEW_CREATION_DELAY } from "../schemas";
import type { AppFeatureSettingsFormRef } from "../types";

function mapFormValuesToPayload(
    values: AppFeatureSettingsFormSchemaOutput,
    server: AppFeatureSettings,
): AppFeatureSettings_UpdatePayload {
    const creationDelay = values.previewSettings.creationDelay.trim() || DEFAULT_PREVIEW_CREATION_DELAY;

    return {
        inheritable: server.inheritable ?? false,
        default: server.default ?? false,
        updateVer: server.updateVer,
        loggingSettings: values.loggingSettings,
        schedJobSettings: values.schedJobSettings,
        terminalSettings: values.terminalSettings,
        previewSettings: {
            enabled: values.previewSettings.enabled,
            creationDelay,
            appsToClone: values.previewSettings.appsToClone.map(app => ({ id: app.id })),
            autoCloneApps: values.previewSettings.autoCloneApps,
            commands: values.previewSettings.commands.map(cmd => ({ id: cmd.id })),
        },
    };
}

export function AppFeatureSettingsRoute() {
    const { id: projectId, env, appId } = useParams<{ id: string; env: string; appId: string }>();
    const formRef = useRef<AppFeatureSettingsFormRef>(null);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });

    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");
    invariant(appId, "appId must be defined");

    const {
        data: appData,
        isLoading: isLoadingApp,
        error: appError,
    } = ProjectAppsQueries.useFindOneById({
        projectID: projectId,
        env,
        appID: appId,
    });
    const isPreviewApp = Boolean(appData?.data.parentApp);

    const { data, isLoading, error, refetch } = AppFeatureSettingsQueries.useFindOne(
        {
            projectID: projectId,
            env,
            appID: appId,
        },
        APP_CONFIGURATION_QUERY_OPTIONS,
    );

    const { mutate: update, isPending } = AppFeatureSettingsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Feature settings updated");
        },
        onError: err => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            } else if (err instanceof Error) {
                toast.error(err.message);
            } else {
                toast.error("Failed to update feature settings");
            }
        },
    });

    function handleSubmit(values: AppFeatureSettingsFormSchemaOutput) {
        if (!canWrite) {
            return;
        }

        invariant(projectId, "projectId must be defined");
        invariant(env, "env must be defined");
        invariant(appId, "appId must be defined");
        invariant(data, "feature settings data must be defined");

        update({
            projectID: projectId,
            env,
            appID: appId,
            payload: mapFormValuesToPayload(values, data.data),
        });
    }

    if (isLoading || isLoadingApp) {
        return <AppLoader />;
    }

    const currentError = error ?? appError;
    if (currentError) {
        return (
            <PageError
                error={currentError}
                onRetry={refetch}
            />
        );
    }

    invariant(data, "feature settings data must be defined");

    return (
        <AppFeatureSettingsForm
            ref={formRef}
            projectID={projectId}
            env={env}
            appID={appId}
            defaultValues={data.data}
            onSubmit={handleSubmit}
            readOnly={!canWrite}
            hidePreviewSettings={isPreviewApp}
        >
            <FormActionBar>
                <ProjectPermissionSubmitButton isPending={isPending} />
            </FormActionBar>
        </AppFeatureSettingsForm>
    );
}
