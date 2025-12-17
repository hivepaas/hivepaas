import { createContext } from "react";

import {
    ProfileApi,
    ProfileApiValidator,
    SessionApi,
    SessionApiValidator,
    SshKeysApi,
    SshKeysApiValidator,
} from "@application/shared/api/services";

function createApplicationApi() {
    return {
        profile: new ProfileApi(new ProfileApiValidator()),
        session: new SessionApi(new SessionApiValidator()),
        providers: {
            sshKeys: new SshKeysApi(new SshKeysApiValidator()),
        },
    };
}

export const ApplicationApiContext = createContext({
    api: createApplicationApi(),
});
