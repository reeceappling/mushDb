'use client'
import {JSX, useEffect, useState} from "react";
import {BaseExternalUrl} from "@/app/components/Constants";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import {createUrlFor, HandleJsonResponse, InlineProps} from "@/app/components/common";
import {TestAgarBatchOk} from "@/app/components/agarBatchServer";
import {TestFruitOK} from "@/app/components/fruitServer";
import {TestJarOK} from "@/app/components/jarServer";
import {TestPcRunOk} from "@/app/components/pcRunServer";
import {TestProjectOk} from "@/app/components/projectServer";
import {TestSaleOk} from "@/app/components/saleServer";
import {useCookies} from "react-cookie";
import {createSelector} from "reselect";

export interface SelectorProps<T> {
    doSelect: (val?: T) => void
    allowCreation?: boolean
    headerLevel?: number
    creatorInPage?: boolean
    txt?: string
}

// TODO: MAKE SURE SELECTOR DISPLAY VALUES PROPERLY DISPLAYS BASE58S WHEN NEEDED
export function SelectorFor(
    inputs: {
        options: string[],
        initial: string,
        updateParent: (str: string) => void,
        disabled: boolean
    }) {
    if(inputs.disabled){
        return <text>{inputs.initial}</text>
    }
    return <select key={inputs.initial} className={"tailwindSelector"} defaultValue={inputs.initial} disabled={inputs.disabled}
                   onChange={(e) => {
                       inputs.updateParent(e.currentTarget.value)
                   }}>
        {inputs.options.map((s, i: number) => {
            return <option value={s} key={s+i}>{s}</option>
        })}
    </select>
}

// TODO: MAKE SURE SELECTOR DISPLAY VALUES PROPERLY DISPLAYS BASE58S WHEN NEEDED
export function SelectorResetsOnSelectFor(
    inputs: {
        options: string[],
        updateParent: (str: string) => void,
    }) {
    if (inputs.options.length === 1) {
        return null
    }
    return <select className={"tailwindSelector"} value={inputs.options[0]} disabled={false}
                   onChange={(e) => {
                       inputs.updateParent(e.target.value)
                   }}>
        {inputs.options.map((s, i: number) => {
            return <option value={s} key={i}>{s}</option>
        })}
    </select>
}

export function DefaultOption(){
    return <option value={""} onClick={()=> {}} onSelect={()=>{}}>{""}</option>
}

export function SelectorResetsOnSelectForCustom<T>(
    inputs: {
        options: T[],
        updateParent: (val: T) => void,
        stringFor: (val: T) => string,
    }) {
    if (inputs.options.length === 0) {
        return null
    }
    console.log("length of options is now: "+inputs.options.length) // TODO: del
    const optMap = new Map<string, T>(inputs.options.map((v)=>{
        return [inputs.stringFor(v), v]
    }));
    return <select className={"tailwindSelector"} value={""} disabled={false} onChange={(e) => {
        console.log("user option selected: "+e.target.value) // TODO: del
        if(e.target.value===""){
            return
        }
        const updated = optMap.get(e.target.value)
        if (updated){
            inputs.updateParent(updated)
        } else {
            console.error(e.target.value+" did not exist in the optMap!") // TODO: del!
        }
    }}>
        <DefaultOption />
        {inputs.options
                .filter((o)=>{return inputs.stringFor(o)!==""})
                .map((s, i: number) => {
                    const str = inputs.stringFor(s)
                    return <option value={str} key={i}>{str}</option>
                })
        }
    </select>
}

