import type { ProjectDomainSettings } from "~/projects/domain";
import type {
    DomainSettings_UpdateOne_Payload,
    DomainSettings_UpdateStatus_Payload,
} from "~/settings/api/services/domain-settings-services";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type ProjectDomainSettings_FindOne_Req = ApiRequestBase<{
    projectID: string;
}>;
export type ProjectDomainSettings_FindOne_Res = ApiResponseBase<ProjectDomainSettings>;

export type ProjectDomainSettings_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    payload: DomainSettings_UpdateOne_Payload;
}>;
export type ProjectDomainSettings_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectDomainSettings_UpdateStatus_Req = ApiRequestBase<{
    projectID: string;
    payload: DomainSettings_UpdateStatus_Payload;
}>;
export type ProjectDomainSettings_UpdateStatus_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectDomainSettings_DeleteOne_Req = ApiRequestBase<{
    projectID: string;
}>;
export type ProjectDomainSettings_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;
