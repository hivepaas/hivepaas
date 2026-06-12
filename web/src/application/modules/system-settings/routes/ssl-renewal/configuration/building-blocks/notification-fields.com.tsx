import { useNotificationSettingsSources } from "~/settings/module-shared/hooks";

import { NotificationSettings } from "@application/shared/form";

import type { SystemSslRenewalConfigurationFormInput } from "../schemas";

import { SectionHeader } from "./section-header.com";

export function NotificationFields({ readOnly = false }: Props) {
    const { sources, manageLink } = useNotificationSettingsSources({ type: "settings" });

    return (
        <>
            <SectionHeader>Notification Configuration</SectionHeader>
            <div className="px-3">
                <NotificationSettings<SystemSslRenewalConfigurationFormInput>
                    names={{
                        successUseDefault: "notification.successUseDefault",
                        success: "notification.success",
                        failureUseDefault: "notification.failureUseDefault",
                        failure: "notification.failure",
                    }}
                    sources={sources}
                    manageLink={manageLink}
                    readOnly={readOnly}
                />
            </div>
        </>
    );
}

interface Props {
    readOnly?: boolean;
}
