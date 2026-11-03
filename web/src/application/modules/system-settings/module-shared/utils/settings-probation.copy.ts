/**
 * How long the dashboard asks to keep a change on trial, and how to say it.
 *
 * The string and the number live together so the warning shown before saving
 * cannot promise a window the request does not ask for. The server clamps the
 * value and has its own floor, so this is a request rather than a guarantee -
 * which is why the wording says "about".
 */
export const CONFIRM_WINDOW = "5m";
export const CONFIRM_WINDOW_HUMAN = "about 5 minutes";

/**
 * The warning shown before a change that will be applied on trial.
 *
 * It exists because the trial is surprising the first time: the change takes
 * effect at once, and then undoes itself unless somebody comes back. Finding that
 * out by watching a change disappear is a bad way to learn it.
 */
export function probationWarning(subject: string): { title: string; description: string } {
    return {
        title: `Apply ${subject} now?`,
        description:
            `The change is applied immediately, and you then have ${CONFIRM_WINDOW_HUMAN} to confirm it. ` +
            "If you do not - because the new configuration locks you out, or simply because you walk away - " +
            "HivePaaS reverts it on its own, from inside the cluster.",
    };
}
