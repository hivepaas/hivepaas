import { memo } from "react";

import { Avatar } from "@components/ui";
import invariant from "tiny-invariant";
import { ProjectsQueries } from "~/projects/data";
import { useProjectUserAccessesDialog } from "~/projects/dialogs/project-user-accesses";

import { TabNavigation } from "@application/shared/components";
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
            label: "Providers & Keys",
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
        projectUserAccessesDialog.actions.open(project.id, project.name, project.envs);
    }

    return (
        <div className="bg-background pt-4 px-4 sm:px-5 rounded-lg">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3">
                <SingleProjectBreadcrumbs project={project} />
                <div className="flex items-center">
                    <ProjectEnvFilter
                        projectId={projectId}
                        envs={project.envs}
                    />
                </div>
            </div>

            <div className="flex items-center gap-3 sm:gap-4 py-3 border-t border-border/40">
                <Avatar
                    name={project.name}
                    src={project.photo}
                    className="size-12 sm:size-14 md:size-16 text-lg sm:text-xl md:text-2xl shrink-0 rounded-xl"
                />
                <div className="flex min-w-0 flex-1 flex-col gap-1.5 justify-center">
                    <div className="flex flex-wrap items-center gap-2">
                        <h2 className="text-lg sm:text-xl font-semibold text-foreground truncate">{project.name}</h2>
                        <ProjectStatusBadge status={project.status} />
                    </div>

                    <div className="flex flex-wrap items-center gap-x-4 gap-y-1.5 text-xs sm:text-sm text-muted-foreground">
                        <div className="flex items-center gap-1.5 min-w-0">
                            <span>Owner:</span>
                            <Avatar
                                name={project.owner.fullName}
                                src={project.owner.photo}
                                className="size-6 text-xs shrink-0"
                                borderless
                            />
                            <span className="truncate font-medium text-foreground max-w-[130px] sm:max-w-[200px]">
                                {project.owner.fullName}
                            </span>
                        </div>

                        <span className="hidden sm:inline text-muted-foreground/30">•</span>

                        <div className="flex items-center gap-1.5">
                            <span>Access:</span>
                            <div className="flex -space-x-1.5 items-center">
                                {visibleAccessUsers.map(user => (
                                    <button
                                        key={user.id}
                                        type="button"
                                        aria-label="Configure project access"
                                        title="Configure project access"
                                        className="rounded-full cursor-pointer bg-transparent p-0 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                                        onClick={openProjectUserAccessesDialog}
                                    >
                                        <Avatar
                                            name={user.fullName || user.email || user.username}
                                            src={user.photo}
                                            className="size-6 text-[10px] border-2 border-background"
                                        />
                                    </button>
                                ))}
                                {extraAccessUsers > 0 && (
                                    <button
                                        type="button"
                                        aria-label="Configure project access"
                                        title="Configure project access"
                                        className="flex size-6 cursor-pointer items-center justify-center rounded-full border-2 border-background bg-primary/10 p-0 text-[10px] font-semibold text-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
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

            <div className="overflow-x-auto">
                <TabNavigation links={links} />
            </div>
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
