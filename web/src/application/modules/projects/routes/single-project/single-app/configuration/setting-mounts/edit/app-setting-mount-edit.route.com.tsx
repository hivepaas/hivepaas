import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { AppSettingMountFormRoute } from "../form-route";

export function AppSettingMountEditRoute() {
    const {
        id: projectId,
        env,
        appId,
        settingMountId,
    } = useParams<{
        id: string;
        env: string;
        appId: string;
        settingMountId: string;
    }>();

    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");
    invariant(appId, "appId must be defined");
    invariant(settingMountId, "settingMountId must be defined");

    return (
        <AppSettingMountFormRoute
            mode="edit"
            projectId={projectId}
            env={env}
            appId={appId}
            settingMountId={settingMountId}
        />
    );
}
