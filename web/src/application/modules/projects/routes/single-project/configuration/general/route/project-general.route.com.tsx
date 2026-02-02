import { useRef } from "react";

import { Button } from "@components/ui";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { ProjectsQueries } from "~/projects/data";
import { ProjectsCommands } from "~/projects/data/commands";
import { ProjectWithSidebar } from "~/projects/module-shared/components";

import { AppLoader } from "@application/shared/components";
import { PageError } from "@application/shared/pages";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { ProjectGeneralForm } from "../form";
import { type ProjectGeneralFormSchemaOutput } from "../schemas";
import { type ProjectGeneralFormRef } from "../types";

export function ProjectGeneralRoute() {
    const { id: projectId } = useParams<{ id: string }>();
    const formRef = useRef<ProjectGeneralFormRef>(null);

    invariant(projectId, "projectId must be defined");

    const { data, isLoading, error, refetch } = ProjectsQueries.useFindOneById({ projectID: projectId });

    const { mutate: update, isPending } = ProjectsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Project information updated");
        },
        onError: err => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            }
        },
    });

    function handleSubmit(values: ProjectGeneralFormSchemaOutput) {
        invariant(projectId, "projectId must be defined");
        invariant(data, "data must be defined");

        update({
            projectID: projectId,
            ...values,
            updateVer: data.data.updateVer,
            status: data.data.status,
        });
    }

    if (isLoading) {
        return <AppLoader />;
    }

    if (error) {
        return (
            <PageError
                error={error}
                onRetry={refetch}
            />
        );
    }

    invariant(data, "data must be defined");

    const { data: project } = data;

    return (
        <ProjectWithSidebar projectId={projectId}>
            <ProjectGeneralForm
                ref={formRef}
                defaultValues={project}
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
            </ProjectGeneralForm>
        </ProjectWithSidebar>
    );
}
