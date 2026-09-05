import { RestartServicesSection } from "../building-blocks/restart-services-section.com";

export function SystemSettingsHivePaaSActionsRoute() {
    return (
        <div className="flex flex-col gap-6">
            <RestartServicesSection />
        </div>
    );
}
