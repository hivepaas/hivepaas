import type { ProjectDomainSettings } from "~/projects/domain";

import { ESslCertType, ESslKeyType } from "@application/shared/enums";

import { type ProjectDomainSettingsFormSchemaInput, ProjectDomainSettingsKeyTypeUnspecified } from "../schemas";

export function mapProjectDomainSettingsToFormInput(data: ProjectDomainSettings): ProjectDomainSettingsFormSchemaInput {
    return {
        allowedDomains: data.allowedDomains.map(value => ({ value })),
        certSettings: {
            certType:
                data.certSettings?.certType === ESslCertType.Custom ? ESslCertType.Custom : ESslCertType.LetsEncrypt,
            email: data.certSettings?.email ?? "",
            keyType: data.certSettings
                ? (data.certSettings.keyType ?? ProjectDomainSettingsKeyTypeUnspecified)
                : ESslKeyType.ECP256,
        },
    };
}
