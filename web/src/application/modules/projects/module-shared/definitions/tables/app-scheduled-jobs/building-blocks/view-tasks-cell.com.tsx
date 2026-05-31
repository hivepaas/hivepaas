import { memo } from "react";

import { Button } from "@components/ui/button";

function View() {
    return (
        <Button
            type="button"
            variant="ghost"
            className="h-8 justify-start px-0 text-link"
            disabled
        >
            View Tasks
        </Button>
    );
}

export const ViewTasksCell = memo(View);
