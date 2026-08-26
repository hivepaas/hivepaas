import { useParams } from "react-router";
import { getProjectEnvFilterParam, useSelectedProjectEnv } from "~/projects/module-shared/hooks";
import { AccessTokenFormRoute } from "~/settings/module-shared/components/access-token-form-route";

export function ProjectAccessTokenCreateRoute() {
    const { id: projectId = "" } = useParams();
    const selectedEnv = useSelectedProjectEnv(projectId);
    const scopedEnv = getProjectEnvFilterParam(selectedEnv);

    return (
        <AccessTokenFormRoute
            mode="create"
            scope={{ type: "project", projectId, env: scopedEnv }}
        />
    );
}
