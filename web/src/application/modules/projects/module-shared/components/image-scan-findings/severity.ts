import type { VulnIgnoreScope, VulnSeverity, VulnSummary } from "~/projects/api/services";

import type { BadgeTone } from "@/components/ui/badge";

/** The severities, the worst first, and how each is shown. */
export const SEVERITIES: { severity: VulnSeverity; key: keyof Omit<VulnSummary, "ignored">; label: string }[] = [
    { severity: "CRITICAL", key: "critical", label: "Critical" },
    { severity: "HIGH", key: "high", label: "High" },
    { severity: "MEDIUM", key: "medium", label: "Medium" },
    { severity: "LOW", key: "low", label: "Low" },
    { severity: "UNKNOWN", key: "unknown", label: "Unknown" },
];

const TONES: Record<VulnSeverity, BadgeTone> = {
    CRITICAL: "red",
    HIGH: "orange",
    MEDIUM: "amber",
    LOW: "sky",
    UNKNOWN: "neutral",
};

export function severityTone(severity: VulnSeverity): BadgeTone {
    return TONES[severity];
}

export function severityLabel(severity: VulnSeverity): string {
    return SEVERITIES.find(item => item.severity === severity)?.label ?? severity;
}

/** An image's ID as Docker shortens one. */
export function shortImageId(imageId: string): string {
    return imageId.replace(/^sha256:/, "").slice(0, 12);
}

/** Where an ignore applies, as the screens say it. */
export const SCOPE_LABELS: Record<VulnIgnoreScope, string> = {
    "app": "This app",
    "project-env": "This environment",
    "project": "This project",
    "global": "Every project",
};

/** Where an ignore applies: its scope, and for a global one whether the apps have it. */
export function ignoreScopeLabel(ignore: { scope: VulnIgnoreScope; inheritable: boolean }): string {
    if (ignore.scope === "global" && !ignore.inheritable) {
        return "HivePaaS only";
    }
    return SCOPE_LABELS[ignore.scope];
}
