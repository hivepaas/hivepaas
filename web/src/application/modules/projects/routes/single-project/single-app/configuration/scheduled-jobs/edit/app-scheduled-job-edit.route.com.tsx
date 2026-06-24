import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { AppScheduledJobFormRoute } from "../form-route";

export function AppScheduledJobEditRoute() {
    const {
        id: projectId,
        appId,
        scheduledJobId,
    } = useParams<{
        id: string;
        appId: string;
        scheduledJobId: string;
    }>();

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");
    invariant(scheduledJobId, "scheduledJobId must be defined");

    return (
        <AppScheduledJobFormRoute
            mode="edit"
            projectId={projectId}
            appId={appId}
            scheduledJobId={scheduledJobId}
        />
    );
}
