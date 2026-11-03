import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useComposeImportApi } from "~/operations/api/hooks";
import type {
    ComposeImport_Apply_Req,
    ComposeImport_Apply_Res,
    ComposeImport_Validate_Req,
    ComposeImport_Validate_Res,
} from "~/operations/api/services";

type ValidateReq = ComposeImport_Validate_Req["data"];
type ValidateRes = ComposeImport_Validate_Res;
type ValidateOptions = Omit<UseMutationOptions<ValidateRes, Error, ValidateReq>, "mutationFn">;

/** Reads a compose file and plans its project. A mutation only because it posts the file: it writes nothing. */
function useValidateCompose(options: ValidateOptions = {}) {
    const { mutations } = useComposeImportApi();

    return useMutation({
        mutationFn: (request: ValidateReq) => mutations.validate(request),
        ...options,
    });
}

type ApplyReq = ComposeImport_Apply_Req["data"];
type ApplyRes = ComposeImport_Apply_Res;
type ApplyOptions = Omit<UseMutationOptions<ApplyRes, Error, ApplyReq>, "mutationFn">;

function useApplyCompose({ onSuccess, ...options }: ApplyOptions = {}) {
    const { mutations } = useComposeImportApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (request: ApplyReq) => mutations.apply(request),
        onSuccess: (response, request, ...rest) => {
            // A project, its env, its apps and settings, deployments queued.
            void queryClient.invalidateQueries();
            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

export const ComposeImportCommands = Object.freeze({
    useValidateCompose,
    useApplyCompose,
});
