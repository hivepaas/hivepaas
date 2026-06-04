import { createContext } from "react";

import {
    ProfileApi,
    ProfileApiValidator,
    SessionApi,
    SessionApiValidator,
    SupportFeedbacksApi,
    SupportFeedbacksApiValidator,
} from "@application/shared/api/services";

function createApplicationApi() {
    return {
        profile: new ProfileApi(new ProfileApiValidator()),
        session: new SessionApi(new SessionApiValidator()),
        support: {
            feedbacks: new SupportFeedbacksApi(new SupportFeedbacksApiValidator()),
        },
    };
}

export const ApplicationApiContext = createContext({
    api: createApplicationApi(),
});
