// no state in here!

import {AreaProps, Data, FormListArea, GroupProps} from "@/app/components/formSubcomponents/shared";
import {useQuery} from "@tanstack/react-query";
import {SelectorFor, SelectorResetsOnSelectFor} from "@/app/components/selector";
import {NumericalArea} from "@/app/components/formSubcomponents/numericInput";
import TextBox from "@/app/components/formSubcomponents/textbox";
import {getOptionsResponse} from "@/app/components/formSubcomponents/server";
import TestAndValidate from "@/app/components/testing/untested";
import {AdditiveEntryForNew, RemoveButton, SugarEntryForNew} from "@/app/components/formSubcomponents/commonClient";
import * as React from "react";
import {useEffect, useState} from "react";
import {Sugar, SugarTypeSelectorForNew} from "@/app/components/formSubcomponents/sugars";

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
export function AdditiveEntriesGroupForNew({
                                            initial,
                                            updateParent,
                                        }: {
    initial: Additive[],
    updateParent: (l: Additive[]) => void
}) {
    const [current, setCurrent] = useState<Additive[]>(initial)

    useEffect(() => {
        setCurrent(initial)
    }, [initial])

    const doUpdate = (upd: Additive[]) => {
        setCurrent(upd)
        updateParent(upd)
    }

    const handleSelectType = (v: string) => {
        const data = [...current, { additive: v, amount: 1, unit: "g" }] // TODO: MODIFY ON OTHERS!
        doUpdate(data)
    }

    return <div>
        {current.length !== 0 && <div className={"inputGrid inputGrid4 gap-8"}>
            {current.map((n, i) => {
                return <div key={n.additive} className={"contentsOnly"}>
                    <AdditiveEntryForNew
                        initial={n}
                        updateParent={(updated: Additive) => {
                            doUpdate(current.map((existing, idx) => idx === i ? updated : existing))
                        }}
                    />
                    <RemoveButton
                        txt={"Remove"}
                        click={() => {
                            doUpdate(current.filter((_, idx) => idx !== i))
                        }}
                    />
                </div>
            })}
        </div>}
        <AdditiveTypeSelectorForNew
            onSelect={(val) => { if (val) handleSelectType(val) }}
            blacklist={current.map((v) => v.additive)}
        />
    </div>
}
