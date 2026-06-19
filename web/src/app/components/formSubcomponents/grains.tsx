// non-client

import {useQuery} from "@tanstack/react-query";
import {SelectorFor} from "@/app/components/selector";
import {getOptionsResponse} from "@/app/components/formSubcomponents/server";
import {
    NumericalArea,
    NumericalAreaWithAbsolutes
} from "@/app/components/formSubcomponents/numericInput";
import {RemoveButton} from "@/app/components/formSubcomponents/commonClient";
import * as React from "react";
import {useEffect, useState} from "react";

export const GrainsList: string[] = ["Oats", "Popcorn", "Wheat", "Rye", "Millett"]



export function GrainsTypeSelector( // TODO: USE THIS!!!!!
    {initial, onSelect,blacklist}:{
        initial?: string,
        onSelect?: (ab?: string)=>void
        blacklist?: string[],
    }){
    const { isPending, error, data } = useQuery({
        queryKey: ['grainsOptions'],
        queryFn: () => getOptionsResponse("grains")
    })
    if(isPending || error !== null){
        return <div>{isPending ? "GRAIN SELECTOR LOADING" : "GRAIN SELECTOR ERROR: "+error.message}</div>
    }
    const filteredOptions = data.filter((val, idx)=>{
        return !(blacklist && blacklist.includes(val))
    })
    return <SelectorFor disabled={onSelect===undefined} options={["", ...filteredOptions]} initial={initial || ""} updateParent={(s)=>{
        if(s===""){
            onSelect && onSelect()
        }
        onSelect && onSelect(s)}
    } />


}

export function GrainsEntriesGroupForNew({initial, updateParent}: {initial: Grain[], updateParent: (l: Grain[])=>void}){
    const [current, setCurrent] = useState<Grain[]>(initial);
    useEffect(()=>{
        setCurrent(initial)
    },[initial])
    const handleSelectGrain = (v: string) => {
        const data = [...(current || []), {grain: v, percentage: 0}];
        setCurrent(data)
    }
    const doUpdate = (upd:Grain[]) => {
        setCurrent(upd)
        updateParent(upd)
    }
    return <div>
        {current.length!==0 && <div className={"inputGrid inputGrid3 gap-8"}>
            {current.map((v,i)=>{
                return <div key={v.grain} className={"contentsOnly"}>
                    <GrainEntryForNew currentValue={v} updateParent={(updated: Grain) => {
                        doUpdate([...(current || [])].map((existing) => {
                            return existing.grain !== v.grain ? existing : updated
                        }))
                    }}/>
                    <RemoveButton txt={"Remove"} click={()=>{
                        doUpdate([...(current || [])].filter((existing) => existing.grain !== v.grain))
                    }} />
                </div>
            })}
        </div>}
        <GrainsTypeSelector onSelect={v=>{v&&handleSelectGrain(v)}} blacklist={current.map((v)=>{return v.grain})} />
    </div>
}

export function GrainEntryForNew({currentValue, updateParent}: {
    currentValue: Grain,
    updateParent: (l: Grain) => void
}) {
    const [err, setErr] = useState<string | undefined>()
    const handleFormChangeAmt = (val: number) => {
        const data = {...currentValue};
        data.percentage = val
        updateParent(data)
    }
    return <>
        <div className={"text-m"}>{currentValue.grain}</div>
        <NumericalAreaWithAbsolutes label="Percentage" mode="floating" min={0.0} max={1.0} readonly={false}
                                    errorMessage={err} value={currentValue.percentage.toString()}
                                    onChange={(val?: string) => {
                                        try {
                                            const n = Number(val)
                                            if (Number.isNaN(n)) {
                                                setErr("NaN input")
                                            } else {
                                                val && handleFormChangeAmt(n)
                                                setErr(undefined)
                                            }
                                        } catch (e) {
                                            setErr(JSON.stringify(e))
                                        }
                                    }}/>
    </>
}

export function GrainsSelector({current, onChange}:{current: Grain[], onChange: (gs: Grain[])=>void}){
    const grainTypeAmountSelectors = ()=>{
        return <div>
            {current.map((g, idx)=>{
                return <div key={g.grain} className={"inlineChildren"}>
                    <div>{g.grain}</div>
                    <NumericalArea min={0.0} max={1000.0} placeholder={"%"} mode={'floating'} onChange={(s?: string)=>{
                        if(!s){
                            return
                        }
                        const newNum = Number(s)
                        const updated = structuredClone(current)
                        updated[idx].percentage = newNum
                        onChange(updated)
                    }}/>{/* TODO: ENSURE WORKS */}
                    <button className={"removeButton"} onClick={e=>{
                        e.stopPropagation();
                        onChange(structuredClone(current).filter((val, i)=>i!==idx)) // TODO: ensure works!
                    }}></button>
                </div>
            })}
        </div>
    }
    const addGrain = (g?: string) => {
        if(g){
            onChange([...current, {grain: g, percentage: 0}])
        }
    }
    return <div>
        {grainTypeAmountSelectors()}
        <GrainsTypeSelector onSelect={addGrain}/>
        {"FIXME!!!!!"}
    </div>
}

export interface Grain {
    grain: string,
    percentage: number,
}

export function IsValidGrain(input: any): boolean {
    return (
        typeof input === 'object' &&
        'grain' in input && typeof input.grain === 'string' &&
        'percentage' in input && typeof input.percentage === 'number'
    )
}