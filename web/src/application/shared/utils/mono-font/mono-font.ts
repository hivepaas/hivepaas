/** The system's monospace faces: what shows while the dashboard's own is loading. */
export const SYSTEM_MONO_FONT_FAMILY = "ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace";

/** The dashboard's monospace face, for code, scripts, logs and terminals. */
export const MONO_FONT_FAMILY = `'Geist Mono Variable', ${SYSTEM_MONO_FONT_FAMILY}`;

/**
 * Calls `onLoaded` once the dashboard's monospace face has loaded at `fontSize`. A terminal measures its
 * cells when its font is set: set to a face still loading, it would keep the cells of the fallback, so it
 * starts on the system's face and takes this one here. Returns a cancel for the effect's cleanup.
 */
export function whenMonoFontLoaded(fontSize: number, onLoaded: () => void): () => void {
    let cancelled = false;

    document.fonts
        .load(`${String(fontSize)}px 'Geist Mono Variable'`)
        .then(() => {
            if (!cancelled) {
                onLoaded();
            }
        })
        .catch(() => {
            // The system's face stays.
        });

    return () => {
        cancelled = true;
    };
}
