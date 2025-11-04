import "./App.css";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { Startup } from "@/application";

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
        <Startup />
    </QueryClientProvider>
  );
}

export default App;
