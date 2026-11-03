import { useId } from "react";

import { inlineLink } from "@/lib/styles";
import { cn } from "@/lib/utils";
import { AlertTriangle, ExternalLink, Info, type LucideIcon, OctagonAlert, X } from "lucide-react";

import { AppLink } from "@application/shared/components/navigation";

import { Button } from "@/components/ui/button";

import type { LicenseBannerLink, LicenseBannerTone } from "./license-banners.utils";

interface ToneStyle {
    icon: LucideIcon;
    box: string;
    iconColor: string;
}

/** The box of WeakAccountSecurityWarning, in each tone; each its own icon too, so the tone is not told by color alone. */
const TONE_STYLES: Record<LicenseBannerTone, ToneStyle> = {
    info: {
        icon: Info,
        box: "border-sky-500/30 bg-sky-500/[0.07] dark:border-sky-400/25 dark:bg-sky-400/[0.07]",
        iconColor: "text-sky-600 dark:text-sky-500",
    },
    warning: {
        icon: AlertTriangle,
        box: "border-amber-500/30 bg-amber-500/[0.07] dark:border-amber-400/25 dark:bg-amber-400/[0.07]",
        iconColor: "text-amber-600 dark:text-amber-500",
    },
    danger: {
        icon: OctagonAlert,
        box: "border-red-500/30 bg-red-500/[0.07] dark:border-red-400/25 dark:bg-red-400/[0.07]",
        iconColor: "text-red-600 dark:text-red-500",
    },
};

interface LicenseBannerProps {
    tone: LicenseBannerTone;
    title: string;
    body: string;
    links: LicenseBannerLink[];
    /** The dismiss button's hint: how long dismissing hides it, as in "Hide for 24 hours". */
    dismissHint: string;
    onDismiss: () => void;
}

/**
 * One license banner: its title, then what it means, then where to go about it.
 * Named by its title, so a screen reader - and a test - can tell two apart.
 */
export function LicenseBanner({ tone, title, body, links, dismissHint, onDismiss }: LicenseBannerProps) {
    const titleId = useId();
    const { icon: Icon, box, iconColor } = TONE_STYLES[tone];

    return (
        <div
            role="status"
            aria-labelledby={titleId}
            className={cn("flex items-start gap-2.5 rounded-lg border px-3.5 py-2.5", box)}
        >
            <Icon className={cn("mt-0.5 size-4 shrink-0", iconColor)} />
            <p className="min-w-0 flex-1 text-sm text-foreground">
                <span
                    id={titleId}
                    className="font-medium"
                >
                    {title}
                </span>
                {/* Not text-muted-foreground: on a tinted box it falls short of
                    the 4.5:1 contrast body text needs, in the light theme. */}
                {body && (
                    <>
                        {" "}
                        <span className="text-foreground/70">{body}</span>
                    </>
                )}
                {/* A space before the links rather than a margin: when they wrap
                    to a line of their own, they start where the text does. */}
                {links.length > 0 && (
                    <>
                        {" "}
                        <span className="inline-flex flex-wrap gap-x-3">
                            {links.map(link => (
                                <LicenseBannerLinkItem
                                    key={link.label}
                                    link={link}
                                />
                            ))}
                        </span>
                    </>
                )}
            </p>
            <Button
                type="button"
                variant="ghost"
                size="icon-sm"
                className="-my-1 -mr-1.5 size-7 shrink-0 text-muted-foreground"
                aria-label="Dismiss"
                title={dismissHint}
                onClick={onDismiss}
            >
                <X className="size-4" />
            </Button>
        </div>
    );
}

function LicenseBannerLinkItem({ link }: { link: LicenseBannerLink }) {
    if (!link.external) {
        return (
            <AppLink.Basic
                to={link.href}
                className={inlineLink}
            >
                {link.label}
            </AppLink.Basic>
        );
    }

    return (
        <a
            href={link.href}
            target="_blank"
            rel="noopener noreferrer"
            className={cn(inlineLink, "inline-flex items-center gap-1")}
        >
            {link.label}
            <ExternalLink className="size-3 shrink-0" />
        </a>
    );
}
