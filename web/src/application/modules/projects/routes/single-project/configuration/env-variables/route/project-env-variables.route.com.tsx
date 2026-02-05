import { useRef } from "react";

import { Button } from "@components/ui/button";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { ProjectEnvVarsCommands } from "~/projects/data/commands/project-env-vars";
import { ProjectEnvVarsQueries } from "~/projects/data/queries/project-env-vars";
import { ProjectWithSidebar } from "~/projects/module-shared/components";

import { AppLoader } from "@application/shared/components";
import { PageError } from "@application/shared/pages";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { ProjectEnvVarsForm } from "../form";
import type { ProjectEnvVarsFormSchemaOutput } from "../schemas";
import type { ProjectEnvVarsFormRef } from "../types";

export function ProjectEnvVariablesRoute() {
    const { id: projectId } = useParams<{ id: string }>();
    const formRef = useRef<ProjectEnvVarsFormRef>(null);

    invariant(projectId, "projectId must be defined");

    const {
        data: envVarsData,
        isLoading: isLoadingEnvVars,
        error: envVarsError,
        refetch: refetchEnvVars,
    } = ProjectEnvVarsQueries.useFindOne({ projectID: projectId });

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
        invariant(projectId, "projectId must be defined");
        invariant(envVarsData, "envVarsData must be defined");

        update({
            projectID: projectId,
            ...values,
            updateVer: envVarsData.data.updateVer,
        });
    }

    if (isLoadingEnvVars) {
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

    return (
        <ProjectWithSidebar projectId={projectId}>
            <ProjectEnvVarsForm
                ref={formRef}
                defaultValues={{
                    buildtime: envVars.buildtime,
                    runtime: envVars.runtime,
                }}
                onSubmit={handleSubmit}
            >
                <div className="flex justify-end gap-2">
                    <Button
                        type="submit"
                        className="min-w-[100px]"
                        disabled={isPending}
                        isLoading={isPending}
                    >
                        Save
                    </Button>
                </div>
            </ProjectEnvVarsForm>
        </ProjectWithSidebar>
    );
}
