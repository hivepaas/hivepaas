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
        const currentBottomOffset = bottomOffset ?? (window.innerWidth >= 768 ? 32 : 20);
        const available = window.innerHeight - elementTopInDocument - currentBottomOffset;
        const computed = Math.max(minHeight, Math.floor(available));
        setFullViewHeight(computed);
    }, [enabled, bottomOffset, minHeight]);

    useEffect(() => {
        if (!enabled) {
            setFullViewHeight(null);
            return undefined;
        }

        recalculate();

        let animFrameId: number | null = null;
        let timeoutId: ReturnType<typeof setTimeout> | null = null;

        const scheduleRecalculate = () => {
            if (animFrameId !== null) {
                window.cancelAnimationFrame(animFrameId);
            }
            animFrameId = window.requestAnimationFrame(() => {
                animFrameId = null;
                recalculate();
            });

            if (timeoutId !== null) {
                clearTimeout(timeoutId);
            }
            timeoutId = setTimeout(() => {
                timeoutId = null;
                recalculate();
            }, 60);
        };

        const handleResize = () => {
            scheduleRecalculate();
        };

        window.addEventListener("resize", handleResize);

        const resizeObserver = new ResizeObserver(() => {
            scheduleRecalculate();
        });

        const element = containerRef.current;
        if (element) {
            resizeObserver.observe(element);
            if (element.parentElement) {
                resizeObserver.observe(element.parentElement);
                if (element.parentElement.parentElement) {
                    resizeObserver.observe(element.parentElement.parentElement);
                }
            }
        }

        const mainElement = document.querySelector("main") ?? document.body;
        resizeObserver.observe(mainElement);

        const mutationObserver = new MutationObserver(() => {
            scheduleRecalculate();
        });

        mutationObserver.observe(document.body, {
            childList: true,
            subtree: true,
            attributes: true,
            attributeFilter: ["style", "class", "aria-expanded", "hidden"],
        });

        scheduleRecalculate();

        return () => {
            window.removeEventListener("resize", handleResize);
            resizeObserver.disconnect();
            mutationObserver.disconnect();
            if (animFrameId !== null) {
                window.cancelAnimationFrame(animFrameId);
            }
            if (timeoutId !== null) {
                clearTimeout(timeoutId);
            }
        };
    }, [enabled, recalculate]);

    return { containerRef, fullViewHeight, recalculate };
}
