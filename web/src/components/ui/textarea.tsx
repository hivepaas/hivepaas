import * as React from "react";

import { cn } from "@/lib/utils";

type TextareaProps = React.ComponentPropsWithoutRef<"textarea"> & {
    minRows?: number;
    maxRows?: number;
};

function getLineHeight(element: HTMLTextAreaElement) {
    const computedStyle = window.getComputedStyle(element);
    const lineHeight = Number.parseFloat(computedStyle.lineHeight);

    if (Number.isFinite(lineHeight)) {
        return lineHeight;
    }

    const fontSize = Number.parseFloat(computedStyle.fontSize);

    return Number.isFinite(fontSize) ? fontSize * 1.2 : 20;
}

const Textarea = React.forwardRef<HTMLTextAreaElement, TextareaProps>(
    ({ className, minRows = 6, maxRows = 10, onPointerDown, style, ...props }, ref) => {
        const textareaRef = React.useRef<HTMLTextAreaElement | null>(null);
        const [minRowsHeight, setMinRowsHeight] = React.useState<number>();
        const [maxRowsHeight, setMaxRowsHeight] = React.useState<number>();
        const [hasManualResize, setHasManualResize] = React.useState(false);

        const shouldLimitHeight = typeof maxRows === "number" && maxRows > 0;

        React.useImperativeHandle(ref, () => textareaRef.current!);

        React.useLayoutEffect(() => {
            // Reset manual resize whenever row constraints change so new limits take effect.
            setHasManualResize(false);

            const textarea = textareaRef.current;

            if (!textarea) {
                setMinRowsHeight(undefined);
                setMaxRowsHeight(undefined);
                return;
            }

            const computeHeights = () => {
                const computedStyle = window.getComputedStyle(textarea);
                const borderTop = Number.parseFloat(computedStyle.borderTopWidth) || 0;
                const borderBottom = Number.parseFloat(computedStyle.borderBottomWidth) || 0;
                const paddingTop = Number.parseFloat(computedStyle.paddingTop) || 0;
                const paddingBottom = Number.parseFloat(computedStyle.paddingBottom) || 0;
                const chrome = paddingTop + paddingBottom + borderTop + borderBottom;
                const lh = getLineHeight(textarea);

                setMinRowsHeight(prev => {
                    const next = typeof minRows === "number" && minRows > 0 ? lh * minRows + chrome : undefined;
                    return prev === next ? prev : next;
                });

                setMaxRowsHeight(prev => {
                    const next = typeof maxRows === "number" && maxRows > 0 ? lh * maxRows + chrome : undefined;
                    return prev === next ? prev : next;
                });
            };

            computeHeights();

            window.addEventListener("resize", computeHeights);

            return () => {
                window.removeEventListener("resize", computeHeights);
            };
        }, [minRows, maxRows]);

        const handlePointerDown = (event: React.PointerEvent<HTMLTextAreaElement>) => {
            onPointerDown?.(event);

            const textarea = textareaRef.current;

            if (!textarea || !shouldLimitHeight || event.defaultPrevented) {
                return;
            }

            const resizeHandleSize = 16;
            const rect = textarea.getBoundingClientRect();
            const isResizeHandlePointer =
                event.clientX >= rect.right - resizeHandleSize && event.clientY >= rect.bottom - resizeHandleSize;

            if (isResizeHandlePointer) {
                setHasManualResize(true);
            }
        };

        return (
            <textarea
                ref={textareaRef}
                data-slot="textarea"
                className={cn(
                    "border-input placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive dark:bg-input/30 flex field-sizing-content w-full rounded-md border bg-transparent px-3 py-2 text-base shadow-xs transition-[color,box-shadow] outline-none focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50 md:text-sm",
                    className,
                )}
                style={{
                    ...style,
                    minHeight: minRowsHeight ?? style?.minHeight,
                    maxHeight: shouldLimitHeight && !hasManualResize ? maxRowsHeight : style?.maxHeight,
                    overflowY: shouldLimitHeight && !hasManualResize ? "auto" : style?.overflowY,
                }}
                onPointerDown={handlePointerDown}
                {...props}
            />
        );
    },
);

Textarea.displayName = "Textarea";

export { Textarea };
export type { TextareaProps };
