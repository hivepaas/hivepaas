/**
 * A creation the plan in force refused: what an ERR_PLAN_LIMIT_REACHED says.
 *
 * limit, max and plan are the error's params, as the server names them. When
 * they are missing or not as expected, limit and plan are empty and max is null,
 * and only the server's message is left to show.
 */
export interface PlanLimitRefusal {
    /** The kind the plan limits, such as "apps". */
    limit: string;
    /** The most of it the plan allows. 0 is a kind the plan leaves out, such as Bee's paid sign-in. */
    max: number | null;
    /** The plan in force, such as "comb". */
    plan: string;
    /** The server's own sentence, in the person's language. */
    message: string;
}

export interface PlanLimitReachedDialogProps {
    refusal?: PlanLimitRefusal;
}

export interface PlanLimitReachedDialogState {
    mode: "open" | "closed";
}

export interface PlanLimitReachedDialogOptions {
    props: PlanLimitReachedDialogProps;
}
