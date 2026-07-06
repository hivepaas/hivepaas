import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";

import { VolumesTable } from "../building-blocks";

function View() {
    return (
        <div className={cn(listBox)}>
            <VolumesTable />
        </div>
    );
}

export function VolumesRoute() {
    return <View />;
}
