import { useParams } from "react-router";
import { BasicAuthFormRoute } from "~/settings/module-shared/components/basic-auth-form-route";

export function ProjectBasicAuthEditRoute() {
    const { basicAuthId = "", id: projectId = "" } = useParams();

    return (
        <BasicAuthFormRoute
            mode="edit"
            scope={{ type: "project", projectId }}
            basicAuthId={basicAuthId}
        />
    );
}
