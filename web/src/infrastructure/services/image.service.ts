import { Err, Ok, type Result } from "oxide.ts";

import { ImageConversionWorker } from "@workers";

type FileType = "image/jpeg" | "image/png" | "image/webp" | "image/gif";

interface Options {
    filename: string;
    quality: number;
    type: FileType;
}

const extensionMap: Record<FileType, string> = {
    "image/jpeg": ".jpg",
    "image/png": ".png",
    "image/webp": ".webp",
    "image/gif": ".gif",
} as const;

async function convert(bitmap: ImageBitmap, options: Options): Promise<Result<{ file: File }, Error>> {
    const worker = new ImageConversionWorker();

    const { filename, quality, type } = options;

    if (filename.trim() === "") {
        return Err(new Error("Filename is required"));
    }

    if (quality < 0 || quality > 1) {
        return Err(new Error("Quality must be between 0 and 1"));
    }

    return new Promise(resolve => {
        worker.postMessage(
            {
                bitmap,
                filename: filename + extensionMap[type],
                quality,
                type,
            },
            [bitmap],
        );

        worker.onmessage = ({ data: file }: MessageEvent<File>) => {
            worker.terminate();

            resolve(Ok({ file }));
        };

        worker.onerror = (ev: ErrorEvent) => {
            worker.terminate();

            resolve(Err(ev.error));
        };
    });
}

/**
 * Load an image from a URL and convert it to a blob URL for local editing
 */
async function loadFromUrl(url: string): Promise<Result<{ blobUrl: string }, Error>> {
    try {
        const response = await fetch(url);

        if (!response.ok) {
            return Err(new Error(`Failed to fetch image: ${response.statusText}`));
        }

        const blob = await response.blob();
        const blobUrl = URL.createObjectURL(blob);

        return Ok({ blobUrl });
    } catch (error) {
        return Err(error instanceof Error ? error : new Error("Failed to load image from URL"));
    }
}

async function convertFileToBase64(file: File): Promise<string> {
    return new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = () => {
            if (typeof reader.result === "string") {
                resolve(reader.result);
            } else {
                reject(new Error("Failed to convert file to base64"));
            }
        };
        reader.onerror = () => {
            reject(new Error("Error reading file"));
        };
        reader.readAsDataURL(file);
    });
}

export const ImageService = Object.freeze({
    convert,
    loadFromUrl,
    convertFileToBase64,
});
