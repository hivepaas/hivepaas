import { type PropsWithChildren } from "react";

import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { Page404NotFound, PageError } from "@application/shared/pages";

import { ProjectsQueries } from "@application/modules/projects/data";

import { isHttp404Exception } from "@infrastructure/api";

import { SingleProjectHeader } from "../header";

export function SingleProjectLayout({ children }: PropsWithChildren) {
    const { id } = useParams<{ id: string }>();

    invariant(id, "Project id must be defined");

    const { error, refetch } = ProjectsQueries.useFindOneById({ projectID: id });

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

    return (
        <div className="flex flex-col gap-5">
            <SingleProjectHeader projectId={id} />
            {children}
        </div>
    );
}
