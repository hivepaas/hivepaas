import { type PropsWithChildren } from "react";

import { AppLoader } from "@application/shared/components";
import { useProfileContext } from "@application/shared/context";
import { SessionQueries } from "@application/shared/data/queries";
import { PageError } from "@application/shared/pages";

import { isSessionInvalidException, session } from "@infrastructure/api";

export function ApplicationProfileInit({ children }: PropsWithChildren) {
    const { profile, setProfile, clearProfile } = useProfileContext();

    const { isLoading, error, refetch } = SessionQueries.useGetProfile({
        onSuccess: ({ data }) => {
            setProfile(data);
        },
        onSessionInvalid: clearProfile,
        enabled: session.hasToken() && !profile,
    });

    if (isLoading) {
        return <AppLoader />;
    }

    /**
     * If the error is not a session invalid exception, show the error page.
     */
    if (error && !isSessionInvalidException(error)) {
        return (
            <PageError
                error={error}
                onRetry={refetch}
            />
        );
    }

    return children;
}