// // TODO: MAKE SURE SELECTOR DISPLAY VALUES PROPERLY DISPLAYS BASE58S WHEN NEEDED
// export default function RecentSelector<T>({props, children}: { // TODO: FIX FOR PERMISSIONED ONES?
//     props: {
//         msgTxt: string,
//         recentEndpt: string,
//         assertType: (atIn: any) => void,
//         closeTxt: string,
//         createTxt?: string,
//         createEndpt: string,
//         lowercase: string,
//         inline: (inlineIn: InlineProps<T>) => JSX.Element,
//         getId: (v: T) => string, // TODO: CHANGE THIS ON ALL
//         doSelect: (v: T) => void,
//         allowCreation?: boolean,
//         creatorInPage?: boolean,
//     },
//     children: React.ReactNode
// }) {
//     // TODO: do selectors need incremented depth?
//     const [reload, setReload] = useState(false)
//     const doReload = () => {
//         setReload(!reload)
//     }
//
//     const [loaded, setLoaded] = useState(false)
//     const [open, setOpen] = useState(false)
//     const [selectable, setSelectable] = useState<T[]>([])
//     const [selected, setSelected] = useState<T | undefined>(undefined)
//     const [err, setErr] = useState<string | undefined>(undefined)
//     const [creatorOpen, setCreatorOpen] = useState(false)
//     ////const [cookies, setCookie, removeCookie] = useCookies(['SessionId']);
//     useEffect(() => {
//         switch (props.recentEndpt) { // TODO: GET RID OF! TESTS ONLY!
//             case "agarBatches":
//                 setSelectable([TestAgarBatchOk(), TestAgarBatchOk(), TestAgarBatchOk()] as T[])
//                 break
//             case "fruits":
//                 setSelectable([TestFruitOK(), TestFruitOK(), TestFruitOK()] as T[])
//                 break;
//             case "jars":
//                 setSelectable([TestJarOK(), TestJarOK(), TestJarOK()] as T[])
//                 break;
//             case "pcRuns":
//                 setSelectable([TestPcRunOk(), TestPcRunOk(), TestPcRunOk()] as T[])
//                 break;
//             case "projects":
//                 setSelectable([TestProjectOk(), TestProjectOk(), TestProjectOk()] as T[])
//                 break;
//             case "sales":
//                 setSelectable([TestSaleOk(), TestSaleOk(), TestSaleOk()] as T[])
//                 break;
//             default:
//                 setErr("bad recentEndpt: " + props.recentEndpt)
//                 break;
//         }
//         setLoaded(true)
//         return
//         fetch(BaseExternalUrl + "/db/recent/" + props.recentEndpt, { // TODO: ensure correct
//             method: "GET",
//             headers: clientPostRequestHeaders,
//         }).then(HandleJsonResponse)
//             .then((data) => {
//                 Array.isArray(data) && data.every(props.assertType)
//                 setSelectable(data as T[])
//                 setLoaded(true)
//             })
//             .catch(ErrHandler(setErr));
//     }, [reload]);
//     //const ch = CreateChannel()
//     // ch.onmessage = (event) => {
//     //     try {
//     //         if (event.data as string === msgTxt) {
//     //             doReload()
//     //         }
//     //     } catch {
//     //         console.error("failed to decode event: " + event.data)
//     //     }
//     //
//     // };
//
//     const toggleOpen = () => {
//         setOpen(!open)
//     }
//     const closeButton = <button className={"basicButton"} onClick={() => {
//         toggleOpen();
//         setCreatorOpen(false)
//     }}>{props.closeTxt}</button>
//     const selectItem = (item: T) => {
//         props.doSelect(item)
//         setSelected(item)
//         setOpen(false)
//     }
//     const createNewSubArea = () => {
//         if (!creatorOpen) {
//             if (props.createTxt) {
//                 return <div className={"centerH gapTop"}>
//                     <button className={"basicButton"} onClick={openCreateNew}>{props.createTxt}</button>
//                 </div>
//             }
//             return null
//         }
//         return <div className={"centerH subFormCreator gapTop"}>
//             <div className={"gapTop"}>{children}</div>
//             <div>
//                 <button className={"basicButton"} onClick={() => {
//                     setCreatorOpen(false)
//                 }}>{"Close This Creator"}</button>
//             </div>
//
//         </div> // TODO: NEEDS SPECIAL STYLING TO ENSURE WE KNOW WHICH FORM IS WHICH INTERNALLY
//
//     }
//     const openCreateNew = (e: React.MouseEvent) => {
//         e.preventDefault()
//         if (!props.creatorInPage) {
//             console.log("CREATOR NOT IN PAGE")
//             window.open(BaseExternalUrl + "/new/" + props.createEndpt, '_blank', 'noopener'); // TODO: ensure ok
//             return
//         }
//         setOpen(false)
//         setCreatorOpen(true)
//     }
//     if (err) {
//         return <ErrorDisplay err={err}/>
//     }
//     let pre = createNewSubArea()
//     if (!open) {
//         return <div>
//             <ErrorDisplay err={err}/>
//             <div className={"centerH"}>
//                 {selected && <div>{props.getId(selected)}</div>}
//                 <button className={"basicButton"}
//                     onClick={toggleOpen}>{"Select a " + (selected ? "different" : "recent") + " " + props.lowercase}</button>
//             </div>
//             {pre}
//         </div>
//     }
//     if (!loaded) {
//         return <div>
//             <ErrorDisplay err={err}/>
//             <div>{"Loading..."}</div>
//         </div>
//     }
//     return <div> {/* TODO: can we do this in the modal????? Div might be weird here*/}
//         <ErrorDisplay err={err}/>
//         {/* TODO: listen for escape key????? */}
//         {closeButton}
//         {selectable.map((opt, i) => {
//             // TODO: HIGHLIGHT CURRENTLY SELECTED!!!!!!!
//             return <div key={i}>
//                 {props.inline({
//                     data: opt, onClick: () => {
//                         selectItem(opt)
//                     }
//                 })}
//             </div>
//         })}
//         {pre}
//         {closeButton}
//     </div>
// }

