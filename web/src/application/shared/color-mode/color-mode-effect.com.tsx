import { useEffect } from "react";

import * as DarkReader from "darkreader";

import { type ColorMode, useColorModeContext } from "./color-mode.store";

const DARK_READER_THEME = {
    brightness: 140,
    contrast: 90,
};

const DARK_READER_FIXES: DarkReader.DynamicThemeFix = {
    invert: [],
    css: `
        .hivepaas-logo-hex {
            fill: #f3f4f6 !important;
            stroke: #f3f4f6 !important;
        }
        .hivepaas-logo-inner,
        .hivepaas-logo-inner rect,
        .hivepaas-logo-inner polygon {
            fill: #181a1b !important;
            stroke: #181a1b !important;
        }
    `,
    ignoreInlineStyle: [".x-logo", ".hivepaas-logo"],
    ignoreImageAnalysis: [".x-logo", ".hivepaas-logo"],
    disableStyleSheetsProxy: false,
    ignoreCSSUrl: [],
};

function applyColorMode(mode: ColorMode) {
    DarkReader.auto(false);

    if (mode === "dark") {
        DarkReader.enable(DARK_READER_THEME, DARK_READER_FIXES);
        return;
    }

    if (mode === "light") {
        DarkReader.disable();
        return;
    }

    DarkReader.auto(DARK_READER_THEME, DARK_READER_FIXES);
}

function resetColorMode() {
    DarkReader.auto(false);
    DarkReader.disable();
}

export function ColorModeEffect() {
    const mode = useColorModeContext(state => state.mode);

    useEffect(() => {
        applyColorMode(mode);

        return resetColorMode;
    }, [mode]);

    return null;
}
