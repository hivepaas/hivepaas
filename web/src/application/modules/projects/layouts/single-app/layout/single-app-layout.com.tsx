import { type PropsWithChildren } from "react";

import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { useProfileContext } from "@application/shared/context";
import { Page404NotFound, PageError, PageNoAccess } from "@application/shared/pages";
import { useConditionalProject } from "@application/shared/permissions";

import { ProjectAppsQueries, ProjectsQueries } from "@application/modules/projects/data";

import { isHttp404Exception } from "@infrastructure/api";

import { SingleAppHeader } from "../header";

export function SingleAppLayout({ children }: PropsWithChildren) {
    const { id, appId } = useParams<{ id: string; appId: string }>();

    invariant(id, "Project id must be defined");
    invariant(appId, "App id must be defined");

    const { profile } = useProfileContext();
    const { canRead } = useConditionalProject({ projectId: id });
    const { error, refetch } = ProjectAppsQueries.useFindOneById({ projectID: id, appID: appId });
    const {
        data: projectData,
        error: projectError,
        refetch: refetchProject,
    } = ProjectsQueries.useFindOneById({
        projectID: id,
    });

    const project = projectData?.data;
    const hasProjectReadAccess =
        project && profile
            ? project.owner.id === profile.id ||
              project.userAccesses.some(user => user.id === profile.id && user.access.read)
            : false;

    if (error || projectError) {
        const pageError = error ?? projectError;
        invariant(pageError, "pageError must be defined");

        if (isHttp404Exception(pageError)) {
            return <Page404NotFound withBackButton={false} />;
        }

        return (
            <PageError
                error={pageError}
                onRetry={() => {
                    void refetch();
                    void refetchProject();
                }}
            />
        );
    }

    if (project && !canRead && !hasProjectReadAccess) {
        return <PageNoAccess />;
    }

    return (
        <div className="flex flex-col gap-5">
            <SingleAppHeader
                projectId={id}
                appId={appId}
            />
            {children}
        </div>
    );
}
