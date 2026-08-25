import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { ProjectBasicAuthTable } from "~/settings/module-shared/components";

import { ProjectProviderSettingsScopeHeader } from "@application/modules/projects/module-shared/components";

export function ProjectBasicAuthRoute() {
    const { id: projectId } = useParams<{ id: string }>();

    invariant(projectId, "projectId must be defined");

    return (
        <div className="flex flex-col gap-4">
            <ProjectProviderSettingsScopeHeader projectId={projectId} />
            <ProjectBasicAuthTable projectId={projectId} />
        </div>
    );
}
