import { Toaster } from "@components/ui/sonner";

import { useColorModeContext } from "./color-mode.store";

/**
 * The toaster, wired to the app's own colour mode.
 *
 * Sonner needs to be told the theme rather than reading the `dark` class: it
 * paints a handful of its own parts - the description line, the close button -
 * from internal light-mode constants that no CSS variable of ours reaches, and
 * only swaps them under its own `data-sonner-theme="dark"`. Left on the wrong
 * theme it renders near-black text and a near-black close icon onto our dark
 * panel, which is how both came to be invisible.
 *
 * The store's modes are the same three values Sonner accepts, "system"
 * included, so this hands the value straight over and lets Sonner do its own
 * `prefers-color-scheme` listening.
 */
export function AppToaster() {
    const mode = useColorModeContext(state => state.mode);

    return <Toaster theme={mode} />;
}
