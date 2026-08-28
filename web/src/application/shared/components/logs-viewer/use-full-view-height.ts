import { type RefObject, useCallback, useEffect, useRef, useState } from "react";

export interface UseFullViewHeightOptions {
    enabled?: boolean;
    bottomOffset?: number;
    minHeight?: number;
}

export interface UseFullViewHeightReturn {
    containerRef: RefObject<HTMLDivElement | null>;
    fullViewHeight: number | null;
    recalculate: () => void;
}

export function useFullViewHeight({
    enabled = true,
    bottomOffset,
    minHeight = 250,
}: UseFullViewHeightOptions = {}): UseFullViewHeightReturn {
    const containerRef = useRef<HTMLDivElement | null>(null);
    const [fullViewHeight, setFullViewHeight] = useState<number | null>(null);

    const recalculate = useCallback(() => {
        if (!enabled) {
            setFullViewHeight(null);
            return;
        }

        const element = containerRef.current;
        if (!element) {
            return;
        }

        const rect = element.getBoundingClientRect();
        // If window is scrolled, account for scroll to measure base layout top
        const scrollTop = window.scrollY || document.documentElement.scrollTop || 0;
        const elementTopInDocument = rect.top + scrollTop;
        const currentBottomOffset = bottomOffset ?? (window.innerWidth >= 640 ? 32 : 20);
        const available = window.innerHeight - elementTopInDocument - currentBottomOffset;
        const computed = Math.max(minHeight, Math.floor(available));
        setFullViewHeight(computed);
    }, [enabled, bottomOffset, minHeight]);

    useEffect(() => {
        if (!enabled) {
            setFullViewHeight(null);
            return;
        }

        recalculate();

        const handleResize = () => {
            recalculate();
        };

        window.addEventListener("resize", handleResize);

        const observer = new ResizeObserver(() => {
            recalculate();
        });

        observer.observe(document.body);

        const timer = setTimeout(recalculate, 60);

        return () => {
            window.removeEventListener("resize", handleResize);
            observer.disconnect();
            clearTimeout(timer);
        };
    }, [enabled, recalculate]);

    return { containerRef, fullViewHeight, recalculate };
}
