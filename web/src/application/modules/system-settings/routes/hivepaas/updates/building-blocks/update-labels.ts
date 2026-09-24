import type { ReleaseChannel, UpdateComponent } from "~/system-settings/domain";

/** What each component the update reaches is called on the screen. */
const COMPONENT_NAMES: Record<string, string> = {
    "db": "PostgreSQL (HivePaaS database)",
    "redis": "Redis",
    "traefik": "Traefik",
    "victoria-logs": "VictoriaLogs (log store)",
    "vlagent": "vlagent (log collector)",
    "registry": "Registry (zot)",
    "app": "HivePaaS",
    "worker": "HivePaaS workers",
};

export function componentName(component: UpdateComponent): string {
    return COMPONENT_NAMES[component.key] ?? component.key;
}

/** The tag of an image, which is what differs between two versions of one. */
export function imageTag(image: string): string {
    if (!image) {
        return "–";
    }
    const lastSlash = image.lastIndexOf("/");
    const colon = image.indexOf(":", lastSlash + 1);
    return colon >= 0 ? image.slice(colon + 1) : image;
}

export function channelLabel(channel: ReleaseChannel): string {
    return channel === "beta" ? "Beta" : "Stable";
}

export function formatReleaseDate(date: Date | null): string | undefined {
    return date ? date.toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" }) : undefined;
}
