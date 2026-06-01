'use client'

import {createContext, useContext} from "react";
import {RequestCookie} from "next/dist/compiled/@edge-runtime/cookies";
import {SessionProvider} from "@/app/components/formSubcomponents/sessionContext/session";

export const CookiesContext = createContext<RequestCookie[]>([])
export function allCookies(cookiesFromContext: RequestCookie[]):string{ // TODO: USE THIS!
    return cookiesFromContext.map(cookie => `${cookie.name}=${cookie.value}`).join('; ');
}
export function CookiesProvider(props:React.PropsWithChildren<{cookies?:RequestCookie[],session?:string}>){
    // Use with: const cookies = useContext(CookiesContext)
    // Use with: const session = useContext(SessionContext)
    return <CookiesContext.Provider value={props.cookies || []}>
        <SessionProvider session={props.session}>
            {props.children}
        </SessionProvider>
    </CookiesContext.Provider>
}