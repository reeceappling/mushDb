'use client'
import {ReactNode} from "react";

export default function H({level, offset, children}: { level?: number, offset?: number, children: ReactNode}){
    switch((level || defaultHeaderLevel)-(offset || 0)){
        case 1: return <h1>{children}</h1>
        case 2: return <h2>{children}</h2>
        case 3: return <h3>{children}</h3>
        case 4: return <h4>{children}</h4>
        case 5: return <h5>{children}</h5>
        case 6: return <h6>{children}</h6>
        default:
            console.error("bad header level: "+level)
            return <h1>{children}</h1>
    }
}

export const defaultHeaderLevel = 2