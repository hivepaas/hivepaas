export const dashedBorderBox =
    "border border-dashed border-primary-stroke rounded-lg px-3 sm:px-4 py-2 sm:py-2.5 bg-muted text-left text-[13px] leading-5";

export const moduleHeaderBox = "bg-background py-2 sm:py-4 px-3 sm:px-5 rounded-lg";

export const listBox =
    "bg-background rounded-lg p-2 sm:p-4 max-w-[1400px] w-full mx-auto has-[[data-form-action-bar]]:pb-0";

export const formBox =
    "bg-background rounded-lg p-2 sm:p-4 max-w-[1200px] w-full mx-auto has-[[data-form-action-bar]]:pb-0";

/**
 * An inline link inside running text.
 *
 * Colour and underline both, on purpose: colour alone excludes anyone who
 * cannot separate the two hues, and underline alone is what the app had - which
 * a reader takes for emphasis rather than a target. No `font-medium`; once the
 * colour carries the affordance, bolding it as well makes a link shout.
 */
export const inlineLink = "text-link underline underline-offset-4 hover:opacity-80";
