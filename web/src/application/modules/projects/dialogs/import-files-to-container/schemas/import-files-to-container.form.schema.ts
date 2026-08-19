import { z } from "zod";

import type { AppContainerFileCompressionFormat } from "~/projects/api/services";

export const ImportFilesToContainerFormSchema = z.object({
    file: z.instanceof(File, { message: "File is required" }),
    path: z.string().trim().min(1, "Destination path is required").max(512),
    extract: z.boolean(),
    compression: z.enum(["auto", "tar", "zip", "gzip", "zstd"]),
    overwrite: z.boolean(),
});

export type ImportFilesToContainerFormInput = z.input<typeof ImportFilesToContainerFormSchema>;
export type ImportFilesToContainerFormOutput = z.output<typeof ImportFilesToContainerFormSchema>;

export function mapImportCompressionToWire(
    compression: ImportFilesToContainerFormOutput["compression"],
): AppContainerFileCompressionFormat {
    if (compression === "auto") {
        return "";
    }

    return compression;
}
