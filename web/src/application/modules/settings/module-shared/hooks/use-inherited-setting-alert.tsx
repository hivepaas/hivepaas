import { useGlobalAlertDialog } from "@application/shared/dialogs";

export function isInheritedProjectSetting(scope: { type: string }, inherited?: boolean): boolean {
    return scope.type === "project" && inherited === true;
}

export function useInheritedSettingAlert() {
    const globalAlertDialog = useGlobalAlertDialog();

    return {
        open: () => {
            globalAlertDialog.actions.open({
                props: {
                    title: "Update setting",
                    description: (
                        <>
                            <span className="text-orange-500">Important:</span> Inherited and imported settings cannot
                            be changed here. Please go to the corresponding scope to make changes.
                        </>
                    ),
                    actionText: "Close",
                    cancelText: "",
                },
            });
        },
    };
}
