import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";

import { NetworksTable } from "../building-blocks";

function View() {
    return (
        <div className={cn(listBox)}>
            <NetworksTable />
        </div>
    );
}

export function NetworksRoute() {
    return <View />;
}
