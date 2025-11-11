import { Dialog, DialogContent } from "@components/ui/dialog";

import { useF2aSetupDialogState } from "../hooks";

// import { type F2aSetupFormRef } from "../types";

// const fnPlaceholder = () => null;

export function F2aSetupDialog() {
    const {
        state,
        // props: { onSuccess = fnPlaceholder, onClose = fnPlaceholder, onError = fnPlaceholder } = {},
        // ...actions
    } = useF2aSetupDialogState();

    // const formRef = useRef<F2aSetupFormRef>(null);

    // const { mutate: create, isPending: isCreatePending } = PositionsPublicCommands.useCreateOne({
    //     onSuccess: (response, request) => {
    //         onSuccess(
    //             {
    //                 id: response.data.id,
    //                 ...request,
    //             },
    //             "create",
    //         );

    //         onClose();
    //     },
    //     onError,
    // });

    // const { mutate: update, isPending: isUpdatePending } = PositionsPublicCommands.useUpdateOne({
    //     onSuccess: (response, request) => {
    //         onSuccess(request, "edit");

    //         onClose();
    //     },
    //     onError,
    // });

    // function handleClose() {
    //     onClose();
    // }

    return (
        <Dialog open={state.mode !== "closed"}>
            <DialogContent></DialogContent>
        </Dialog>
    );
}
