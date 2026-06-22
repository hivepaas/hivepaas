import { useParams } from "react-router";
import { BasicAuthFormRoute } from "~/settings/module-shared/components/basic-auth-form-route";

export function ProjectBasicAuthCreateRoute() {
    const { id: projectId = "" } = useParams();

    return (
        <BasicAuthFormRoute
            mode="create"
            scope={{ type: "project", projectId }}
        />
    );
}
