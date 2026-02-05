import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";

import { NodesTable } from "../building-blocks";

function View() {
    return (
        <div className={cn(listBox)}>
            <NodesTable />
        </div>
    );
}

export function NodesRoute() {
    return <View />;
}
