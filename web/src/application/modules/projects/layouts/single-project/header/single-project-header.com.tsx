import { memo } from "react";

import { Avatar } from "@components/ui";
import invariant from "tiny-invariant";
import { ProjectsQueries } from "~/projects/data";
import { useProjectUserAccessesDialog } from "~/projects/dialogs/project-user-accesses";

import { BackButton, TabNavigation } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

import { ProjectEnvFilter, ProjectStatusBadge } from "@application/modules/projects/module-shared/components";

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
            label: "Settings",
            activePathPrefixes: [ROUTE.projects.single.configuration.$route(projectId)],
        },
        {
            route: ROUTE.projects.single.providerConfiguration.accessTokens.$route(projectId),
            label: "Provider Settings",
            activePathPrefixes: [ROUTE.projects.single.providerConfiguration.$route(projectId)],
        },
        {
            route: ROUTE.projects.single.sources.githubApps.$route(projectId),
            label: "Sources",
            activePathPrefixes: [ROUTE.projects.single.sources.$route(projectId)],
        },
        {
            route: ROUTE.projects.single.clusterResources.networks.$route(projectId),
            label: "Cluster Resources",
            activePathPrefixes: [ROUTE.projects.single.clusterResources.$route(projectId)],
        },
    ];

    function openProjectUserAccessesDialog() {
        projectUserAccessesDialog.actions.open(project.id, project.name);
    }

    return (
        <div className="bg-background pt-4 px-5 rounded-lg">
            <div className="flex items-center justify-between gap-3">
                <div className="flex justify-between w-full min-w-0 flex-wrap items-center gap-3">
                    <SingleProjectBreadcrumbs project={project} />
                    <ProjectEnvFilter
                        projectId={projectId}
                        envs={project.envs}
                    />
                </div>
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

            <div className="flex items-center gap-4 mt-4 pb-4">
                <BackButton />
                <div className="w-full flex flex-wrap items-end gap-4">
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
                                    <button
                                        key={user.id}
                                        type="button"
                                        aria-label="Manage project access"
                                        title="Manage project access"
                                        className="rounded-full cursor-pointer bg-transparent p-0 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                                        onClick={openProjectUserAccessesDialog}
                                    >
                                        <Avatar
                                            name={user.fullName || user.email || user.username}
                                            src={user.photo}
                                            className="size-8 border-2 border-background"
                                        />
                                    </button>
                                ))}
                                {extraAccessUsers > 0 && (
                                    <button
                                        type="button"
                                        aria-label="Manage project access"
                                        title="Manage project access"
                                        className="flex size-8 cursor-pointer items-center justify-center rounded-full border-2 border-background bg-primary/10 p-0 text-xs font-semibold text-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                                        onClick={openProjectUserAccessesDialog}
                                    >
                                        +{extraAccessUsers}
                                    </button>
                                )}
                            </div>
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

interface ProjectAccessUser {
    id: string;
    fullName: string;
    email: string;
    username: string;
    photo: string | null;
}

export const SingleProjectHeader = memo(View);
