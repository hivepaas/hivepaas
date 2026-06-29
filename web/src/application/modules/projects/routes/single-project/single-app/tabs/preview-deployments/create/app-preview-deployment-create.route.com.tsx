import { useLocation, useParams } from "react-router";
import invariant from "tiny-invariant";
import type { AppPreviews_PrepareCreate_Res } from "~/projects/api/services";

import { AppPreviewDeploymentFormRoute } from "../form-route";

type PreparedPreview = AppPreviews_PrepareCreate_Res["data"];

function getPreparedPreviewFromState(state: unknown): PreparedPreview | undefined {
    if (typeof state !== "object" || state === null || !("preparedPreview" in state)) {
        return undefined;
    }

    const { preparedPreview } = state as { preparedPreview?: unknown };

    if (typeof preparedPreview !== "object" || preparedPreview === null) {
        return undefined;
    }

    const value = preparedPreview as Partial<PreparedPreview>;

    if (typeof value.repoURL !== "string") {
        return undefined;
    }

    return value as PreparedPreview;
}

export function AppPreviewDeploymentCreateRoute() {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();
    const location = useLocation();

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    return (
        <AppPreviewDeploymentFormRoute
            projectId={projectId}
            appId={appId}
            initialPreparedPreview={getPreparedPreviewFromState(location.state)}
        />
    );
}
