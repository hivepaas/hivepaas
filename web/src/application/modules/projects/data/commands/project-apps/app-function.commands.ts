import { type UseMutationOptions, useMutation } from "@tanstack/react-query";

import { useAppFunctionApi } from "../../../api/hooks/project-apps";
import { type AppFunction_TestRun_Req, type AppFunction_TestRun_Res } from "../../../api/services";

type TestRunReq = AppFunction_TestRun_Req["data"];
type TestRunRes = AppFunction_TestRun_Res;
type TestRunOptions = Omit<UseMutationOptions<TestRunRes, Error, TestRunReq>, "mutationFn">;

/**
 * Test-runs a function: nothing it changes is cached, so nothing is invalidated.
 */
function useTestRun(options: TestRunOptions = {}) {
    const { mutations } = useAppFunctionApi();

    return useMutation({
        mutationFn: mutations.testRun,
        ...options,
    });
}

export const AppFunctionCommands = Object.freeze({
    useTestRun,
});
