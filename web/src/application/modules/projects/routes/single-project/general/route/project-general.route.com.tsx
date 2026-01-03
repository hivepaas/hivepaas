import { useRef } from "react";

import { Button } from "@components/ui";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
// TODO: Import ProjectsCommands.useUpdateOne when API is available
// import { ProjectsCommands } from "~/projects/data/commands";
import { ProjectsQueries } from "~/projects/data";

import { AppLoader } from "@application/shared/components";
import { PageError } from "@application/shared/pages";

// TODO: Import when API is available
// import { isValidationException } from "@infrastructure/api";
// import { ValidationException } from "@infrastructure/exceptions/validation";

import { ProjectGeneralForm } from "../form";
import { type ProjectGeneralFormSchemaOutput } from "../schemas";
import { type ProjectGeneralFormRef } from "../types";

export function ProjectGeneralRoute() {
    const { id: projectId } = useParams<{ id: string }>();
    const formRef = useRef<ProjectGeneralFormRef>(null);

    invariant(projectId, "projectId must be defined");

    const { data, isLoading, error, refetch } = ProjectsQueries.useFindOneById({ projectID: projectId });

    // TODO: Implement when API is available
    // const { mutate: update, isPending } = ProjectsCommands.useUpdateOne({
    //     onSuccess: () => {
    //         toast.success("Project information updated");
    //     },
    //     onError: err => {
    //         if (isValidationException(err)) {
    //             formRef.current?.onError(ValidationException.fromHttp(err));
    //         }
    //     },
    // });

    function handleSubmit(_values: ProjectGeneralFormSchemaOutput) {
        invariant(projectId, "projectId must be defined");
        invariant(data, "data must be defined");

        // TODO: Implement when API is available
        // update({
        //     projectID: projectId,
        //     ..._values,
        //     updateVer: data.data.updateVer,
        // });

        toast.info("Update functionality will be available soon");
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
        <div className="bg-background rounded-lg p-4 max-w-7xl w-full mx-auto">
            <ProjectGeneralForm
                ref={formRef}
                defaultValues={project}
                onSubmit={handleSubmit}
            >
                <div className="flex justify-end mt-4">
                    <Button
                        type="submit"
                        className="min-w-[100px]"
                        // TODO: Enable when API is available
                        // disabled={isPending}
                        // isLoading={isPending}
                    >
                        Save
                    </Button>
                </div>
            </ProjectGeneralForm>
        </div>
    );
}
