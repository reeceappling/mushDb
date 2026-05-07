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

export interface Additive {
    additive: string
    amount: number
    unit: string
}

export const AdditivesList: string[] = ["Add1", "Add2", "Add3"]

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
                                onSelect && onSelect()
                            }
                            onSelect && onSelect(s as string)
                        }
                        }/>
}

export default function AdditivesArea(props: AreaProps<Additive>) {
    return FormListArea(AdditiveEntriesGroup)(props)
}

export function AdditiveEntriesGroupForNew({currentEntries, updateParent}: {
    currentEntries: Additive[],
    updateParent: (l: Additive[]) => void
}) {
    const handleSelect = (v: string) => {
        let data = [...(currentEntries || []), {additive: v, amount: 0, unit: ""}];
        updateParent(data)
    }
    return <div>
        {currentEntries.length !== 0 && <div className={"inputGrid inputGrid4 gap-8"}>
            {currentEntries.map((n, i) => {
                return <div key={n.additive} className={"contentsOnly"}>
                    <AdditiveEntryForNew currentValue={n} updateParent={(updated: Additive) => {
                        updateParent([...(currentEntries || [])].map((existing) => {
                            return existing.additive !== n.additive ? existing : updated
                        }))
                    }}/>
                    <RemoveButton txt={"Remove"} click={() => {
                        updateParent([...(currentEntries || [])].filter((existing) => existing.additive !== n.additive))
                    }}/>
                </div>
            })}
        </div>}
        <AdditiveTypeSelectorForNew onSelect={(val) => {
            val && handleSelect(val)
        }} blacklist={currentEntries.map((v) => {
            return v.additive
        })}/>
    </div>
}

export function AdditiveEntriesGroup({
                                         initialEntries,
                                         preexisting,
                                         readonly,
                                         updateParent,
                                         blacklist
                                     }: GroupProps<Additive>) {

    const handleFormChangeAdditive = (index: number, ad: string) => {
        let data = initialEntries ? [...initialEntries] : []
        data[index].data.additive = ad
        updateParent(data)
    }
    const handleFormChangeAmt = (index: number, val: number) => {
        // TODO: HANDLE NON-NUMBERS!!!
        let data = initialEntries ? [...initialEntries] : []
        data[index].data.amount = val
        updateParent(data)
    }
    const handleFormChangeUnit = (index: number, val: string) => {
        let data = [...(initialEntries || [])]
        data[index].data.unit = val
        updateParent(data)
    }
    const addFields = (e: React.MouseEvent) => {
        e.preventDefault()
        let data = [...(initialEntries || []), {
            data: {additive: "UNDEFINED", amount: 0.0, unit: 'UNDEFINED'},
            disabled: false
        }] // TODO: FIX DEFAULT
        updateParent(data)
    }
    const removeFields = (index: number) => {
        let data = [...(initialEntries || [])]
        data.splice(index, 1);
        updateParent(data)
    }
    const disableField = (name: string) => {
        let data = [...(initialEntries || [])].map((v, i) => {
            let val = v
            if (val.data.additive === name) {
                val.disabled = !val.disabled
            }
            return val
        });
        updateParent(data)
    }
    const groupClasses = () => {
        let out = "additiveEntryGroup"
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
    const entryClasses = (entry: Data<Additive>) => {
        let out = "additiveEntry"
        if (entry.disabled) {
            out += " disabled"
        } else {
            out += " enabled"
        }
        return out
    }
    return <div className={groupClasses()}>
        {initialEntries?.map((input, index) => {
            if (readonly) {
                return <div className={"readonly " + entryClasses(input)} key={index}>
                    {input.data.amount.toString() + " " + input.data.unit + " of " + input.data.additive + (input.disabled ? " (disabled)" : "" /* TODO: remove? */)}
                </div>
            }
            return <div className={"editable " + entryClasses(input)} key={index}>
                {readonly ? input.data.additive :
                    <TestAndValidate todos={["on delete, not properly changing"]}>
                        <AdditiveSelector
                            initial={input.data.additive} blacklist={blacklist?.map((ad) => {
                            return ad.data.additive
                        })} onSelect={(a) => {
                            a && handleFormChangeAdditive(index, a)
                        }}/>
                        <NumericalArea label="Amount" mode="floating" min={0} readonly={readonly} errorMessage={"FIXME"}
                                       value={input.data.amount.toString()} onChange={(val?: string) => {
                            val && handleFormChangeAmt(index, Number(val))
                        }}/>
                        <TextBox label={"Unit"} readonly={readonly} value={input.data.unit} fieldName={"FIXME"}
                                 updateTextHandler={(t) => {
                                     handleFormChangeUnit(index, t)
                                 }}/>
                    </TestAndValidate>}
                {readonly || <button className={"removeButton"} onClick={() => {
                    (preexisting ? disableField(input.data.additive) : removeFields(index))
                }}>
                    {preexisting ? (input.disabled ? "enable" : "disable") : "remove"}
                </button>}
            </div>
        })}
        {preexisting ? null : <button className={"basicButton"} onClick={addFields}>{"Add More.."}</button>}
    </div>
}
