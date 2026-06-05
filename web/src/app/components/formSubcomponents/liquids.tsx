// non-client

import {useQuery} from "@tanstack/react-query";
import {getOptionsResponse} from "./server";
import {SelectorResetsOnSelectFor} from "@/app/components/selector";
import {AreaProps, Data, FormListArea, GroupProps} from "@/app/components/formSubcomponents/shared";
import {LiquidEntryForNew, RemoveButton} from "@/app/components/formSubcomponents/commonClient";
import {NumericalArea} from "@/app/components/formSubcomponents/numericInput";
import * as React from "react";

export interface Liquid {
    name: string,
    pct: number,
}

export const LiquidsList: string[] = ["TapWater", "DistilledWater", "GrainWater"]


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

export function LiquidTypeSelector( // TODO: USE THIS!!!!!
    {current, onSelect, blacklist, readonly}: {
        readonly: boolean,
        current?: string,
        onSelect?: (ab?: string) => void,
        blacklist?: string[],
    }) {
    if (readonly) {
        return current
    }
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
        } else {
            onSelect && onSelect(s as string)
        }
    }}/>
}

export default function LiquidsArea(props: AreaProps<Liquid>) {
    return FormListArea(LiquidEntriesGroup)(props)
}

export function LiquidEntriesGroupForNew({currentEntries, updateParent}: {
    currentEntries: Liquid[],
    updateParent: (l: Liquid[]) => void
}) {
    const handleSelectLiquid = (liq: string) => {
        updateParent([...(currentEntries || []), {name: liq, pct: 0}])
    }
    return <div>
        {currentEntries.length!==0 && <div className={"inputGrid inputGrid3 gap-8"}>
        {currentEntries.map((l, i) => {
            return <div key={l.name} className={"contentsOnly"}>{/*<div className={"flex my-4 text-m"}><div className={"InputGrid InputGrid4"}><div className={"inlineChildren mb-1"}>*/}
                <LiquidEntryForNew currentValue={l} updateParent={(l: Liquid) => {
                    updateParent([...(currentEntries || [])].map((existingLiquid) => {
                        return existingLiquid.name !== l.name ? existingLiquid : l
                    }))
                }}/>
                <RemoveButton txt={"Remove"} click={()=>{
                    updateParent([...(currentEntries || [])].filter((existing) => existing.name !== l.name))
                }} />
            </div>
        })}
        </div>}
        <LiquidsTypeSelectorForNew onSelect={(liq) => {
            liq && handleSelectLiquid(liq)
        }} blacklist={currentEntries.map((v) => {
            return v.name
        })}/>
    </div>
}

// TODO: FIX
// TODO: NOT PROPERLY WORKING
export function LiquidEntriesGroup({
                                       initialEntries,
                                       preexisting,
                                       readonly,
                                       updateParent,
                                       blacklist
                                   }: GroupProps<Liquid>) {
    const handleFormChangeName = (index: number, liq: string) => {
        const data = [...(initialEntries || [])];
        data[index].data.name = liq
        updateParent(data)
    }
    const handleFormChangePct = (index: number, val: number) => {
        // TODO: HANDLE NON-NUMBERS!!!
        const data = [...(initialEntries || [])];
        data[index].data.pct = val
        updateParent(data)
    }
    const addFields = (e: React.MouseEvent) => {
        e.preventDefault()
        const data = [...(initialEntries || []), {data: {pct: 0.0, name: ''}, disabled: false}]
        updateParent(data)
    }
    const removeFields = (index: number) => {
        const data = [...(initialEntries || [])];
        data.splice(index, 1) // TODO: THIS WONT WORK PROPERLY WITH INDEX
        updateParent(data)
    }
    const disableField = (index: number) => {
        const data = [...(initialEntries || [])]
        data[index].disabled = !data[index].disabled
        updateParent(data)
    }
    const groupClasses = () => {
        let out = ""
        if (preexisting) {
            out = "exists"
        } else {
            out = "new"
        }
        if (readonly) {
            out += " readonly"
        } else {
            out += " editable"
        }
        return out
    }
    const entryClasses = (note: Data<Liquid>) => {
        let out = "liquidEntry"
        if (note.disabled) {
            out += " disabled"
        } else {
            out += " enabled"
        }
        return out
    }
    const valuesInUse = () => {
        return (initialEntries || []).map((v) => {
            return v.data.name
        })
    }
    // TODO: match this up with AdditiveEntriesGroup
    return <div className={groupClasses()}>{/* TODO: CLASS STYLINGS!!!! */}
        {(initialEntries || []).map((input, index) => {
            return (
                <div className={"editable "+entryClasses(input)} key={index}> {/* TODO: CLASS STYLINGS!!!! */}
                    {input.disabled ? "disabled" : null /* TODO: remove? */}
                    <LiquidTypeSelector current={input.data.name} onSelect={(liq) => {
                        liq && handleFormChangeName(index, liq)
                    }} blacklist={valuesInUse().filter((v) => {
                        return v !== input.data.name
                    })} readonly={readonly}/>
                    {" volume: "}{readonly ? input.data.pct.toString() + "%" :
                        <NumericalArea label="Percentage by volume" mode="floating" min={0.0} max={1.0}
                                       readonly={readonly} errorMessage={"FIXME"} value={input.data.pct.toString()}
                                       onChange={(val?: string) => {
                                           val && handleFormChangePct(index, Number(val))
                                       }}/>}
                    {/* TODO: VALIDATE PERCENTAGES? */}
                    {readonly ? null :
                        <button onClick={() => {
                            preexisting ? disableField(index) : removeFields(index)
                        }}>{preexisting ? (input.disabled ? "enable" : "disable") : "remove"}</button>}
                </div>)
        })}
        {preexisting ? null : <button className={"basicButton"} onClick={addFields}>{"Add More"}</button>}
        {/* TODO: validate that the editing version of this works*/}
    </div>
}

