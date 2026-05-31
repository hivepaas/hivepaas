import { memo } from "react";

import { Avatar, Button } from "@components/ui";
import { PlusCircle } from "lucide-react";
import invariant from "tiny-invariant";
import { ProjectsQueries } from "~/projects/data";
import { useProjectUserAccessesDialog } from "~/projects/dialogs/project-user-accesses";

import { BackButton, TabNavigation } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

import { ProjectStatusBadge } from "@application/modules/projects/module-shared/components";

import { SingleProjectBreadcrumbs } from "../buidling-blocks";

import { SingleProjectHeaderSkeleton } from "./single-project-header.skeleton.com";

function View({ projectId }: Props) {
    const { data, isLoading, error } = ProjectsQueries.useFindOneById({ projectID: projectId });

    const projectUserAccessesDialog = useProjectUserAccessesDialog();

    if (isLoading) {
        return <SingleProjectHeaderSkeleton />;
    }

    if (error) {
        return null;
    }

    invariant(data, "data must be defined");
    const { data: project } = data;
    const accessUsers = [project.owner, ...project.userAccesses].reduce<ProjectAccessUser[]>((users, user) => {
        if (users.some(item => item.id === user.id)) {
            return users;
        }

        return [
            ...users,
            {
                id: user.id,
                fullName: user.fullName,
                email: user.email,
                username: user.username,
                photo: user.photo,
            },
        ];
    }, []);
    const visibleAccessUsers = accessUsers.slice(0, 3);
    const extraAccessUsers = Math.max(accessUsers.length - visibleAccessUsers.length, 0);

    const links = [
        {
            route: ROUTE.projects.single.apps.$route(projectId),
            label: "Apps",
            activePathPrefixes: [ROUTE.projects.single.apps.$route(projectId)],
        },
        {
            route: ROUTE.projects.single.configuration.general.$route(projectId),
            label: "Configuration",
            activePathPrefixes: [ROUTE.projects.single.configuration.$route(projectId)],
        },
        {
            route: ROUTE.projects.single.providerConfiguration.accessTokens.$route(projectId),
            label: "Provider Configuration",
            activePathPrefixes: [ROUTE.projects.single.providerConfiguration.$route(projectId)],
        },
    ];
    return (
        <div className="bg-background pt-4 px-5 rounded-lg">
            <div className="flex items-center justify-between">
                <SingleProjectBreadcrumbs project={project} />
                {/* <div className="flex items-center gap-2">
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
                </div> */}
            </div>

            <div className="flex flex-wrap items-center gap-4 mt-4 pb-4">
                <BackButton />
                <div className="flex min-w-0 items-center gap-4">
                    <Avatar
                        name={project.name}
                        src={project.photo}
                        className="size-20 text-2xl"
                    />
                    <div className="flex min-w-0 flex-col gap-3">
                        <div className="flex items-center gap-2">
                            <h2 className="text-[20px] font-semibold text-foreground">{project.name}</h2>
                            <ProjectStatusBadge status={project.status} />
                        </div>
                        <div className="flex items-center gap-3">
                            <span className="text-sm text-muted-foreground">Owner</span>
                            <Avatar
                                name={project.owner.fullName}
                                src={project.owner.photo}
                                className="size-8"
                                borderless
                            />
                            <span className="truncate text-sm font-medium">{project.owner.fullName}</span>
                        </div>
                    </div>
                </div>

                <div className="ml-auto flex flex-wrap items-center justify-end gap-4">
                    <div className="flex items-center gap-3">
                        <span className="text-sm text-muted-foreground">Access</span>
                        <div className="flex -space-x-2">
                            {visibleAccessUsers.map(user => (
                                <Avatar
                                    key={user.id}
                                    name={user.fullName || user.email || user.username}
                                    src={user.photo}
                                    className="size-8 border-2 border-background"
                                />
                            ))}
                            {extraAccessUsers > 0 && (
                                <span className="flex size-8 items-center justify-center rounded-full border-2 border-background bg-primary/10 text-xs font-semibold text-primary">
                                    +{extraAccessUsers}
                                </span>
                            )}
                        </div>
                    </div>
                    <Button
                        type="button"
                        variant="outline"
                        onClick={() => {
                            projectUserAccessesDialog.actions.open(project.id, project.name);
                        }}
                    >
                        <PlusCircle className="size-4" />
                        Add User
                    </Button>
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

interface ProjectAccessUser {
    id: string;
    fullName: string;
    email: string;
    username: string;
    photo: string | null;
}

export const SingleProjectHeader = memo(View);
