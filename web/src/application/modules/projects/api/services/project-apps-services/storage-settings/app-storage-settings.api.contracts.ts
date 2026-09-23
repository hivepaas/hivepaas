import { type AppStorageSettings } from "~/projects/domain";

import { type ApiRequestBase, type ApiResponseBase } from "@infrastructure/api";

export type AppStorageSettings_FindOne_Req = ApiRequestBase<{ projectID: string; env: string; appID: string }>;
export type AppStorageSettings_FindOne_Res = ApiResponseBase<AppStorageSettings>;

export type AppStorageSettings_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
    payload: AppStorageSettings & {
        /**
         * Deletes what is already in the directories the mounts being added
         * reach, before they are mounted. It is how the preflight findings are
         * answered, and it is off unless asked for.
         */
        resetStorage?: boolean;
    };
}>;
export type AppStorageSettings_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

/** One mount being added whose directory already holds something. */
export type AppStorageFinding = {
    /** The path inside the container, which is how the screen finds the row. */
    target: string;
    volume: { id: string; name: string };
    /** The directory inside the volume, so an operator can go and look. */
    path: string;
};

export type AppStorageSettings_Preflight_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
    payload: { mounts: AppStorageSettings["mounts"] };
}>;
export type AppStorageSettings_Preflight_Res = ApiResponseBase<AppStorageFinding[]>;
