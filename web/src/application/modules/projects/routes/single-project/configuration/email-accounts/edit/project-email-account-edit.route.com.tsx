import { useParams } from "react-router";
import { getProjectEnvFilterParam, useSelectedProjectEnv } from "~/projects/module-shared/hooks";
import { EmailAccountFormRoute } from "~/settings/module-shared/components/email-account-form-route";

export function ProjectEmailAccountEditRoute() {
    const { emailAccountId = "", id: projectId = "" } = useParams();
    const selectedEnv = useSelectedProjectEnv(projectId);
    const scopedEnv = getProjectEnvFilterParam(selectedEnv);

    return (
        <EmailAccountFormRoute
            mode="edit"
            scope={{ type: "project", projectId, env: scopedEnv }}
            emailAccountId={emailAccountId}
        />
    );
}
