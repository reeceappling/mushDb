// non-client

import {useQuery} from "@tanstack/react-query";
import {getOptionsResponse} from "./server";
import {SelectorResetsOnSelectFor} from "@/app/components/selector";
import {LiquidEntryForNew, RemoveButton} from "@/app/components/formSubcomponents/commonClient";
import * as React from "react";
import {useEffect, useState} from "react";

export interface Liquid {
    name: string,
    pct: number,
}

export function IsValidLiquid(input: any): boolean {
    return (
        typeof input === 'object' &&
        'name' in input && typeof input.name === 'string' &&
        'pct' in input && typeof input.pct === 'number'
    )
}

export function LiquidsTypeSelectorForNew(
    {onSelect, blacklist}: {
        onSelect?: (ab?: string) => void,
        blacklist?: string[],
    }) {
    const {isPending, error, data} = useQuery({
        queryKey: ['liquidsOptions'],
        queryFn: () => getOptionsResponse("liquids")
    })
    if (isPending || error !== null) {
        return <div>{isPending ? "LIQUID SELECTOR LOADING" : "LIQUID SELECTOR ERROR: " + error.message}</div>
    }
    const filteredOptions = data.filter((val, idx) => {
        return !(blacklist && blacklist.includes(val))
    })
    return <SelectorResetsOnSelectFor options={["", ...filteredOptions]} updateParent={(s) => {
        if (s === "") {
            onSelect && onSelect()
        }
        onSelect && onSelect(s as string)
    }
    }/>
}

export function LiquidsAreaReadOnly({values}: {values?:Liquid[]}) {
    if (!values || values.length===0){
        return null
    }
    return <div>
        {"Liquids: "}
        {values.map((v, i) => {
            return <div key={v.name}>{v.name + " " + v.pct + " %"}</div>
        })}
    </div>
}

interface LiquidWithInitial {
    name: string,
    pct: number,
    start: number,
}

export function LiquidEntriesGroupForNew({initial, updateParent}: {
    initial?: Liquid[],
    updateParent: (l: Liquid[]) => void
}) {
    const [current, setCurrent] = useState<LiquidWithInitial[]>((initial||[]).map(l=>{
        return {...l, start:l.pct}
    }))
    useEffect(() => {
        setCurrent((initial||[]).map(l=>{
            return {...l, start:l.pct}
        }))
    },[initial])
    const volUnused = ()=>{
        let out = 100
        current.forEach(liq=>{
            out-=liq.pct
        })
        return out
    }
    const doUpdate = (upd:LiquidWithInitial[]) => {
        setCurrent(upd)
        updateParent(upd)
    }
    const handleSelectLiquid = (liq: string) => {
        const unused = volUnused()
        const upd = [...(structuredClone(current) || []), {name: liq, pct: unused, start: unused}]
        setCurrent(upd)
        updateParent(upd)
    }
    return <div>
        {current.length!==0 && <div className={"inputGrid inputGrid3 gap-8"}>
        {current.map((l, i) => {
            return <div key={l.name} className={"contentsOnly"}>{/*<div className={"flex my-4 text-m"}><div className={"InputGrid InputGrid4"}><div className={"inlineChildren mb-1"}>*/}
                <LiquidEntryForNew initial={{name:l.name,pct:(initial && initial.length>i)?initial[i].pct:l.start}} updateParent={(l: Liquid) => {
                    doUpdate([...(current || [])].map((existingLiquid) => {
                        return existingLiquid.name !== l.name ? existingLiquid : {...l, start: existingLiquid.start}
                    }))
                }}/>
                <RemoveButton txt={"Remove"} click={()=>{
                    doUpdate([...(current || [])].filter((existing,ind) => existing.name !== l.name))
                }} />
            </div>
        })}
        </div>}
        <LiquidsTypeSelectorForNew onSelect={(liq) => {
            liq && handleSelectLiquid(liq)
        }} blacklist={current.map((v) => {
            return v.name
        })}/>
    </div>
}

