import {createContext, useContext} from "react";

export const DepthContext = createContext(0)
export function DepthProvider(props:React.PropsWithChildren<{}>){
    const depth = useContext(DepthContext)
    const newDepth = depth + 1
    return <DepthContext.Provider value={newDepth}>
        {/*TODO: working<TestAndValidate todos={["validate depths"]}>{"Depth: "+depth}</TestAndValidate>*/}
        {props.children}
    </DepthContext.Provider>
}
export function Subsection(props:React.PropsWithChildren<{}>){
    return <div className={"depthSubsection"}>
        {props.children}
    </div>
}