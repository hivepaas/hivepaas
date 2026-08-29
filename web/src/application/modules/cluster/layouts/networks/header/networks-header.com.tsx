import { memo } from "react";

import { moduleHeaderBox } from "@lib/styles";
import { cn } from "@lib/utils";

function View() {
    return (
        <div className={cn(moduleHeaderBox)}>
            <div className="py-0.5 sm:py-1.5">
                <h1 className="text-base sm:text-lg font-bold text-foreground flex items-center gap-1.5 tracking-tight">
                    <span className="text-amber-500/50 dark:text-amber-400/50 font-mono font-normal select-none">
                        /
                    </span>
                    <span>Networks</span>
                    <span className="text-amber-500/50 dark:text-amber-400/50 font-mono font-normal select-none">
                        /
                    </span>
                </h1>
            </div>
        </div>
    );
}

export const NetworksHeader = memo(View);
