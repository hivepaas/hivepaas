import { SecuritySwitchSection } from "./security-switch-section.com";

export function ReturnSecretsViaApiSection() {
    return (
        <SecuritySwitchSection
            field="returnSecretsViaApi"
            title="Return Secrets via API"
            action="Returning Secrets via API"
            description={
                <p>
                    You can enable or disable the ability to return secrets on the dashboard or via the API here. Note:
                    Even when enabled, only administrators and users with the &quot;Can Reveal Secrets&quot; capability
                    can reveal secrets. Mounting a secret into an app&apos;s container, under the app&apos;s Setting
                    Mounts, does not depend on this switch: it takes the &quot;Can Reveal Secrets&quot; capability
                    alone.
                </p>
            }
        />
    );
}
