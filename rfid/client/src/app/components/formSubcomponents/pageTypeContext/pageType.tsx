'use client'
import {createContext} from "react";

export const PageTypeContext = createContext("unknown")
export function PageTypeProvider(props:React.PropsWithChildren<{pageType: string}>){
    return <PageTypeContext.Provider value={props.pageType}>
        {/*TODO: working<TestAndValidate todos={["validate depths"]}>{"Depth: "+depth}</TestAndValidate>*/}
        {props.children}
    </PageTypeContext.Provider>
}