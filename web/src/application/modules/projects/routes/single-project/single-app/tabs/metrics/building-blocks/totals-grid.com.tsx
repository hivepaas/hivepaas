/** A range's numbers, side by side above its charts. */
export function TotalsGrid({ items }: { items: { label: string; value: string }[] }) {
    return (
        <div className="grid grid-cols-2 gap-3 sm:grid-cols-5">
            {items.map(item => (
                <div
                    key={item.label}
                    className="flex flex-col gap-0.5 rounded-md border px-3 py-2"
                >
                    <span className="text-xs text-muted-foreground">{item.label}</span>
                    <span className="text-lg font-semibold tabular-nums">{item.value}</span>
                </div>
            ))}
        </div>
    );
}
