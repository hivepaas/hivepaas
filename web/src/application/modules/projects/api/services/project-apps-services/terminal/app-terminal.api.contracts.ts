import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export interface AppTerminalInfo {
    enabled: boolean;
    supportedShells: string[];
}

export type AppTerminal_GetInfo_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
}>;

export type AppTerminal_GetInfo_Res = ApiResponseBase<AppTerminalInfo>;

export interface AppTerminalResizeMessage {
    type: "resize";
    width: number;
    height: number;
}

export type AppTerminalWs_Open_Req = {
    data: {
        projectID: string;
        appID: string;
        shell: string;
        width?: number;
        height?: number;
    };
};
