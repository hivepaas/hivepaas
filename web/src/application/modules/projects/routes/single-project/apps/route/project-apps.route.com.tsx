import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { AppLoader } from "@application/shared/components";
import { PageError } from "@application/shared/pages";

import { ProjectsQueries } from "@application/modules/projects/data";

export function ProjectAppsRoute() {
    const { id: projectId } = useParams<{ id: string }>();

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

    // TODO: Uncomment when API is available
    // invariant(data, "data must be defined");
    invariant(data, "data must be defined");
    const { data: project } = data;

    return (
        <div className="bg-background rounded-lg p-4 max-w-7xl w-full mx-auto">
            <div className="text-muted-foreground">
                {/* TODO: Implement SingleProjectForm when ready */}
                <p>Project Apps Route - Project ID: {project.name}</p>
            </div>
        </div>
    );
}
