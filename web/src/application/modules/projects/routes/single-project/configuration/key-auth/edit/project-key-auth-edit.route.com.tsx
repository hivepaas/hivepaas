import { useParams } from "react-router";
import { getProjectEnvFilterParam, useSelectedProjectEnv } from "~/projects/module-shared/hooks";
import { KeyAuthFormRoute } from "~/settings/module-shared/components/key-auth-form-route";

export function ProjectKeyAuthEditRoute() {
    const { keyAuthId = "", id: projectId = "" } = useParams();
    const selectedEnv = useSelectedProjectEnv(projectId);
    const scopedEnv = getProjectEnvFilterParam(selectedEnv);

    return (
        <KeyAuthFormRoute
            mode="edit"
            scope={{ type: "project", projectId, env: scopedEnv }}
            keyAuthId={keyAuthId}
        />
    );
}
