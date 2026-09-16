import { type UseMutationOptions, useMutation } from "@tanstack/react-query";
import { useSpecExportApi } from "~/operations/api/hooks";
import type { SpecExport_Export_Req, SpecExport_Export_Res } from "~/operations/api/services";

type ExportReq = SpecExport_Export_Req["data"];
type ExportRes = SpecExport_Export_Res;
type ExportOptions = Omit<UseMutationOptions<ExportRes, Error, ExportReq>, "mutationFn">;

function useExportSpec(options: ExportOptions = {}) {
    const { mutations } = useSpecExportApi();

    return useMutation({
        mutationFn: (request: ExportReq) => mutations.exportSpec(request),
        ...options,
    });
}

export const SpecExportCommands = Object.freeze({
    useExportSpec,
});
