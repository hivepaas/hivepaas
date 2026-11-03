import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export interface AppTerminalInfo {
    enabled: boolean;
    supportedShells: string[];
}

export type AppTerminal_GetInfo_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
}>;

export type AppTerminal_GetInfo_Res = ApiResponseBase<AppTerminalInfo>;

export interface AppTerminalResizeMessage {
    type: "resize";
    width: number;
    height: number;
}

export interface AppTerminalInitMessage {
    type: "init";
    containerId: string;
    nodeId: string;
}

export type AppTerminalWs_Open_Req = {
    data: {
        projectID: string;
        env: string;
        appID: string;
        shell: string;
        width?: number;
        height?: number;
    };
};
