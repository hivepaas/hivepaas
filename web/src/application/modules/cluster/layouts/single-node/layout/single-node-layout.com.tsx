import { type PropsWithChildren } from "react";

import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { Page404NotFound, PageError } from "@application/shared/pages";

import { NodesQueries } from "@application/modules/cluster/data";

import { isHttp404Exception } from "@infrastructure/api";

import { SingleNodeHeader } from "../header";

export function SingleNodeLayout({ children }: PropsWithChildren) {
    const { id } = useParams<{ id: string }>();

    invariant(id, "Node id must be defined");

    const { error, refetch } = NodesQueries.useFindOneById({ id });

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
            <SingleNodeHeader nodeId={id} />
            {children}
        </div>
    );
}
