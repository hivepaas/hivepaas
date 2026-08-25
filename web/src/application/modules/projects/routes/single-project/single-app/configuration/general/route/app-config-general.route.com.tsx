import { useEffect, useRef } from "react";

import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { ProjectAppsCommands } from "~/projects/data/commands";
import { APP_CONFIGURATION_QUERY_OPTIONS } from "~/projects/data/constants";
import { ProjectAppsQueries, ProjectsQueries } from "~/projects/data/queries";
import { ProjectPermissionSubmitButton } from "~/projects/module-shared/components";

import { AppLoader, FormActionBar } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PageError } from "@application/shared/pages";
import { useConditionalModule } from "@application/shared/permissions";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { AppConfigGeneralForm } from "../form";
import { type AppConfigGeneralFormSchemaOutput } from "../schemas";
import { type AppConfigGeneralFormRef } from "../types";

export function AppConfigGeneralRoute() {
    const { id: projectId, env, appId } = useParams<{ id: string; env: string; appId: string }>();
    const formRef = useRef<AppConfigGeneralFormRef>(null);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });

    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");
    invariant(appId, "appId must be defined");

    const { data, isLoading, error, refetch } = ProjectAppsQueries.useFindOneById(
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

    const { mutate: update, isPending: isUpdating } = ProjectAppsCommands.useUpdateOne({
        onError: err => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            }
        },
    });
    const { mutate: updatePhoto, isPending: isUpdatingPhoto } = ProjectAppsCommands.useUpdatePhoto({});

    function handleSubmit(values: AppConfigGeneralFormSchemaOutput) {
        if (!canWrite) {
            return;
        }

        invariant(projectId, "projectId must be defined");
        invariant(env, "env must be defined");
        invariant(appId, "appId must be defined");
        invariant(data, "data must be defined");

        const { photoUpload, name, tags, note } = values;

        update(
            {
                projectID: projectId,
                env,
                appID: appId,
                name,
                tags,
                note,
                updateVer: data.data.updateVer,
            },
            {
                onSuccess: () => {
                    if (!photoUpload) {
                        toast.success("App information updated");
                        return;
                    }

                    updatePhoto(
                        {
                            projectID: projectId,
                            env,
                            appID: appId,
                            photo: photoUpload,
                        },
                        {
                            onSuccess: () => {
                                toast.success("App information updated");
                            },
                        },
                    );
                },
            },
        );
    }

    useEffect(() => {
        const app = data?.data;

        if (!app) {
            return;
        }

        formRef.current?.setValues({
            photo: app.photo === "" ? null : app.photo,
            photoUpload: null,
            name: app.name,
            tags: app.tags,
            note: app.note,
        });
    }, [data?.data]);

    if (isLoading || isProjectLoading) {
        return <AppLoader />;
    }

    if (error || projectError) {
        const pageError = error ?? projectError;
        invariant(pageError, "pageError must be defined");

        return (
            <PageError
                error={pageError}
                onRetry={() => {
                    void refetch();
                    void refetchProject();
                }}
            />
        );
    }

    invariant(data, "data must be defined");
    invariant(projectData, "projectData must be defined");

    const { data: app } = data;
    const { envs } = projectData.data;

    return (
        <AppConfigGeneralForm
            ref={formRef}
            defaultValues={app}
            envs={envs}
            projectID={projectId}
            env={env}
            appID={appId}
            onSubmit={handleSubmit}
            readOnly={!canWrite}
        >
            <FormActionBar>
                <ProjectPermissionSubmitButton isPending={isUpdating || isUpdatingPhoto} />
            </FormActionBar>
        </AppConfigGeneralForm>
    );
}