export default function CloseableSelector<T>({props}: { // TODO: FIX FOR PERMISSIONED ONES?
    props: {
        msgTxt: string,// TODO: del?
        createSelector: (selectHandler:(onSelect: T)=>void)=>JSX.Element,
        createCreator?: (selectHandler:(onSelect: T)=>void)=>JSX.Element,
        closeTxt: string,
        createTxt?: string,
        createEndpt?: string,
        lowercase: string,
        getId: (v: T) => string, // TODO: CHANGE THIS ON ALL
        doSelect: (v: T) => void,
        allowCreation?: boolean,
        creatorInPage?: boolean,
    },
}) {
    const [open, setOpen] = useState(false)
    const [selected, setSelected] = useState<T | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)
    const [creatorOpen, setCreatorOpen] = useState(false)

    const toggleOpen = () => {
        setOpen(!open)
    }
    const closeButton = <button className={"basicButton"} onClick={() => {
        toggleOpen();
        setCreatorOpen(false)
    }}>{props.closeTxt}</button>
    const selectItem = (item: T) => {
        props.doSelect(item)
        setSelected(item)
        setOpen(false)
    }
    const creator = props.createCreator? props.createCreator(selectItem): null
    const createNewSubArea = () => {
        if (!creatorOpen) {
            if (props.createTxt) {
                return <div className={"centerH gapTop"}>
                    <button className={"basicButton"} onClick={openCreateNew}>{props.createTxt}</button>
                </div>
            }
            return null
        }
        return <div className={"centerH subFormCreator gapTop"}>
            <div className={"gapTop"}>{creator}</div>
            <div>
                <button className={"basicButton"} onClick={() => {
                    setCreatorOpen(false)
                }}>{"Close This Creator"}</button>

            </div>

        </div>

    }
    const openCreateNew = (e: React.MouseEvent) => {
        e.preventDefault()
        if (!props.creatorInPage) {
            window.open(createUrlFor(props.createEndpt||"unknown"), '_blank', 'noopener'); // TODO: ensure ok
            return
        }
        setOpen(false)
        setCreatorOpen(true)
    }
    if (err) {
        return <ErrorDisplay err={err}/>
    }
    let pre = createNewSubArea()
    if (!open) {
        return <div>
            <ErrorDisplay err={err}/>
            <div className={"centerH"}>
                {selected && <div>{props.getId(selected)}</div>}
                <button className={"basicButton"}
                        onClick={toggleOpen}>{"Select a " + (selected ? "different" : "") + " " + props.lowercase}</button>
            </div>
        </div>
    }
    return <div>
        <ErrorDisplay err={err}/>
        {/* TODO: listen for escape key????? */}
        {closeButton}
        {props.createSelector(selectItem)}
        {pre}
        {closeButton}
    </div>
}

