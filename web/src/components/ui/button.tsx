import * as React from "react";

import { cn } from "@/lib/utils";
import { Slot } from "@radix-ui/react-slot";
import { type VariantProps, cva } from "class-variance-authority";
import { Loader2 } from "lucide-react";

const buttonVariants = cva(
    "inline-flex items-center cursor-pointer justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-all disabled:pointer-events-none disabled:opacity-50 disabled:shadow-none [&_svg]:pointer-events-none [&_svg:not([class*='size-'])]:size-4 shrink-0 [&_svg]:shrink-0 outline-none focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive",
    {
        variants: {
            variant: {
                default:
                    "bg-gradient-to-r from-[#4e46b4] via-[#5c54c7] to-[#7066e0] text-white border border-white/20 hover:brightness-105 active:scale-[0.98] transition-all duration-200",
                destructive:
                    "bg-gradient-to-r from-red-600 via-rose-600 to-red-500 text-white border border-white/20 hover:brightness-105 active:scale-[0.98] transition-all duration-200 focus-visible:ring-destructive/20 dark:focus-visible:ring-destructive/40",
                destructiveOutline:
                    "border border-destructive text-destructive hover:bg-destructive/10 active:scale-[0.98] focus-visible:ring-destructive/20 dark:focus-visible:ring-destructive/40 dark:border-destructive/50 transition-all duration-200",
                outline:
                    "border bg-background hover:bg-accent hover:text-accent-foreground active:scale-[0.98] dark:bg-input/30 dark:border-input dark:hover:bg-input/50 transition-all duration-200",
                secondary:
                    "bg-secondary text-secondary-foreground hover:bg-secondary/80 active:scale-[0.98] transition-all duration-200",
                ghost: "hover:bg-accent hover:text-accent-foreground dark:hover:bg-accent/50",
                link: "text-primary underline-offset-4 hover:underline",
                tactile:
                    "bg-gradient-to-b from-primary/90 via-primary to-primary text-primary-foreground border border-primary/90 dark:border-white/20 shadow-[inset_0_1px_0_0_rgba(255,255,255,0.25),0_1px_2px_0_rgba(0,0,0,0.1)] hover:from-primary hover:to-primary/90 hover:shadow-[inset_0_1px_0_0_rgba(255,255,255,0.35),0_2px_4px_0_rgba(0,0,0,0.15)] active:scale-[0.98] active:translate-y-px active:shadow-[inset_0_1px_2px_0_rgba(0,0,0,0.2)] dark:shadow-[inset_0_1px_0_0_rgba(255,255,255,0.5),0_1px_2px_0_rgba(0,0,0,0.3)]",
                glow: "bg-gradient-to-r from-[#4e46b4] via-[#5c54c7] to-[#7066e0] text-white border border-white/20 shadow-md shadow-indigo-500/25 hover:shadow-lg hover:shadow-indigo-500/40 hover:brightness-105 active:translate-y-0 active:scale-[0.98] active:shadow-sm transition-all duration-200",
            },
            size: {
                "default": "h-9 px-4 py-2 has-[>svg]:px-3",
                "sm": "h-8 rounded-md gap-1.5 px-3 has-[>svg]:px-2.5",
                "lg": "h-10 rounded-md px-6 has-[>svg]:px-4",
                "icon": "size-9",
                "icon-sm": "size-8",
                "icon-lg": "size-10",
            },
        },
        defaultVariants: {
            variant: "default",
            size: "default",
        },
    },
);

function Button({
    className,
    variant,
    size,
    asChild = false,
    isLoading = false,
    disabled,
    children,
    ...props
}: React.ComponentProps<"button"> &
    VariantProps<typeof buttonVariants> & {
        asChild?: boolean;
        isLoading?: boolean;
    }) {
    const classNames = cn(buttonVariants({ variant, size, className }));

    if (asChild) {
        return (
            <Slot
                data-slot="button"
                data-disabled={isLoading || disabled || undefined}
                className={classNames}
                aria-busy={isLoading || undefined}
                aria-disabled={isLoading || disabled || undefined}
                {...props}
            >
                {children}
            </Slot>
        );
    }

    return (
        <button
            data-slot="button"
            className={classNames}
            disabled={isLoading || disabled}
            aria-busy={isLoading || undefined}
            {...props}
        >
            {isLoading && <Loader2 className="animate-spin" />}
            {children}
        </button>
    );
}

export { Button, buttonVariants };
