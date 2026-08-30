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
        svg.hivepaas-logo,
        svg.hivepaas-logo * {
            filter: none !important;
        }
        .hivepaas-logo .hivepaas-logo-hex {
            fill: #111827 !important;
            stroke: #374151 !important;
        }
        .hivepaas-logo .hivepaas-logo-inner,
        .hivepaas-logo .hivepaas-logo-inner rect,
        .hivepaas-logo .hivepaas-logo-inner polygon {
            fill: #fef7eb !important;
            stroke: #fef7eb !important;
        }
        .tab-active-indicator,
        .active-indicator {
            background-color: #f59e0b !important;
        }
    `,
    ignoreInlineStyle: [
        ".x-logo",
        ".hivepaas-logo",
        "svg.hivepaas-logo",
        "svg.hivepaas-logo *",
        "[class*='terminalFrame']",
        "[class*='terminalHost']",
    ],
    ignoreImageAnalysis: [
        ".x-logo",
        ".hivepaas-logo",
        "svg.hivepaas-logo",
        "svg.hivepaas-logo *",
        "[class*='terminalFrame']",
        "[class*='terminalHost']",
    ],
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
