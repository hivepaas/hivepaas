import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useSpecImportApi } from "~/operations/api/hooks";
import type {
    SpecImport_Apply_Req,
    SpecImport_Apply_Res,
    SpecImport_Validate_Req,
    SpecImport_Validate_Res,
} from "~/operations/api/services";

type ValidateReq = SpecImport_Validate_Req["data"];
type ValidateRes = SpecImport_Validate_Res;
type ValidateOptions = Omit<UseMutationOptions<ValidateRes, Error, ValidateReq>, "mutationFn">;

/** Plans an import. A mutation only because it posts the bundle: it writes nothing. */
function useValidateImport(options: ValidateOptions = {}) {
    const { mutations } = useSpecImportApi();

    return useMutation({
        mutationFn: (request: ValidateReq) => mutations.validate(request),
        ...options,
    });
}

type ApplyReq = SpecImport_Apply_Req["data"];
type ApplyRes = SpecImport_Apply_Res;
type ApplyOptions = Omit<UseMutationOptions<ApplyRes, Error, ApplyReq>, "mutationFn">;

function useApplyImport({ onSuccess, ...options }: ApplyOptions = {}) {
    const { mutations } = useSpecImportApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (request: ApplyReq) => mutations.apply(request),
        onSuccess: (response, request, ...rest) => {
            // An import can write any project, env, app or setting of its scope,
            // and queue deployments and tasks: nothing cached can be assumed to
            // be what is there now.
            void queryClient.invalidateQueries();
            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

export const SpecImportCommands = Object.freeze({
    useValidateImport,
    useApplyImport,
});
