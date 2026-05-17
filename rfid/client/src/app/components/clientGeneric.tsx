'use client'
import {ReaderOptionsContextProvider} from "@/app/components/formSubcomponents/readerWriterButtons/readerOptsContext";
import React, {ReactNode, useContext} from "react";
import {GoogleApiClient} from "@/app/components/Constants";
import TopBar from "@/app/components/TopBar";
import {QueryClient, QueryClientProvider} from "@tanstack/react-query";
import {CookiesProvider} from "react-cookie";
import {GoogleOAuthProvider} from "@react-oauth/google";
import {ListResult} from "@/app/components/formSubcomponents/shared";
import {DepthProvider} from "@/app/components/formSubcomponents/depthContext/depth";
import {PageTypeContext, PageTypeProvider} from "@/app/components/formSubcomponents/pageTypeContext/pageType";


export default function PageWrapper(
    {
        props, children
    }: {
        props: {
            readers: string[]
            pageType: string
        },
        children: ReactNode,
    }) {
    const queryClient = new QueryClient({
        defaultOptions: {
            queries: {
                staleTime: 5 * 60 * 1000, // Fresh for 5 minutes (data won't refetch in background during this time)
                gcTime: 10 * 60 * 1000,  // Stays in memory for 10 minutes after going inactive
            },
        },
    })
    return <ReaderOptionsContextProvider initialState={{options: props.readers, selected: undefined}}>
        <CookiesProvider>
            <GoogleOAuthProvider clientId={GoogleApiClient}>
                <QueryClientProvider client={queryClient}>
                    <PageTypeProvider pageType={props.pageType}>
                        <TopBar/>
                        {children}
                    </PageTypeProvider>
                </QueryClientProvider>
            </GoogleOAuthProvider>
        </CookiesProvider>
    </ReaderOptionsContextProvider>
}

// export function Redirector(itemType: string) {
//     return (newId: string) => {
//         console.log("redirecting to " + BaseExternalUrl + "/view/" + itemType + "/" + newId)
//         window.location.assign(BaseExternalUrl + "/view/" + itemType + "/" + newId)
//     }
// }

// TODO: should this be completely removed? I highly doubt it
export function LatestListDisplay<T>({text, data, isPartOfLatestMostRecent, constructor}: {
    text?: string,
    data: T[],
    constructor: (data: T, key: number) => React.JSX.Element,
    isPartOfLatestMostRecent?: boolean
}) {
    const pageType = useContext(PageTypeContext)
    return <DepthProvider>
        <div className={"latestListDisplay"+(pageType==="list"?" listPageItem":"")}>
            <div>{(text || "MOST RECENT") + ":"}</div>
            {data.map((val, i) => {
                return constructor(val, i)
            })}
        </div>
    </DepthProvider>
}

// export function LatestMostRecentListDisplay<T>({data, constructor}: {
//     data: ListResult<T>,
//     constructor: (data: T, key: number) => React.JSX.Element
// }) {
//     const standardArea = (inpData: ListResult<T>) => {
//         if (inpData.standard === undefined || inpData.standard.length === 0) {
//             return <div></div> // TODO: ok?
//         }
//         return <LatestListDisplay text={"Standard"} data={inpData.standard} constructor={constructor}/>
//     }
//     const recentArea = (inpData: ListResult<T>) => {
//         if (inpData.recent === undefined || inpData.recent.length === 0) {
//             return <div></div> // TODO: ok?
//         }
//         return <LatestListDisplay data={inpData.recent} constructor={constructor}/>
//     }
//     return <>{/* Each std/recent area has its own depth provider*/}
//         {standardArea(data)}
//         {recentArea(data)}
//     </>
// }