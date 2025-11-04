import { memo } from "react";

import { NotFound404Image } from "@assets/images";

function View() {
    return (
        <div className="page-404-not-found">
            <div className="body">
                <div className="image-wrapper">
                    <NotFound404Image />
                </div>
                <h1 className="title">Sorry, the page not found</h1>
                <div className="description">
                    {"The link you followed probably broken\nor the page has been removed"}
                </div>
            </div>
        </div>
    );
}

export const Page404NotFound = memo(View);
