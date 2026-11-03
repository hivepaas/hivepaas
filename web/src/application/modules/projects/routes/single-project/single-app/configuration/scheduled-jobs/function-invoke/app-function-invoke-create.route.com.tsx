import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { AppFunctionInvokeFormRoute } from "./app-function-invoke-form-route.com";

export function AppFunctionInvokeCreateRoute() {
    const { id: projectId, env, appId } = useParams<{ id: string; env: string; appId: string }>();

    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");
    invariant(appId, "appId must be defined");

    return (
        <AppFunctionInvokeFormRoute
            mode="create"
            projectId={projectId}
            env={env}
            appId={appId}
        />
    );
}
