// no state in here!

import {AreaProps, Data, FormListArea, GroupProps} from "@/app/components/formSubcomponents/shared";
import {useQuery} from "@tanstack/react-query";
import {SelectorFor, SelectorResetsOnSelectFor} from "@/app/components/selector";
import {NumericalArea} from "@/app/components/formSubcomponents/numericInput";
import TextBox from "@/app/components/formSubcomponents/textbox";
import {getOptionsResponse} from "@/app/components/formSubcomponents/server";
import TestAndValidate from "@/app/components/testing/untested";
import {AdditiveEntryForNew, RemoveButton} from "@/app/components/formSubcomponents/commonClient";
import * as React from "react";
import {useEffect, useState} from "react";
import {Sugar} from "@/app/components/formSubcomponents/sugars";

export interface Additive {
    additive: string
    amount: number
    unit: string
}

export const AdditivesList: string[] = ["Add1", "Add2", "Add3"] // TODO: swap out with actual additives

export function IsValidAdditive(input: any): boolean {
    return (
        typeof input === 'object' &&
        'additive' in input && typeof input.additive === 'string' &&
        'amount' in input && typeof input.amount === 'number' &&
        'unit' in input && typeof input.unit === 'string'
    )
}

export function AdditiveTypeSelectorForNew(
    {onSelect, blacklist}: {
        onSelect?: (ab?: string) => void,
        blacklist?: string[],
    }) {
    const {isPending, error, data} = useQuery({
        queryKey: ['additivesOptions'],
        queryFn: () => getOptionsResponse("additives")
    })
    if (isPending || error !== null) {
        return <div>{isPending ? "ADDITIVE SELECTOR LOADING" : "ADDITIVE SELECTOR ERROR: " + error.message}</div>
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

export function AdditiveSelector(
    {initial, onSelect, blacklist}: {
        initial?: string,
        onSelect?: (ab?: string) => void,
        blacklist?: string[],
    }) {
    const {isPending, error, data} = useQuery({
        queryKey: ['additivesOptions'],
        queryFn: () => getOptionsResponse("additives")
    })
    if (isPending || error !== null) {
        return <div>{isPending ? "ADDITIVE SELECTOR LOADING" : "ADDITIVE SELECTOR ERROR: " + error.message}</div>
    }
    const filteredOptions = data.filter((val, idx) => {
        return !(blacklist && blacklist.includes(val))
    })
    return <SelectorFor disabled={onSelect === undefined} options={["", ...filteredOptions]} initial={initial || ""}
                        updateParent={(s) => {
                            if (s === "") {
                                onSelect && onSelect(undefined)
                            }
                            onSelect && onSelect(s as string)
                        }
                        }/>
}

export function AdditivesAreaReadOnly({values}: {values?:Additive[]}) {
    if (!values || values.length===0){
        return null
    }
    return <div>
        {"Additives: "}
        {values.map((v, i) => {
            return <div key={v.additive}>{v.additive + " - " + v.amount + " " + v.unit}</div>
        })}
    </div>
}

export function AdditiveEntriesGroupForNew({initial, updateParent}: {
    initial: Additive[],
    updateParent: (l: Additive[]) => void
}) {
    const [current, setCurrent] = useState<Additive[]>(initial);
    useEffect(()=>{
        setCurrent(initial)
    },[initial])
    const handleSelect = (v: string) => {
        const data = [...(current || []), {additive: v, amount: 0, unit: ""}];
        setCurrent(data)
    }
    const doUpdate = (upd:Additive[]) => {
        setCurrent(upd)
        updateParent(upd)
    }
    return <div>
        {current.length !== 0 && <div className={"inputGrid inputGrid4 gap-8"}>
            {current.map((n, i) => {
                return <div key={n.additive} className={"contentsOnly"}>
                    <AdditiveEntryForNew initial={{
                        additive:n.additive,
                        amount: initial.length>i?initial[i].amount:1.0, // TODO: ok?
                        unit: initial.length>i?initial[i].unit:"g", // TODO: ok?
                    }} updateParent={(updated: Additive) => {
                        doUpdate([...(current || [])].map((existing) => {
                            return existing.additive !== n.additive ? existing : updated
                        }))
                    }}/>
                    <RemoveButton txt={"Remove"} click={() => {
                        doUpdate([...(current || [])].filter((existing) => existing.additive !== n.additive))
                    }}/>
                </div>
            })}
        </div>}
        <AdditiveTypeSelectorForNew onSelect={(val) => {
            val && handleSelect(val)
        }} blacklist={current.map((v) => {
            return v.additive
        })}/>
    </div>
}
