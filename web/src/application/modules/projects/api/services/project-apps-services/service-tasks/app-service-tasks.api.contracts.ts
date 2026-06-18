import type { AppServiceTask } from "~/projects/domain";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type AppServiceTasks_FindMany_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
}>;

export type AppServiceTasks_FindMany_Res = ApiResponseBase<AppServiceTask[]>;
