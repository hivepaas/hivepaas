import { useRef } from "react";

import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { AppCloneSettingsCommands, AppCloneSettingsQueries, ProjectsQueries } from "~/projects/data";
import { APP_CONFIGURATION_QUERY_OPTIONS } from "~/projects/data/constants";
import { ProjectPermissionSubmitButton, ProjectPermissionTooltipAction } from "~/projects/module-shared/components";

import { AppLoader, FormActionBar } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PageError } from "@application/shared/pages";
import { useConditionalModule } from "@application/shared/permissions";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { Button } from "@/components/ui/button";

import { AppCloneSettingsForm } from "../form";
import { mapFormToUpdatePayload } from "../form/app-clone-settings.form-mappers";
import type { AppCloneSettingsFormSchemaOutput } from "../schemas";
import type { AppCloneSettingsFormRef } from "../types";

export function AppCloneSettingsRoute() {
    const { id: projectId, env, appId } = useParams<{ id: string; env: string; appId: string }>();
    const formRef = useRef<AppCloneSettingsFormRef>(null);
    const pendingExecuteRef = useRef(false);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });

    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");
    invariant(appId, "appId must be defined");

    const { data, isLoading, error, refetch } = AppCloneSettingsQueries.useFindOne(
        {
            projectID: projectId,
            env,
            appID: appId,
        },
        APP_CONFIGURATION_QUERY_OPTIONS,
    );

    const {
        data: projectData,
        isLoading: isProjectLoading,
        error: projectError,
        refetch: refetchProject,
    } = ProjectsQueries.useFindOneById(
        {
            projectID: projectId,
        },
        APP_CONFIGURATION_QUERY_OPTIONS,
    );

    const { mutate: update, isPending: isUpdating } = AppCloneSettingsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("App clone settings saved");

            if (pendingExecuteRef.current) {
                pendingExecuteRef.current = false;
                execute({
                    projectID: projectId,
                    env,
                    appID: appId,
                });
            }
        },
        onError: err => {
            pendingExecuteRef.current = false;

            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            }
        },
    });

    const { mutate: execute, isPending: isExecuting } = AppCloneSettingsCommands.useExecute({
        onSuccess: () => {
            toast.success("App clone started");
        },
        onError: err => {
            if (err instanceof Error) {
                toast.error(err.message);
            } else {
                toast.error("Failed to execute app clone");
            }
        },
    });

    function handleSubmit(values: AppCloneSettingsFormSchemaOutput) {
        if (!canWrite) {
            return;
        }

        invariant(projectId, "projectId must be defined");
        invariant(env, "env must be defined");
        invariant(appId, "appId must be defined");
        invariant(data, "app clone settings data must be defined");

        update({
            projectID: projectId,
            env,
            appID: appId,
            payload: mapFormToUpdatePayload(values, data.data),
        });
    }

    function handleExecuteClick() {
        if (!canWrite) {
            return;
        }

        pendingExecuteRef.current = true;
        formRef.current?.submit();
    }

    if (isLoading || isProjectLoading) {
        return <AppLoader />;
    }

    if (error) {
        return (
            <PageError
                error={error}
                onRetry={() => {
                    void refetch();
                }}
            />
        );
    }

    if (projectError) {
        return (
            <PageError
                error={projectError}
                onRetry={() => {
                    void refetchProject();
                }}
            />
        );
    }

    invariant(data, "app clone settings data must be defined");
    invariant(projectData, "project data must be defined");

    const isPending = isUpdating || isExecuting;

    return (
        <AppCloneSettingsForm
            ref={formRef}
            projectId={projectId}
            env={env}
            envs={projectData.data.envs}
            defaultValues={data.data}
            onSubmit={handleSubmit}
            readOnly={!canWrite}
        >
            <FormActionBar>
                <ProjectPermissionTooltipAction
                    projectId={projectId}
                    action="write"
                >
                    {({ isDenied }) => (
                        <Button
                            type="button"
                            className="min-w-[120px]"
                            disabled={isPending || isDenied}
                            isLoading={isExecuting}
                            onClick={handleExecuteClick}
                        >
                            Save and Clone
                        </Button>
                    )}
                </ProjectPermissionTooltipAction>
                <ProjectPermissionSubmitButton
                    projectId={projectId}
                    isPending={isPending}
                />
            </FormActionBar>
        </AppCloneSettingsForm>
    );
}
