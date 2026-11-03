import { useParams } from "react-router";
import { getProjectEnvFilterParam, useSelectedProjectEnv } from "~/projects/module-shared/hooks";
import { ImPlatformFormRoute } from "~/settings/module-shared/components/im-platform-form-route";

export function ProjectImPlatformEditRoute() {
    const { id: projectId = "", imPlatformId = "" } = useParams();
    const selectedEnv = useSelectedProjectEnv(projectId);
    const scopedEnv = getProjectEnvFilterParam(selectedEnv);

    return (
        <ImPlatformFormRoute
            mode="edit"
            scope={{ type: "project", projectId, env: scopedEnv }}
            imPlatformId={imPlatformId}
        />
    );
}
