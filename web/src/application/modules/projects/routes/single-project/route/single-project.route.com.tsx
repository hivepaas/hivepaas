import { useParams } from "react-router";
import invariant from "tiny-invariant";
// TODO: Import when API is available
// import { ProjectsCommands, ProjectsQueries } from "~/projects/data";

import { AppLoader } from "@application/shared/components";

// TODO: Import when form is created
// import { SingleProjectForm } from "../form";
// import { type SingleProjectFormSchemaOutput } from "../schemas";
// import { type SingleProjectFormRef } from "../types";

export function SingleProjectRoute() {
    const { id: projectId } = useParams<{ id: string }>();

    invariant(projectId, "projectId must be defined");

    // TODO: Implement when API is available
    // const { data, isLoading, error } = ProjectsQueries.useFindOneById({ id: projectId });
    const data = null;
    const isLoading = false;
    const error: Error | null = null;

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
        return <div className="text-red-500">Error: {error.message}</div>;
    }

    // TODO: Uncomment when API is available
    // invariant(data, "data must be defined");
    // const { data: project } = data;

    return (
        <div className="bg-background rounded-lg p-4 max-w-7xl w-full mx-auto">
            <div className="text-muted-foreground">
                {/* TODO: Implement SingleProjectForm when ready */}
                <p>Single Project Route - Project ID: {projectId}</p>
                <p className="mt-2 text-sm">This component needs to be implemented with the project form.</p>
            </div>
        </div>
    );
}

