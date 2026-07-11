import { type ComponentProps, useEffect, useRef, useState } from "react";

import { cn } from "@lib/utils";

export function FormActionBar({
    children,
    className,
    contentClassName,
    sticky = true,
    ...props
}: ComponentProps<"div"> & Props) {
    const sentinelRef = useRef<HTMLDivElement>(null);
    const [isStuck, setIsStuck] = useState(false);

    useEffect(() => {
        if (!sticky) return;
        const sentinel = sentinelRef.current;
        if (!sentinel) return;

        const observer = new IntersectionObserver(([entry]) => {
            if (entry) {
                setIsStuck(!entry.isIntersecting);
            }
        });

        observer.observe(sentinel);
        return () => {
            observer.disconnect();
        };
    }, [sticky]);

    return (
        <>
            <div
                className={cn(
                    sticky
                        ? cn(
                              "sticky bottom-0 z-10 mt-6 shrink-0 pt-4 bg-background border-t transition-[padding-bottom] duration-200 ease-in-out",
                              isStuck ? "pb-4" : "pb-0",
                          )
                        : "mt-6 shrink-0 px-0 pb-6",
                    className,
                )}
                {...props}
            >
                <div className={cn("flex items-center justify-end gap-3", contentClassName)}>{children}</div>
            </div>
            {sticky && (
                <div
                    ref={sentinelRef}
                    className="h-px mb-[-24px]"
                    aria-hidden
                />
            )}
        </>
    );
}

interface Props {
    contentClassName?: string;
    sticky?: boolean;
}
