import { memo } from "react";

import { Skeleton } from "@components/ui";

function View() {
    return (
        <div className="bg-background pt-4 px-5 rounded-lg">
            <div className="py-3 border-b border-border">
                <Skeleton className="w-full h-8" />
            </div>
        </div>
    );
}
export const SingleUserHeaderSkeleton = memo(View);
