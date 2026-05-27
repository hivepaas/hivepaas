import { useEffect, useRef } from "react";

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

    const { mutate: update, isPending: isUpdating } = ProjectsCommands.useUpdateOne({
        onError: err => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            }
        },
    });
    const { mutate: updatePhoto, isPending: isUpdatingPhoto } = ProjectsCommands.useUpdatePhoto({});

    function handleSubmit(values: ProjectGeneralFormSchemaOutput) {
        invariant(projectId, "projectId must be defined");
        invariant(data, "data must be defined");

        const { photoUpload } = values;

        update(
            {
                projectID: projectId,
                name: values.name,
                envs: values.envs,
                tags: values.tags,
                note: values.note,
                owner: { id: values.ownerId },
                updateVer: data.data.updateVer,
                status: data.data.status,
            },
            {
                onSuccess: () => {
                    if (!photoUpload) {
                        toast.success("Project information updated");
                        return;
                    }

                    updatePhoto(
                        {
                            projectID: projectId,
                            photo: photoUpload,
                        },
                        {
                            onSuccess: () => {
                                toast.success("Project information updated");
                            },
                        },
                    );
                },
            },
        );
    }

    useEffect(() => {
        const project = data?.data;

        if (!project) {
            return;
        }

        formRef.current?.setValues({
            photo: project.photo === "" ? null : project.photo,
            photoUpload: null,
            name: project.name,
            envs: project.envs,
            tags: project.tags,
            note: project.note,
            ownerId: project.owner.id,
        });
    }, [data?.data]);

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
                        disabled={isUpdating || isUpdatingPhoto}
                        isLoading={isUpdating || isUpdatingPhoto}
                    >
                        Save
                    </Button>
                </div>
            </ProjectGeneralForm>
        </ProjectWithSidebar>
    );
}
