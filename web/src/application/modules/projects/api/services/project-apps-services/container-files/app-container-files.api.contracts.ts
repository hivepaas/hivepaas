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
    /** Told how much of the file has left the browser, as it goes. */
    onProgress?: (sent: number, total: number) => void;
}>;

export type AppContainerFiles_UploadOne_Res = ApiResponseBase<{
    path: string;
    message: string;
}>;

/**
 * The query of an upload's stream - which a request without the websocket
 * upgrade only checks - as the server takes it.
 */
export function appContainerFilesUploadQuery(
    data: AppContainerFiles_UploadOne_Req["data"],
): Record<string, string | number | boolean> {
    const { nodeId, containerId, path, file, extract, compressionFormat, overwrite } = data;

    return {
        nodeId,
        containerId,
        path,
        extract,
        compressionFormat,
        overwrite,
        fileName: file.name,
        fileSize: file.size,
        progress: true,
    };
}
