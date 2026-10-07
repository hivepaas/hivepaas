import { RestartServicesSection } from "../building-blocks/restart-services-section.com";
import { SystemAppsSection } from "../building-blocks/system-apps-section.com";

export function SystemSettingsHivePaaSActionsRoute() {
    return (
        <div className="flex flex-col gap-6">
            <SystemAppsSection />
            <RestartServicesSection />
        </div>
    );
}
