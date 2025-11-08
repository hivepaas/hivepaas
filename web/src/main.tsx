import { StrictMode } from "react";

import { createRoot } from "react-dom/client";

import { Toaster } from "@components/ui/sonner.tsx";
import { DeviceInfo } from "@infrastructure/device";

import App from "./App.tsx";
import "./index.css";

DeviceInfo.init();

createRoot(document.getElementById("root")!).render(
    <StrictMode>
        <App />
        <Toaster theme="light" />
    </StrictMode>,
);
