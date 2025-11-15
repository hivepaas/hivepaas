import { createContext } from "react";

import { UsersApi, UsersApiValidator } from "~/user-management/api/services";

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
