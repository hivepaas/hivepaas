import type { TraefikServiceSettings } from "~/system-settings/domain";

import type { TraefikGeneralFormInput } from "../schemas";

export function mapTraefikServiceSettingsToFormInput(settings: TraefikServiceSettings): TraefikGeneralFormInput {
    return {
        appSettings: {
            replicas: settings.appSettings.replicas,
        },
    };
}
