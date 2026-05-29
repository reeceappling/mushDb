'use client'

import {createContext} from "react";

export const SessionContext = createContext("")
export function SessionProvider(props:React.PropsWithChildren<{session?:string}>){
    return <SessionContext.Provider value={props.session || ""}>
        {props.children}
    </SessionContext.Provider>
}