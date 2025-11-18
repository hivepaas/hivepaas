 import { type Dispatch, type SetStateAction, useState } from "react";

import { useDebounce } from "react-use";

export function useDebouncedSearch(
    ms: number = 250,
    initialValue: string = "",
): [string, Dispatch<SetStateAction<string>>, string] {
    const [search, setSearch] = useState(initialValue);
    const [debouncedSearch, setDebouncedSearch] = useState(initialValue);

    useDebounce(
        () => {
            setDebouncedSearch(search.trim());
        },
        ms,
        [search],
    );

    return [debouncedSearch, setSearch, search];
}
