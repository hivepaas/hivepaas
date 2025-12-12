import { NodesTable } from "../building-blocks";

function View() {
    return (
        <div className="bg-background rounded-lg p-4">
            <NodesTable />
        </div>
    );
}

export function NodesRoute() {
    return <View />;
}
