import { z } from "zod";
import type { AppContainerFileCompressionFormat } from "~/projects/api/services";

export const ExportContainerFilesFormSchema = z.object({
    path: z.string().trim().min(1, "Path is required").max(512),
    isDir: z.boolean(),
    compression: z.enum(["none", "gzip", "zstd"]),
});

export type ExportContainerFilesFormInput = z.input<typeof ExportContainerFilesFormSchema>;
export type ExportContainerFilesFormOutput = z.output<typeof ExportContainerFilesFormSchema>;

export function mapExportCompressionToWire(
    compression: ExportContainerFilesFormOutput["compression"],
): AppContainerFileCompressionFormat {
    if (compression === "none") {
        return "";
    }

    return compression;
}
