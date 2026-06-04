import { Startup } from "@/application";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import { ColorModeEffect } from "@application/shared/color-mode";

import "./App.css";

const queryClient = new QueryClient({
    defaultOptions: {
        queries: {
            refetchOnWindowFocus: false,
            staleTime: 60 * 1000,
            retry: false,
        },
    },
});

function App() {
    return (
        <QueryClientProvider client={queryClient}>
            <ColorModeEffect />
            <Startup />
        </QueryClientProvider>
    );
}

export default App;
