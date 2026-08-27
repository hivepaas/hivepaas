import { memo } from "react";

import { moduleHeaderBox } from "@lib/styles";
import { cn } from "@lib/utils";

function View() {
    return (
        <div className={cn(moduleHeaderBox)}>
            <div className="py-0.5 sm:py-1.5">
                <h1 className="text-base sm:text-lg font-bold text-foreground">Networks</h1>
            </div>
        </div>
    );
}

export const NetworksHeader = memo(View);
