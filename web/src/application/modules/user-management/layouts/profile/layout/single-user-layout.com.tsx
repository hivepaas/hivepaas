import { type PropsWithChildren } from "react";

import invariant from "tiny-invariant";

import { useProfileContext } from "@application/shared/context";
import { Page404NotFound, PageError } from "@application/shared/pages";

import { UsersQueries } from "@application/modules/user-management/data/queries";

import { isHttp404Exception } from "@infrastructure/api";

import { ProfileHeader } from "../header";

export function ProfileLayout({ children }: PropsWithChildren) {
    const { profile } = useProfileContext();

    invariant(profile, "Profile must be defined");
    const { error, refetch } = UsersQueries.useFindOneById({ id: profile.id });

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
            <ProfileHeader userId={profile.id} />
            {children}
        </div>
    );
}
