import { StrictMode } from "react";

import { createRoot } from "react-dom/client";

import "@fontsource-variable/ibm-plex-sans";
import "@fontsource/ibm-plex-mono/400";
import "@fontsource/ibm-plex-mono/500";
import "@fontsource/ibm-plex-mono/600";
import "@fontsource/ibm-plex-mono/700";
import { DeviceInfo } from "@infrastructure/device";

import { EnvConfig } from "@config";

import { AppToaster } from "@application/shared/color-mode";
import "@application/shared/utils/time-ago";

import App from "./App";
import "./index.css";

if (!EnvConfig.API_URL) {
    throw new Error("API_URL is not configured");
}

DeviceInfo.init();

createRoot(document.getElementById("root")!).render(
    <StrictMode>
        <App />
        <AppToaster />
    </StrictMode>,
);
