import type { AppScheduledJobs_Upsert_Payload } from "~/projects/api/services";
import type { AppScheduledJob } from "~/projects/domain";
import {
    EAppScheduledJobScheduleMode,
    EAppScheduledJobTaskPriority,
    EAppScheduledJobType,
} from "~/projects/module-shared/enums";
import { headerLinesOf, parseHeaderLines } from "~/projects/module-shared/utils";

import {
    createDefaultJobScheduleFormValues,
    mapJobScheduleFormValuesToPayload,
    mapJobScheduleToFormValues,
} from "../job-schedule-fields";

import {
    FUNCTION_INVOKE_METHODS,
    type FunctionInvokeFormInput,
    type FunctionInvokeFormOutput,
    methodSendsNoBody,
} from "./function-invoke.form.schema";

/** A new call is a GET of the root, run by hand until given a schedule, and can be canceled. */
export function createEmptyFunctionInvokeFormDefaults(): FunctionInvokeFormInput {
    return {
        name: "",
        ...createDefaultJobScheduleFormValues(EAppScheduledJobScheduleMode.None),
        method: "GET",
        path: "/",
        headers: "",
        body: "",
        timeout: "",
        maxRetry: 0,
        retryDelay: "",
        priority: EAppScheduledJobTaskPriority.Default,
        controlEnabled: true,
        notification: {
            successUseDefault: true,
            success: undefined,
            failureUseDefault: true,
            failure: undefined,
        },
    };
}

/** A method the form does not offer reads as GET. */
function knownMethod(method: string | undefined): FunctionInvokeFormInput["method"] {
    return FUNCTION_INVOKE_METHODS.find(item => item === method) ?? "GET";
}

export function mapFunctionInvokeToFormInput(job: AppScheduledJob): FunctionInvokeFormInput {
    const invoke = job.functionInvoke;

    return {
        name: job.name,
        ...mapJobScheduleToFormValues(job.schedule),
        method: knownMethod(invoke?.method),
        path: invoke?.path ?? "/",
        headers: headerLinesOf(invoke?.headers ?? {}),
        body: invoke?.body ?? "",
        timeout: job.timeout,
        maxRetry: job.maxRetry,
        retryDelay: job.retryDelay,
        priority: job.priority,
        controlEnabled: !job.controlDisabled,
        notification: {
            successUseDefault: job.notification?.successUseDefault ?? true,
            success: job.notification?.success,
            failureUseDefault: job.notification?.failureUseDefault ?? true,
            failure: job.notification?.failure,
        },
    };
}

/** A function's call is its app's: the function it calls. A method without a body sends none. */
export function mapFunctionInvokeFormToPayload(
    values: FunctionInvokeFormOutput,
    appId: string,
): AppScheduledJobs_Upsert_Payload {
    const headers = parseHeaderLines(values.headers);

    return {
        inheritable: false,
        default: false,
        name: values.name,
        jobType: EAppScheduledJobType.FunctionInvoke,
        schedule: mapJobScheduleFormValuesToPayload(values),
        app: { id: appId },
        priority: values.priority,
        maxRetry: values.maxRetry ?? 0,
        ...(values.retryDelay ? { retryDelay: values.retryDelay } : {}),
        retryBackoff: false,
        ...(values.timeout ? { timeout: values.timeout } : {}),
        controlDisabled: !values.controlEnabled,
        functionInvoke: {
            method: values.method,
            path: values.path || "/",
            ...(Object.keys(headers).length > 0 ? { headers } : {}),
            ...(!methodSendsNoBody(values.method) && values.body ? { body: values.body } : {}),
        },
        notification: {
            successUseDefault: values.notification.successUseDefault,
            ...(!values.notification.successUseDefault && values.notification.success
                ? { success: values.notification.success }
                : {}),
            failureUseDefault: values.notification.failureUseDefault,
            ...(!values.notification.failureUseDefault && values.notification.failure
                ? { failure: values.notification.failure }
                : {}),
        },
    };
}
