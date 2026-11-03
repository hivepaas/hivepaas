import type { PropsWithChildren } from "react";

import { cn } from "@/lib/utils";
import { ExternalLink } from "lucide-react";

interface Props {
    href: string;
    className?: string;
}

/** A page of the docs on plans and licenses, opened in a new tab. */
export function LicenseDocsLink({ href, className, children }: PropsWithChildren<Props>) {
    return (
        <a
            href={href}
            target="_blank"
            rel="noopener noreferrer"
            className={cn("inline-flex items-center gap-1 underline-offset-4 hover:underline", className)}
        >
            {children}
            <ExternalLink className="size-3.5 shrink-0" />
        </a>
    );
}
