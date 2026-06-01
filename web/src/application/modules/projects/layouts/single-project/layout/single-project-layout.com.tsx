import { type PropsWithChildren } from "react";

import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { useProfileContext } from "@application/shared/context";
import { Page404NotFound, PageError, PageNoAccess } from "@application/shared/pages";
import { useConditionalProject } from "@application/shared/permissions";

import { ProjectsQueries } from "@application/modules/projects/data";

import { isHttp404Exception } from "@infrastructure/api";

import { SingleProjectHeader } from "../header";

export function SingleProjectLayout({ children }: PropsWithChildren) {
    const { id } = useParams<{ id: string }>();

    invariant(id, "Project id must be defined");

    const { profile } = useProfileContext();
    const { canRead } = useConditionalProject({ projectId: id });
    const { data, error, refetch } = ProjectsQueries.useFindOneById({ projectID: id });

    const project = data?.data;
    const hasProjectReadAccess =
        project && profile
            ? project.owner.id === profile.id ||
              project.userAccesses.some(user => user.id === profile.id && user.access.read)
            : false;

    if (error) {
        if (isHttp404Exception(error)) {
            return <Page404NotFound withBackButton={false} />;
        }

        return (
            <PageError
                error={error}
                onRetry={refetch}
            />
        );
    }

    if (project && !canRead && !hasProjectReadAccess) {
        return <PageNoAccess />;
    }

    return (
        <div className="flex flex-col gap-5">
            <SingleProjectHeader projectId={id} />
            {children}
        </div>
    );
}
