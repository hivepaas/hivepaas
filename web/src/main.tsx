import { StrictMode } from "react";

import { createRoot } from "react-dom/client";

import { Toaster } from "@components/ui/sonner";
import { DeviceInfo } from "@infrastructure/device";

import { EnvConfig } from "@config";

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
        <Toaster theme="light" />
    </StrictMode>,
);
