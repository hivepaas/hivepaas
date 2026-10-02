import { useRef } from "react";

import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import {
    AppAutoscaleQueries,
    AppServiceSettingsCommands,
    AppServiceSettingsQueries,
    ProjectAppsQueries,
} from "~/projects/data";
import { APP_CONFIGURATION_QUERY_OPTIONS } from "~/projects/data/constants";
import { ProjectPermissionSubmitButton } from "~/projects/module-shared/components";
import { isFunctionApp } from "~/projects/module-shared/utils";

import { AppLink, AppLoader, FormActionBar } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useConditionalModule } from "@application/shared/permissions";

import { EServiceMode } from "@application/modules/projects/module-shared/enums";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { AutoscaleSection } from "../building-blocks";
import { AppConfigAvailabilityForm } from "../form";
import { type AppConfigAvailabilitySchemaOutput } from "../schemas";
import { type AppConfigAvailabilityFormRef } from "../types";

export function AppConfigAvailabilityRoute() {
    const { id: projectId, env, appId } = useParams<{ id: string; env: string; appId: string }>();
    const formRef = useRef<AppConfigAvailabilityFormRef>(null);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });

    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");
    invariant(appId, "appId must be defined");

    const { data, isLoading } = AppServiceSettingsQueries.useFindOne(
        {
            projectID: projectId,
            env,
            appID: appId,
        },
        APP_CONFIGURATION_QUERY_OPTIONS,
    );

    // The header's query: the app is already loaded. Only a function autoscales.
    const { data: app } = ProjectAppsQueries.useFindOneById({
        projectID: projectId,
        env,
        appID: appId,
        getStats: true,
    });
    const isFunction = app ? isFunctionApp(app.data) : false;
    const { data: autoscale, isLoading: isAutoscaleLoading } = AppAutoscaleQueries.useFindOne(
        { projectID: projectId, env, appID: appId },
        { ...APP_CONFIGURATION_QUERY_OPTIONS, enabled: isFunction },
    );

    const { mutate: update, isPending } = AppServiceSettingsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Availability and scaling settings updated");
        },
        onError: err => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            } else if (err instanceof Error) {
                toast.error(err.message);
            } else {
                toast.error("Failed to update availability and scaling settings");
            }
        },
    });

    function handleSubmit(values: AppConfigAvailabilitySchemaOutput) {
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
            payload: {
                updateVer: data?.data.updateVer ?? 0,
                modeSpec: {
                    mode: values.mode,
                    serviceReplicas: values.mode === EServiceMode.Replicated ? values.serviceReplicas : null,
                    jobMaxConcurrent: values.mode === EServiceMode.ReplicatedJob ? values.jobMaxConcurrent : null,
                    jobTotalCompletions: values.mode === EServiceMode.ReplicatedJob ? values.jobTotalCompletions : null,
                },
                placement: {
                    constraints: values.constraints,
                    preferences: values.preferences,
                },
            },
        });
    }

    if (isLoading || isAutoscaleLoading) {
        return <AppLoader />;
    }

    return (
        <div className="flex flex-col gap-4">
            <div className={cn(dashedBorderBox)}>
                <span className="font-semibold text-orange-500">Note:</span> If you change the configuration here,
                please check the application&rsquo;s scheduling results in{" "}
                <AppLink.Basic
                    to={ROUTE.projects.single.apps.single.instances.$route(projectId, env, appId)}
                    className="text-link underline-offset-4 hover:underline"
                >
                    Instances
                </AppLink.Basic>
                .
            </div>

            {isFunction && autoscale && (
                <>
                    <AutoscaleSection
                        projectId={projectId}
                        env={env}
                        appId={appId}
                        autoscale={autoscale.data}
                        readOnly={!canWrite}
                    />
                    <div className="h-px bg-muted" />
                </>
            )}

            <AppConfigAvailabilityForm
                ref={formRef}
                defaultValues={data?.data}
                onSubmit={handleSubmit}
                readOnly={!canWrite}
                autoscaled={isFunction && Boolean(autoscale?.data.enabled)}
            >
                <FormActionBar>
                    <ProjectPermissionSubmitButton isPending={isPending} />
                </FormActionBar>
            </AppConfigAvailabilityForm>
        </div>
    );
}
