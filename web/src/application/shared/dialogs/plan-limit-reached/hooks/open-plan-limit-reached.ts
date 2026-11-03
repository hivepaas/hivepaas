import { planLimitRefusalOf } from "../utils";

import { usePlanLimitReachedDialogState } from "./use-plan-limit-reached.dialog.state";

/**
 * Opens the plan-limit dialog when an error is a creation the plan in force
 * refused, and says whether it did. When it did, the caller should say nothing
 * more: the dialog is the whole answer, and a toast beside it would repeat it.
 *
 * useApiErrorNotifications does this for every failure that reaches notifyError;
 * call it where a failure does not. It reads the store directly rather than
 * through a hook, so any callback can call it.
 */
export function openPlanLimitReachedDialog(error: unknown): boolean {
    const refusal = planLimitRefusalOf(error);

    if (refusal === null) {
        return false;
    }

    usePlanLimitReachedDialogState.getState().open({ props: { refusal } });

    return true;
}
