import { useEffect, useState } from "react";

import { Request } from "./types";
import { Proxy } from "./api";

import { RequestView } from "./RequestView";

export function IncomingView(props: { proxy: Proxy }) {
    const [requests, setRequests] = useState<Array<Request>>([]);

    useEffect(() => {
        props.proxy.manageRequests(() => {
            console.log("update cb", props.proxy.requests);
            const newObj = Object.assign([], props.proxy.requests);
            setRequests(newObj);
        });
    }, []);

    return (
        <div className="flex flex-col w-full h-full">
            <h1 className="ml-2 text-2xl font-bold mb-2">Incoming Requests</h1>
            <div className="w-full overflow-y-auto h-[80%]">
                {requests.map((request, index) => (
                    <RequestView key={index} request={request} />
                ))}
            </div>
        </div>
    );
}
