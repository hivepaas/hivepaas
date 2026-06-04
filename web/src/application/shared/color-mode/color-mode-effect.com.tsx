import { useEffect } from "react";

import * as DarkReader from "darkreader";

import { type ColorMode, useColorModeContext } from "./color-mode.store";

const DARK_READER_THEME = {
    brightness: 140,
    contrast: 90,
};

function applyColorMode(mode: ColorMode) {
    DarkReader.auto(false);

    if (mode === "dark") {
        DarkReader.enable(DARK_READER_THEME);
        return;
    }

    if (mode === "light") {
        DarkReader.disable();
        return;
    }

    DarkReader.auto(DARK_READER_THEME);
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
