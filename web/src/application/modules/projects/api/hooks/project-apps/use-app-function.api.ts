import { use, useMemo } from "react";

import { match } from "oxide.ts";

import { type AppFunction_TestRun_Req } from "../../../api/services";
import { ProjectsApiContext } from "../../api-context/projects.api.context";

function createHook() {
    return function useAppFunctionApi() {
        const { api } = use(ProjectsApiContext);

        const mutations = useMemo(
            () => ({
                // A failed run is shown in the test panel, not as a notification.
                testRun: async (request: AppFunction_TestRun_Req["data"]) => {
                    const result = await api.projects.apps.function.$.testRun({ data: request });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            throw error;
                        },
                    });
                },
            }),
            [api],
        );

        return { mutations };
    };
}

export const useAppFunctionApi = createHook();
