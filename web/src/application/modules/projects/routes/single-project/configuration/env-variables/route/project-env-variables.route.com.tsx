import { useRef } from "react";

import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { ProjectsQueries } from "~/projects/data";
import { ProjectEnvVarsCommands } from "~/projects/data/commands/project-env-vars";
import { ProjectEnvVarsQueries } from "~/projects/data/queries/project-env-vars";
import { ProjectPermissionSubmitButton } from "~/projects/module-shared/components";
import {
    PROJECT_ENV_FILTER_ALL,
    getProjectEnvFilterParam,
    useSelectedProjectEnv,
} from "~/projects/module-shared/hooks";

import { AppLoader, FormActionBar } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PageError } from "@application/shared/pages";
import { useConditionalModule } from "@application/shared/permissions";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { ProjectEnvVarsForm } from "../form";
import type { ProjectEnvVarsFormSchemaOutput } from "../schemas";
import type { ProjectEnvVarsFormRef } from "../types";

export function ProjectEnvVariablesRoute() {
    const { id: projectId } = useParams<{ id: string }>();
    const formRef = useRef<ProjectEnvVarsFormRef>(null);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });

    invariant(projectId, "projectId must be defined");

    const selectedEnv = useSelectedProjectEnv(projectId);
    const scopedEnv = getProjectEnvFilterParam(selectedEnv);
    const { data: projectData } = ProjectsQueries.useFindOneById({ projectID: projectId });
    const projectEnvs = projectData?.data.envs ?? [];

    const {
        data: envVarsData,
        isLoading: isLoadingEnvVars,
        error: envVarsError,
        refetch: refetchEnvVars,
        isFetching: isFetchingEnvVars,
    } = ProjectEnvVarsQueries.useFindOne({ projectID: projectId, env: scopedEnv });

    const { mutate: update, isPending } = ProjectEnvVarsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Environment variables updated");
        },
        onError: (err: Error) => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            }
        },
    });

    function handleSubmit(values: ProjectEnvVarsFormSchemaOutput) {
        if (!canWrite) {
            return;
        }

        invariant(projectId, "projectId must be defined");
        invariant(envVarsData, "envVarsData must be defined");

        update({
            projectID: projectId,
            env: scopedEnv,
            ...values,
            updateVer: envVarsData.data.updateVer,
        });
    }

    // Show loader on initial load or when switching env scope before cached data arrives.
    if (isLoadingEnvVars || (isFetchingEnvVars && !envVarsData)) {
        return <AppLoader />;
    }

    if (envVarsError) {
        return (
            <PageError
                error={envVarsError}
                onRetry={refetchEnvVars}
            />
        );
    }

    invariant(envVarsData, "envVarsData must be defined");

    const { data: envVars } = envVarsData;
    const scopeKey = scopedEnv ?? PROJECT_ENV_FILTER_ALL;

    return (
        <ProjectEnvVarsForm
            key={scopeKey}
            ref={formRef}
            selectedEnv={selectedEnv}
            envs={projectEnvs}
            defaultValues={{
                buildtime: envVars.buildtime,
                runtime: envVars.runtime,
            }}
            inheritedValues={{
                buildtime: envVars.inheritedBuildtimeEnvVars,
                runtime: envVars.inheritedRuntimeEnvVars,
            }}
            onSubmit={handleSubmit}
            readOnly={!canWrite}
        >
            <FormActionBar contentClassName="gap-2">
                <ProjectPermissionSubmitButton isPending={isPending} />
            </FormActionBar>
        </ProjectEnvVarsForm>
    );
}
