export function formatNodeMemory(memory: string | null | undefined): string {
    if (!memory) return "-";

    return memory.replace(/^(\d+(?:\.\d+)?)([a-z]+)$/i, (_, value: string, unit: string) => {
        return `${value} ${unit.toUpperCase()}`;
    });
}
