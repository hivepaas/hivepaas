import { useParams } from "react-router";
import { getProjectEnvFilterParam, useSelectedProjectEnv } from "~/projects/module-shared/hooks";
import { BasicAuthFormRoute } from "~/settings/module-shared/components/basic-auth-form-route";

export function ProjectBasicAuthEditRoute() {
    const { basicAuthId = "", id: projectId = "" } = useParams();
    const selectedEnv = useSelectedProjectEnv(projectId);
    const scopedEnv = getProjectEnvFilterParam(selectedEnv);

    return (
        <BasicAuthFormRoute
            mode="edit"
            scope={{ type: "project", projectId, env: scopedEnv }}
            basicAuthId={basicAuthId}
        />
    );
}
