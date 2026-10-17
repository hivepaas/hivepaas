import { z } from "zod";

import { HttpException } from "@infrastructure/exceptions/http";

import type { PlanLimitRefusal } from "../types";

/**
 * A creation the plan in force does not allow. A 412, like every license error:
 * see the backend's hivepaas_app/hperrors/errors_license.go.
 */
export const PLAN_LIMIT_REACHED = "ERR_PLAN_LIMIT_REACHED";

/**
 * The params the refusal exposes, as the backend's planLimitReached sets them
 * (licenseserviceimpl/check.go). Max is a JSON integer, and 0 for a kind the plan
 * leaves out, such as Bee's paid sign-in.
 */
const PlanLimitParamsSchema = z.object({
    Limit: z.string().min(1),
    Max: z.number().int().nonnegative(),
    Plan: z.string().min(1),
});

/**
 * The refusal an error stands for: a creation the plan in force does not allow.
 * Null for any other error.
 *
 * Params that are missing or not of this shape still make a refusal, one with
 * only the server's message: the person is still owed the dialog that says why
 * nothing was created, and where to go about it.
 */
export function planLimitRefusalOf(error: unknown): PlanLimitRefusal | null {
    if (!(error instanceof HttpException) || error.code !== PLAN_LIMIT_REACHED) {
        return null;
    }

    const params = PlanLimitParamsSchema.safeParse(error.problem.params);

    if (!params.success) {
        return { limit: "", max: null, plan: "", message: error.message };
    }

    return { limit: params.data.Limit, max: params.data.Max, plan: params.data.Plan, message: error.message };
}

/** Whether an error is a creation the plan in force refused. */
export function isPlanLimitReached(error: unknown): boolean {
    return planLimitRefusalOf(error) !== null;
}
