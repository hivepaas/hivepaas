import * as React from "react";

import { cn } from "@/lib/utils";
import { Slot } from "@radix-ui/react-slot";
import { type VariantProps, cva } from "class-variance-authority";

const badgeVariants = cva(
    "inline-flex items-center justify-center rounded-md border px-2.5 py-1 text-xs font-medium w-fit whitespace-nowrap shrink-0 [&>svg]:size-3 gap-1 [&>svg]:pointer-events-none focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive transition-[color,box-shadow] overflow-hidden",
    {
        variants: {
            variant: {
                default: "border-transparent bg-primary text-primary-foreground [a&]:hover:bg-primary/90",
                secondary: "border-transparent bg-secondary text-secondary-foreground [a&]:hover:bg-secondary/90",
                destructive:
                    "border-transparent bg-destructive text-white [a&]:hover:bg-destructive/90 focus-visible:ring-destructive/20 dark:focus-visible:ring-destructive/40 dark:bg-destructive/60",
                outline: "text-foreground [a&]:hover:bg-accent [a&]:hover:text-accent-foreground",
            },
        },
        defaultVariants: {
            variant: "default",
        },
    },
);

/**
 * The colors a badge can take. A badge with a tone is one of two kinds:
 *
 * - `soft` (the default): a tinted fill, a border and text in the tone. For what
 *   classifies a row - a kind, a provider, a tag - and asks nothing of the reader.
 * - `solid`: a filled badge with white text. For a status, the one thing in a row
 *   the reader should see first.
 *
 * Give a badge a tone rather than color classes of its own, so that the same
 * meaning has the same color everywhere and both themes are covered.
 */
export type BadgeTone =
    | "neutral"
    | "red"
    | "orange"
    | "amber"
    | "yellow"
    | "lime"
    | "green"
    | "emerald"
    | "teal"
    | "cyan"
    | "sky"
    | "blue"
    | "indigo"
    | "violet"
    | "purple"
    | "fuchsia"
    | "pink"
    | "rose";

export type BadgeAppearance = "soft" | "solid";

const SOFT_TONE_CLASS_NAMES: Record<BadgeTone, string> = {
    neutral: "border-slate-500/30 bg-slate-500/10 text-slate-700 dark:text-slate-400",
    red: "border-red-500/30 bg-red-500/10 text-red-700 dark:text-red-400",
    orange: "border-orange-500/30 bg-orange-500/10 text-orange-700 dark:text-orange-400",
    amber: "border-amber-500/30 bg-amber-500/10 text-amber-700 dark:text-amber-400",
    yellow: "border-yellow-500/30 bg-yellow-500/10 text-yellow-700 dark:text-yellow-400",
    lime: "border-lime-500/30 bg-lime-500/10 text-lime-700 dark:text-lime-400",
    green: "border-green-500/30 bg-green-500/10 text-green-700 dark:text-green-400",
    emerald: "border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-400",
    teal: "border-teal-500/30 bg-teal-500/10 text-teal-700 dark:text-teal-400",
    cyan: "border-cyan-500/30 bg-cyan-500/10 text-cyan-700 dark:text-cyan-400",
    sky: "border-sky-500/30 bg-sky-500/10 text-sky-700 dark:text-sky-400",
    blue: "border-blue-500/30 bg-blue-500/10 text-blue-700 dark:text-blue-400",
    indigo: "border-indigo-500/30 bg-indigo-500/10 text-indigo-700 dark:text-indigo-400",
    violet: "border-violet-500/30 bg-violet-500/10 text-violet-700 dark:text-violet-400",
    purple: "border-purple-500/30 bg-purple-500/10 text-purple-700 dark:text-purple-400",
    fuchsia: "border-fuchsia-500/30 bg-fuchsia-500/10 text-fuchsia-700 dark:text-fuchsia-400",
    pink: "border-pink-500/30 bg-pink-500/10 text-pink-700 dark:text-pink-400",
    rose: "border-rose-500/30 bg-rose-500/10 text-rose-700 dark:text-rose-400",
};

