import { memo } from "react";

import { Button } from "@components/ui";
import { Trash2 } from "lucide-react";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { ProjectsQueries } from "~/projects/data";
import { ProjectsCommands } from "~/projects/data/commands";

import { BackButton, TabNavigation } from "@application/shared/components";
import { PopConfirm } from "@application/shared/components/pop-confirm";
import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

import { ProjectStatusBadge } from "@application/modules/projects/module-shared/components";

import { SingleProjectBreadcrumbs } from "../buidling-blocks";

import { SingleProjectHeaderSkeleton } from "./single-project-header.skeleton.com";

function View({ projectId }: Props) {
    const { data, isLoading, error } = ProjectsQueries.useFindOneById({ projectID: projectId });

    const { navigate } = useAppNavigate();

    const { mutate: deleteOne, isPending: isDeleting } = ProjectsCommands.useDeleteOne({});

    if (isLoading) {
        return <SingleProjectHeaderSkeleton />;
    }

    if (error) {
        return null;
    }

    invariant(data, "data must be defined");
    const { data: project } = data;

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

    const links = [
        {
            route: ROUTE.projects.single.configuration.general.$route(projectId),
            label: "Configuration",
        },
        {
            route: ROUTE.projects.single.apps.$route(projectId),
            label: "Apps",
        },
    ];
    return (
        <div className="bg-background pt-4 px-5 rounded-lg">
            <div className="flex items-center justify-between">
                <SingleProjectBreadcrumbs project={project} />
                <div className="flex items-center gap-2">
                    <PopConfirm
                        title="Remove project"
                        variant="destructive"
                        confirmText="Remove"
                        cancelText="Cancel"
                        description="Confirm deletion of this item?"
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
                        <div className="flex items-center gap-2">
                            <h2 className="text-[20px] font-semibold text-foreground">{project.name}</h2>
                            <ProjectStatusBadge status={project.status} />
                        </div>
                    </div>
                </div>
            </div>

            <div className="border-b border-border" />

            <TabNavigation links={links} />
        </div>
    );
}

interface Props {
    projectId: string;
}

export const SingleProjectHeader = memo(View);
