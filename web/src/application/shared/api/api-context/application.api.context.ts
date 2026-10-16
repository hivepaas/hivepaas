import { createContext } from "react";

import {
    LicenseApi,
    LicenseApiValidator,
    ProfileApi,
    ProfileApiValidator,
    SessionApi,
    SessionApiValidator,
    SettingUsageApi,
    SettingUsageApiValidator,
    SupportFeedbacksApi,
    SupportFeedbacksApiValidator,
} from "@application/shared/api/services";

function createApplicationApi() {
    return {
        profile: new ProfileApi(new ProfileApiValidator()),
        session: new SessionApi(new SessionApiValidator()),
        settingUsage: new SettingUsageApi(new SettingUsageApiValidator()),
        license: new LicenseApi(new LicenseApiValidator()),
        support: {
            feedbacks: new SupportFeedbacksApi(new SupportFeedbacksApiValidator()),
        },
    };
}

export const ApplicationApiContext = createContext({
    api: createApplicationApi(),
});