// export default function RecentSelector<T>(
//     {msgTxt, recentEndpt, assertType, closeTxt, createTxt, newForm, createEndpt, lowercase, inline,getId}:
//         {
//             msgTxt: string,
//             recentEndpt: string,
//             assertType: (atIn: any)=>void,
//             closeTxt: string,
//             createTxt?: string,
//             newForm?: JSX.Element,
//             createEndpt: string,
//             lowercase: string,
//             inline: (inlineIn: InlineProps<T>)=>JSX.Element,
//             getId: (v: T)=>string // TODO: CHANGE THIS ON ALL
//         }
// ): ((outProps: SelectorProps<T>) => JSX.Element) {
//     return function ({doSelect, allowCreation, headerLevel, creatorInPage}: SelectorProps<T>) {
//         const [reload, setReload] = useState(false)
//         const doReload = () => {
//             setReload(!reload)
//         }
//
//         const [loaded, setLoaded] = useState(false)
//         const [open, setOpen] = useState(false)
//         const [selectable, setSelectable] = useState<T[]>([])
//         const [selected, setSelected] = useState<T | undefined>(undefined)
//         const [err, setErr] = useState<string | undefined>(undefined)
//         const [creatorOpen, setCreatorOpen] = useState(false)
//         useEffect(() => {
//             switch(recentEndpt){ // TODO: GET RID OF! TESTS ONLY!
//                 case "agarBatches":
//                     setSelectable([TestAgarBatchOk(),TestAgarBatchOk(),TestAgarBatchOk()] as T[])
//                     break
//                 case "fruits":
//                     setSelectable([TestFruitOK(),TestFruitOK(),TestFruitOK()] as T[])
//                     break;
//                 case "jars":
//                     setSelectable([TestJarOK(),TestJarOK(),TestJarOK()] as T[])
//                     break;
//                 case "pcRuns":
//                     setSelectable([TestPcRunOk(),TestPcRunOk(),TestPcRunOk()] as T[])
//                     break;
//                 case "projects":
//                     setSelectable([TestProjectOk(),TestProjectOk(),TestProjectOk()] as T[])
//                     break;
//                 case "sales":
//                     setSelectable([TestSaleOk(),TestSaleOk(),TestSaleOk()] as T[])
//                     break;
//                 default:
//                     setErr("bad recentEndpt: "+recentEndpt)
//                     break;
//             }
//             setLoaded(true)
//             return
//             fetch(BaseExternalUrl+"/db/recent/"+recentEndpt, { // TODO: ensure correct
//                 method: "GET",
//                 headers: clientPostRequestHeaders,
//             })
//                 .then(HandleJsonResponse)
//                 .then((data) => {
//                     Array.isArray(data) && data.every(assertType)
//                     setSelectable(data as T[])
//                     setLoaded(true)
//                 })
//                 .catch(ErrHandler(setErr));
//         }, [reload]);
//         //const ch = CreateChannel()
//         // ch.onmessage = (event) => {
//         //     try {
//         //         if (event.data as string === msgTxt) {
//         //             doReload()
//         //         }
//         //     } catch {
//         //         console.error("failed to decode event: " + event.data)
//         //     }
//         //
//         // };
//
//         const toggleOpen = () => {
//             setOpen(!open)
//         }
//         const closeButton = <button onClick={() => {
//             toggleOpen();
//             setCreatorOpen(false)
//         }}>{closeTxt}</button>
//         const selectItem = (item: T) => {
//             doSelect(item)
//             setSelected(item)
//             setOpen(false)
//         }
//         const createNewSubArea = () => {
//             if(!allowCreation){
//                 return null
//             }
//             if (!creatorOpen) {
//                 if(createTxt){
//                     return <div className={"centerH"}><button onClick={openCreateNew}>{createTxt}</button></div>
//                 }
//                 return null
//             }
//             if(newForm!==undefined) {
//                 return <div className={"centerH"}>
//                     {newForm}
//                 </div>
//             }
//             return null
//         }
//         const openCreateNew = (e: React.MouseEvent) => {
//             e.preventDefault()
//             if (!creatorInPage) {
//                 window.open(BaseExternalUrl+"/new/"+createEndpt, '_blank', 'noopener'); // TODO: ensure ok
//                 return
//             }
//             setOpen(false)
//             setCreatorOpen(true)
//         }
//         if(err){
//             return <ErrorDisplay err={err}/>
//         }
//         let pre = createNewSubArea()
//         if (!open) {
//             return <div>
//                 <ErrorDisplay err={err} headerLevel={headerLevel}/>
//                 <div className={"centerH"}>
//                     {selected && <div>{getId(selected)}</div>}
//                     <button onClick={toggleOpen}>{"Select a "+(selected?"different":"recent")+" "+lowercase}</button>
//                 </div>
//                 {pre}
//             </div>
//         }
//         if (!loaded) {
//             return <div>
//                 <ErrorDisplay err={err} headerLevel={headerLevel}/>
//                 <div>{"Loading..."}</div>
//             </div>
//         }
//         return <div> {/* TODO: can we do this in the modal????? Div might be weird here*/}
//             <ErrorDisplay err={err} headerLevel={headerLevel}/>
//             {/* TODO: listen for escape key????? */}
//             {closeButton}
//             {selectable.map((opt, i) => {
//                 // TODO: HIGHLIGHT CURRENTLY SELECTED!!!!!!!
//                 return <div key={i}>
//                     {inline({data:opt, onClick:() => {selectItem(opt)}, headerLevel:headerLevel})}
//                 </div>
//             })}
//             {pre}
//             {closeButton}
//         </div>
//     }
// }