import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { ProjectAppsTable } from "../building-blocks";

export function ProjectAppsRoute() {
    const { id: projectId } = useParams<{ id: string }>();
    invariant(projectId, "projectId must be defined");

    return (
        <div className={cn(listBox)}>
            <ProjectAppsTable projectId={projectId} />
        </div>
    );
}
