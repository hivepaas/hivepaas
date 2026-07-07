import type { PaginationState, SortingState } from "@infrastructure/data";
import type {
    ProjectCommandTemplate,
    ProjectCommandTemplateArgGroup,
    ProjectCommandTemplateConsoleSize,
    ProjectCommandTemplateEnvVar,
} from "~/projects/domain";

import type { ECommandTemplateKind, ESettingStatus } from "@application/shared/enums";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type ProjectCommandTemplate_FindManyPaginated_Req = ApiRequestBase<{
    projectID: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type ProjectCommandTemplate_FindManyPaginated_Res = ApiResponsePaginated<ProjectCommandTemplate>;

export type ProjectCommandTemplate_FindOneById_Req = ApiRequestBase<{
    projectID: string;
    id: string;
}>;

export type ProjectCommandTemplate_FindOneById_Res = ApiResponseBase<ProjectCommandTemplate>;

export type ProjectCommandTemplate_CreateOne_Payload = {
    availableInProjects: boolean;
    default: boolean;
    name: string;
    kind: ECommandTemplateKind;
    command: string;
    script: string;
    workingDir: string;
    envVars: ProjectCommandTemplateEnvVar[];
    argGroups: ProjectCommandTemplateArgGroup[];
    consoleSize: ProjectCommandTemplateConsoleSize;
    tty: boolean;
};

export type ProjectCommandTemplate_CreateOne_Req = ApiRequestBase<{
    projectID: string;
    payload: ProjectCommandTemplate_CreateOne_Payload;
}>;

export type ProjectCommandTemplate_CreateOne_Res = ApiResponseBase<{ id: string }>;

export type ProjectCommandTemplate_UpdateOne_Payload = ProjectCommandTemplate_CreateOne_Payload & {
    updateVer: number;
};

export type ProjectCommandTemplate_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    id: string;
    payload: ProjectCommandTemplate_UpdateOne_Payload;
}>;

export type ProjectCommandTemplate_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectCommandTemplate_UpdateStatus_Payload = {
    updateVer: number;
    status?: ESettingStatus;
    expireAt?: Date | null;
    availableInProjects?: boolean;
    default?: boolean;
};

export type ProjectCommandTemplate_UpdateStatus_Req = ApiRequestBase<{
    projectID: string;
    id: string;
    payload: ProjectCommandTemplate_UpdateStatus_Payload;
}>;

export type ProjectCommandTemplate_UpdateStatus_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectCommandTemplate_DeleteOne_Req = ApiRequestBase<{
    projectID: string;
    id: string;
}>;

export type ProjectCommandTemplate_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;
