import { DatabaseEncryptionSection } from "../building-blocks/database-encryption-section.com";

export function SystemSettingsHivePaaSSecurityRoute() {
    return (
        <div className="flex flex-col gap-6">
            <DatabaseEncryptionSection />
        </div>
    );
}
