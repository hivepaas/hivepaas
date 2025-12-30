import { ProjectsTable } from "../building-blocks";

function View() {
    return (
        <div className="bg-background rounded-lg p-4">
            <ProjectsTable />
        </div>
    );
}

export function ProjectsRoute() {
    return <View />;
}
