import { memo } from "react";

import { Button } from "@components/ui";
import { Trash2 } from "lucide-react";
import { toast } from "sonner";
import { ProjectsCommands } from "~/projects/data/commands";

// TODO: Import ProjectsQueries.useFindOneById when available
// import { ProjectsQueries } from "~/projects/data/queries";
// TODO: Import ProjectStatusBadge when needed
// import { ProjectStatusBadge } from "~/projects/module-shared/components";

import { BackButton } from "@application/shared/components";
import { PopConfirm } from "@application/shared/components/pop-confirm";
import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

function View({ projectId }: Props) {
    // TODO: Implement useFindOneById when API is available
    // const { data, isLoading, error } = ProjectsQueries.useFindOneById({ id: projectId });
    // const isLoading = false;
    // const error: Error | null = null;

    const { navigate } = useAppNavigate();

    const { mutate: deleteOne, isPending: isDeleting } = ProjectsCommands.useDeleteOne({});

    // TODO: Uncomment when API is available
    // invariant(data, "data must be defined");
    // const { data: project } = data;

    const handleRemove = () => {
        deleteOne(
            { projectID: projectId },
            {
                onSuccess: () => {
                    toast.success("Project removed successfully");
                    navigate.modules(ROUTE.projects.list.$route);
                },
            },
        );
    };

    // TODO: Replace with actual project data when API is available
    return (
        <div className="bg-background pt-4 px-5 rounded-lg">
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                    <span className="text-muted-foreground">Project: {projectId}</span>
                </div>
                <div className="flex items-center gap-2">
                    <PopConfirm
                        title="Remove project"
                        variant="destructive"
                        confirmText="Remove"
                        cancelText="Cancel"
                        description="Are you sure you want to remove this project?"
                        onConfirm={handleRemove}
                    >
                        <Button
                            variant="outline"
                            disabled={isDeleting}
                        >
                            <Trash2 className="mr-2 size-4" />
                            Remove
                        </Button>
                    </PopConfirm>
                </div>
            </div>

            <div className="flex items-center gap-4 mt-4 pb-4">
                <BackButton />
                <div className="flex items-center gap-4">
                    <div className="flex flex-col gap-3">
                        {/* TODO: Replace with actual project data when API is available */}
                        <div className="flex items-center gap-2">
                            <h2 className="text-[20px] font-semibold text-foreground">Project {projectId}</h2>
                            {/* <ProjectStatusBadge status={project.status} /> */}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}

interface Props {
    projectId: string;
}

export const SingleProjectHeader = memo(View);
