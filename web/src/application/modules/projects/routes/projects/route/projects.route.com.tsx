import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";

import { ProjectsTable } from "../building-blocks";

function View() {
    return (
        <div className={cn(listBox)}>
            <ProjectsTable />
        </div>
    );
}

export function ProjectsRoute() {
    return <View />;
}
