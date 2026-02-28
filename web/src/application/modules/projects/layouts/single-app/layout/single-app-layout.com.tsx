import { type PropsWithChildren } from "react";

import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { Page404NotFound, PageError } from "@application/shared/pages";

import { ProjectAppsQueries } from "@application/modules/projects/data";

import { isHttp404Exception } from "@infrastructure/api";

import { SingleAppHeader } from "../header";

export function SingleAppLayout({ children }: PropsWithChildren) {
    const { id, appId } = useParams<{ id: string; appId: string }>();

    invariant(id, "Project id must be defined");
    invariant(appId, "App id must be defined");

    const { error, refetch } = ProjectAppsQueries.useFindOneById({ projectID: id, appID: appId });

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
            <SingleAppHeader
                projectId={id}
                appId={appId}
            />
            {children}
        </div>
    );
}
