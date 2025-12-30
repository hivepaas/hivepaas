import { memo } from "react";

import { moduleHeaderBox } from "@lib/styles";
import { cn } from "@lib/utils";

function View() {
    return (
        <div className={cn(moduleHeaderBox)}>
            <div className="py-2">
                <h1 className="text-lg font-bold">Nodes</h1>
            </div>
        </div>
    );
}

export const NodesHeader = memo(View);
