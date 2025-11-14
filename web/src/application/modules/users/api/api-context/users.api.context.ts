import { createContext } from "react";

import { UsersApi, UsersApiValidator } from "~/users/api/services";

function createApi() {
    /**
     * Users
     */
    const usersApiValidator = new UsersApiValidator();

    return {
        users: {
            $: new UsersApi(usersApiValidator),
        },
    };
}

export const UsersApiContext = createContext({
    api: createApi(),
});
