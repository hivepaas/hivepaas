import { memo } from "react";

import { Skeleton } from "@components/ui";

import { Separator } from "@/components/ui/separator";

function View() {
    return (
        <div className="bg-background pt-4 px-5 rounded-lg flex flex-col gap-3 pb-3">
            <Skeleton className="w-full h-8" />
            <Separator className="opacity-50" />
            <Skeleton className="w-full h-12" />
            <Separator className="opacity-50" />
            <Skeleton className="w-full h-8" />
        </div>
    );
}
export const SingleProjectHeaderSkeleton = memo(View);
