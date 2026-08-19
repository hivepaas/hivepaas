import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type AppContainerFileCompressionFormat = "" | "gzip" | "zstd" | "zip" | "tar";

export type AppContainerFiles_DownloadOne_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
    nodeId: string;
    containerId: string;
    path: string;
    isDir: boolean;
    compressionFormat: AppContainerFileCompressionFormat;
}>;

export type AppContainerFiles_DownloadOne_Res = ApiResponseBase<{
    blob: Blob;
    filename?: string;
}>;

export type AppContainerFiles_UploadOne_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
    nodeId: string;
    containerId: string;
    path: string;
    file: File;
    extract: boolean;
    compressionFormat: AppContainerFileCompressionFormat;
    overwrite: boolean;
}>;

export type AppContainerFiles_UploadOne_Res = ApiResponseBase<{
    path: string;
    message: string;
}>;
