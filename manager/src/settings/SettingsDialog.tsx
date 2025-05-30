import { Proxy } from "../api/api";
import { IoClose } from "react-icons/io5";
import { ProxySettingsView } from "./ProxySettings";
import {
    Tabs,
    TabsContent,
    TabsList,
    TabsTrigger,
} from "../components/ui/tabs";
import { RequestsViewConfig } from "../types";
import { ViewSettings } from "./ViewSettings";

export function SettingsDialog(props: {
    proxy: Proxy | null;
    open: boolean;
    setOpen: React.Dispatch<React.SetStateAction<boolean>>;
    requestsViewConfig: RequestsViewConfig;
    setRequestsViewConfig: React.Dispatch<
        React.SetStateAction<RequestsViewConfig>
    >;
}) {
    if (props.proxy === null) return <></>;

    return (
        <dialog
            open={props.open}
            hidden={!props.open}
            className="fixed inset-0 z-50 flex items-center justify-center w-full h-full"
        >
            <div className="w-[90%] h-[90%] bg-gray-100 rounded shadow-lg flex flex-col overflow-hidden">
                {/* Entire inner content in vertical flex layout */}
                <div className="flex flex-col flex-grow min-h-0 px-2 py-2">
                    <button
                        className="ml-auto bg-black w-6 aspect-square"
                        onClick={() => props.setOpen(false)}
                    >
                        <IoClose className="text-white mx-auto" />
                    </button>

                    <h1 className="text-3xl text-center">Settings</h1>
                    <p className="text-center mb-4">
                        configuration stuff goes here :/
                    </p>

                    {/* Flex-grow Tabs with min-h-0 is critical */}
                    <Tabs
                        defaultValue="proxy"
                        className="flex flex-col flex-grow min-h-0"
                    >
                        <TabsList>
                            <TabsTrigger value="requests-view">
                                Request View
                            </TabsTrigger>
                            <TabsTrigger value="proxy">Proxy</TabsTrigger>
                        </TabsList>

                        {/* This is now scrollable if needed */}
                        <TabsContent
                            value="proxy"
                            className="flex-grow min-h-0 overflow-y-auto pr-1"
                        >
                            <ProxySettingsView proxy={props.proxy} />
                        </TabsContent>

                        <TabsContent
                            className="flex-grow min-h-0 overflow-y-auto pr-1"
                            value="requests-view"
                        >
                            <ViewSettings
                                requestsViewConfig={props.requestsViewConfig}
                                setRequestsViewConfig={
                                    props.setRequestsViewConfig
                                }
                            />
                        </TabsContent>
                    </Tabs>
                </div>
            </div>
        </dialog>
    );
}
