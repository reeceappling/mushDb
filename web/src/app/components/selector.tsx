'use client'
import {JSX, useState} from "react";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import {createUrlFor, Entry} from "@/app/components/common";

export interface SelectorProps<T> {
    doSelect: (val?: T) => void
    allowCreation?: boolean
    headerLevel?: number
    creatorInPage?: boolean
    txt?: string
    hideDisposed?: boolean
}

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
    const optMap = new Map<string, T>(inputs.options.map((v)=>{
        return [inputs.stringFor(v), v]
    }));
    return <select className={"tailwindSelector"} value={""} disabled={false} onChange={(e) => {
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

export default function CloseableSelector<T extends Entry>({props}: {
    props: {
        createSelector: (selectHandler:(onSelect: T)=>void)=>JSX.Element,
        createCreator?: (selectHandler:(onSelect: T)=>void)=>JSX.Element,
        closeTxt: string,
        createTxt?: string,
        createEndpt?: string,
        lowercase: string,
        doSelect: (v: T) => void,
        allowCreation?: boolean,
        creatorInPage?: boolean,
    },
}) {
    const [open, setOpen] = useState(false)
    const [selected, setSelected] = useState<T | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)
    const [creatorOpen, setCreatorOpen] = useState(false)
    // TODO: handle making closeable using escape key!
    // useEffect(() => {
    //     const handleKeyDown = (event: KeyboardEvent) => {
    //         if (event.key === "Escape") {
    //             setOpen(false); // Trigger your close logic here
    //         }
    //     };
    //     document.addEventListener("keydown", handleKeyDown);
    //     return () => {
    //         document.removeEventListener("keydown", handleKeyDown);
    //     };
    // }, []);

    const toggleOpen = () => {
        setOpen(!open)
    }
    const closeButton = <button className={"basicButton"} onClick={() => {
        toggleOpen();
        setCreatorOpen(false)
    }}>{props.closeTxt}</button>
    const deselectButton = <button className={"removeButtonSmall"} onClick={() => {
        setSelected(undefined)
        setCreatorOpen(false)
    }}>{"Clear Selection"}</button>
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
            window.open(createUrlFor(props.createEndpt||"unknown"), '_blank', 'noopener');
            return
        }
        setOpen(false)
        setCreatorOpen(true)
    }
    if (err) {
        return <ErrorDisplay err={err}/>
    }
    const pre = createNewSubArea()
    if (!open) {
        return <div>
            <ErrorDisplay err={err}/>
            <div className={"centerH"}>
                {selected && <div>{selected.getId()}</div>}
                <button className={"basicButton"}
                        onClick={toggleOpen}>{"Select a " + (selected ? "different" : "") + " " + props.lowercase}</button>
                {selected && deselectButton}
            </div>
        </div>
    }
    return <div>
        <ErrorDisplay err={err}/>
        {/* TODO: listen for escape key????? */}
        {closeButton}
        {props.createSelector(selectItem)}
        {pre}
        {closeButton}{selected && deselectButton}
    </div>
}