const SOLID_TONE_CLASS_NAMES: Record<BadgeTone, string> = {
    neutral: "border-transparent bg-slate-500 text-white",
    red: "border-transparent bg-red-600 text-white",
    orange: "border-transparent bg-orange-500 text-white",
    amber: "border-transparent bg-amber-500 text-white",
    yellow: "border-transparent bg-yellow-500 text-white",
    lime: "border-transparent bg-lime-500 text-white",
    green: "border-transparent bg-green-500 text-white",
    emerald: "border-transparent bg-emerald-500 text-white",
    teal: "border-transparent bg-teal-500 text-white",
    cyan: "border-transparent bg-cyan-500 text-white",
    sky: "border-transparent bg-sky-500 text-white",
    blue: "border-transparent bg-blue-500 text-white",
    indigo: "border-transparent bg-indigo-500 text-white",
    violet: "border-transparent bg-violet-500 text-white",
    purple: "border-transparent bg-purple-500 text-white",
    fuchsia: "border-transparent bg-fuchsia-500 text-white",
    pink: "border-transparent bg-pink-500 text-white",
    rose: "border-transparent bg-rose-500 text-white",
};

function toneClassName(tone: BadgeTone, appearance: BadgeAppearance): string {
    return appearance === "solid" ? SOLID_TONE_CLASS_NAMES[tone] : SOFT_TONE_CLASS_NAMES[tone];
}

function Badge({
    className,
    variant,
    tone,
    appearance = "soft",
    asChild = false,
    ...props
}: React.ComponentProps<"span"> &
    VariantProps<typeof badgeVariants> & { asChild?: boolean; tone?: BadgeTone; appearance?: BadgeAppearance }) {
    const Comp = asChild ? Slot : "span";

    return (
        <Comp
            data-slot="badge"
            className={cn(
                // A tone brings its own fill, border and text: it sits on the outline shape.
                badgeVariants({ variant: tone ? "outline" : variant }),
                tone && toneClassName(tone, appearance),
                className,
            )}
            {...props}
        />
    );
}

/**
 * What a status means, whatever it is the status of. A status badge takes its
 * color from its meaning, so that "failed" is the same red on a task, a
 * deployment and a node.
 *
 * - `success`: working, done, allowed.
 * - `failure`: failed, down, disabled, missing, expired.
 * - `progress`: happening now.
 * - `waiting`: queued, not started.
 * - `attention`: works, or will, but wants a look: pending, locked, degraded, skipped.
 * - `inactive`: stopped on purpose or no longer there: canceled, paused, shut down.
 * - `info`: neither good nor bad.
 */
export type StatusMeaning = "success" | "failure" | "progress" | "waiting" | "attention" | "inactive" | "info";

export const STATUS_TONES: Record<StatusMeaning, BadgeTone> = {
    success: "green",
    failure: "red",
    progress: "purple",
    waiting: "blue",
    attention: "amber",
    inactive: "neutral",
    info: "sky",
};

/** The classes of a solid badge for a status of this meaning. */
export function statusClassName(meaning: StatusMeaning): string {
    return SOLID_TONE_CLASS_NAMES[STATUS_TONES[meaning]];
}

const STATUS_DOT_CLASS_NAMES: Record<StatusMeaning, string> = {
    success: "bg-green-500",
    failure: "bg-red-600",
    progress: "bg-purple-500",
    waiting: "bg-blue-500",
    attention: "bg-amber-500",
    inactive: "bg-slate-500",
    info: "bg-sky-500",
};

/** The fill of a status dot of this meaning: the badge's color without its text. */
export function statusDotClassName(meaning: StatusMeaning): string {
    return STATUS_DOT_CLASS_NAMES[meaning];
}

export { Badge, badgeVariants };
