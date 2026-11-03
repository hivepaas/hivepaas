import { RestartTraefikSection } from "../building-blocks/restart-traefik-section.com";

export function SystemSettingsTraefikActionsRoute() {
    return (
        <div className="flex flex-col gap-6">
            <RestartTraefikSection />
        </div>
    );
}
