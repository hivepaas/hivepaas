import { type PropsWithChildren } from "react";

import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { Page404NotFound, PageError } from "@application/shared/pages";

import { UsersQueries } from "@application/modules/user-management/data/queries";

import { isHttp404Exception } from "@infrastructure/api";

import { SingleUserHeader } from "../header/single-user-header.com";

export function SingleUserLayout({ children }: PropsWithChildren) {
    const { id: userId } = useParams<{ id: string }>();

    invariant(userId, "userId must be defined");
    const { error, refetch } = UsersQueries.useFindOneById({ id: userId });

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
            <SingleUserHeader userId={userId} />
            {children}
        </div>
    );
}
