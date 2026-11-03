import { type ApiRequestBase, type ApiResponseBase } from "@infrastructure/api";

import {
    type FunctionFile,
    type FunctionTestRequest,
    type FunctionTestRunResult,
} from "../../../../domain/apps/function";

/**
 * A test run: the function's code as the editor has it, not yet saved, and the
 * request to call it with.
 */
export type AppFunction_TestRun_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
    files: FunctionFile[];
    request: FunctionTestRequest;
}>;

export type AppFunction_TestRun_Res = ApiResponseBase<FunctionTestRunResult>;
