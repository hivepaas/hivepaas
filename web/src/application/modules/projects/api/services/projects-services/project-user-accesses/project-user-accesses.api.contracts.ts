import type { EUserRole } from "@application/shared/enums";

import { type ApiRequestBase, type ApiResponseBase } from "@infrastructure/api";

export interface ProjectUserAccessActions {
    read: boolean;
    execute: boolean;
    write: boolean;
    delete: boolean;
}

export interface ProjectUserAccessBase {
    id: string;
    username: string;
    email: string;
    fullName: string;
    photo: string | null;
    role: EUserRole;
    access: ProjectUserAccessActions;
}

export interface EnvUserAccessesData {
    name: string;
    userAccesses: ProjectUserAccessBase[];
}

export interface ProjectUserAccessesData {
    ownerAccess: ProjectUserAccessBase;
    userAccesses: ProjectUserAccessBase[];
    envUserAccesses: EnvUserAccessesData[];
    moduleUserAccesses: ProjectUserAccessBase[];
    currentUserActions: {
        canUpdateProjectUserAccesses: boolean;
        canViewModuleUserAccesses: boolean;
    };
    updateVer: number;
}

/**
 * Find project user accesses
 */
export type ProjectUserAccesses_FindOne_Req = ApiRequestBase<{
    projectID: string;
}>;

export type ProjectUserAccesses_FindOne_Res = ApiResponseBase<ProjectUserAccessesData>;

/**
 * Update project user accesses
 */
export type ProjectUserAccesses_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    updateVer: number;
    envUserAccesses: {
        name: string;
        userAccesses: {
            id: string;
            access: ProjectUserAccessActions;
        }[];
    }[];
}>;

export type ProjectUserAccesses_UpdateOne_Res = ApiResponseBase<{
    type: "success";
}>;
