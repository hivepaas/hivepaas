import { AllowPrivilegedAppsSection } from "../building-blocks/allow-privileged-apps-section.com";
import { DatabaseEncryptionSection } from "../building-blocks/database-encryption-section.com";
import { ReturnSecretsViaApiSection } from "../building-blocks/return-secrets-via-api-section.com";

export function SystemSettingsHivePaaSSecurityRoute() {
    return (
        <div className="flex flex-col gap-6">
            <DatabaseEncryptionSection />
            <ReturnSecretsViaApiSection />
            <AllowPrivilegedAppsSection />
        </div>
    );
}
