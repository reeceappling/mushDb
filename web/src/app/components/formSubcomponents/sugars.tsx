import {ChangeEvent} from "react";
import {AreaProps, Data, FormListArea, GroupProps} from "@/app/components/formSubcomponents/shared";
import {SelectorFor, SelectorResetsOnSelectFor} from "@/app/components/selector";
import {useQuery} from "@tanstack/react-query";
import {NumericalArea} from "@/app/components/formSubcomponents/numericInput";
import TextBox from "@/app/components/formSubcomponents/textbox";
import {getOptionsResponse} from "@/app/components/formSubcomponents/server";
import TestAndValidate from "@/app/components/testing/untested";
import {RemoveButton, SugarEntryForNew} from "@/app/components/formSubcomponents/commonClient";
import * as React from "react";

export interface Sugar {
    type: string,
    amount: number,
    unit: string,
}

export const SugarsList: string[] = ["dextrose", "honey", "other"]


export function IsValidSugar(input: any): boolean {
    return (
        typeof input === 'object' &&
        'type' in input && typeof input.type === 'string' &&
        'amount' in input && typeof input.amount === 'number' &&
        'unit' in input && typeof input.unit === 'string'
    )
}

export function SugarTypeSelectorForNew( // TODO: USE THIS!!!!!
    {current, onSelect, readonly, blacklist}:{
        readonly: boolean,
        current?: string,
        onSelect?: (ab?: string)=>void
        blacklist?: string[], // TODO: use?
    }){
    if(readonly){
        return <div>{current || "FIXME"}</div>
    }
    const { isPending, error, data } = useQuery({
        queryKey: ['sugarsOptions'],
        queryFn: () => getOptionsResponse("sugars")
    })
    if(isPending || error !== null){
        return <div>{isPending ? "SUGAR SELECTOR LOADING" : "SUGAR SELECTOR ERROR: "+error.message}</div>
    }
    const filteredOptions = data.filter((val, idx)=>{
        return !(blacklist && blacklist.includes(val))
    })
    return <SelectorResetsOnSelectFor options={["", ...filteredOptions]} updateParent={(s)=>{
        if(s===""){
            onSelect && onSelect()
        }
        onSelect && onSelect(s as string)}
    } />
}

export function SugarTypeSelector( // TODO: USE THIS!!!!!
    {initial, onSelect, blacklist}: {
        initial?: string,
        onSelect?: (ab?: string) => void,
        blacklist?: string[],
    }) {
    const {isPending, error, data} = useQuery({
        queryKey: ['sugarsOptions'],
        queryFn: () => getOptionsResponse("sugars")
    })
    if (isPending || error !== null) {
        return <div>{isPending ? "SUGAR SELECTOR LOADING" : "SUGAR SELECTOR ERROR: " + error.message}</div>
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
} // TODO: "honey","maple syrup","simple syrup"

export default function SugarsArea(props: AreaProps<Sugar>) {
    return FormListArea(SugarEntriesGroup)(props)
}
export function SugarsAreaReadOnly({values}: {values?:Sugar[]}) {
    if (!values || values.length===0){
        return null
    }
    return <div>
        {"Sugars: "}
        {values.map((v, i) => {
            return <div key={v.type}>{v.type + " - " + v.amount + " " + v.unit}</div>
        })}
    </div>
}

export function SugarEntriesGroupForNew({currentEntries, updateParent}: {currentEntries: Sugar[], updateParent: (l: Sugar[])=>void}){
    const handleSelect = (v: string) => {
        const data = [...(currentEntries || []), {type: v, amount: 0, unit: ""}];
        updateParent(data)
    }
    return <div>
        {currentEntries.length!==0 && <div className={"inputGrid inputGrid4 gap-8"}>
            {currentEntries.map((n,i)=>{
            return <div key={n.type} className={"contentsOnly"}>
                <SugarEntryForNew currentValue={n} updateParent={(updated: Sugar) => {
                    updateParent([...(currentEntries || [])].map((existing) => {
                        return existing.type !== n.type ? existing : updated
                    }))
                }}/>
                <RemoveButton txt={"Remove"} click={()=>{
                    updateParent([...(currentEntries || [])].filter((existing) => existing.type !== n.type))
                }} />
            </div>
        })}
        </div>}
        <SugarTypeSelectorForNew onSelect={(val)=>{val && handleSelect(val)}} blacklist={currentEntries.map((v)=>{return v.type})} readonly={false} />
    </div>
}

// TODO: REMOVAL IS NOT MOVING THE TYPES UP! PROBABLY OVERHAUL
export function SugarEntriesGroup({initialEntries, preexisting, readonly, updateParent, blacklist}: GroupProps<Sugar>) {
    const handleFormChangeSugarType = (index: number, event: ChangeEvent<HTMLInputElement>) => {
        const data = initialEntries ? [...initialEntries] : []
        data[index].data.type = event.target.value
        updateParent(data)
    }
    const handleSelSugarType = (index: number, newType: string) => {
        const data = initialEntries ? [...initialEntries] : []
        data[index].data.type = newType
        updateParent(data)
    }
    const handleFormChangeAmt = (index: number, amt: number) => {
        const data = initialEntries ? [...initialEntries] : []
        data[index].data.amount = amt
        if (isNaN(data[index].data.amount)) {
            console.log(amt + " could not be parsed to a number in handleFormChangeAmt in SugarEntriesGroup")
            return
        }
        updateParent(data)
    }
    const handleFormChangeUnit = (index: number, txt: string) => {
        const data = [...(initialEntries || [])]
        data[index].data.unit = txt
        updateParent(data)
    }
    const addFields = (e: React.MouseEvent) => {
        e.preventDefault()
        const data = [...(initialEntries || []), {
            data: {type: "NEW SUGAR TYPE", amount: 0.0, unit: 'NEW SUGAR UNIT'},
            disabled: false
        }]
        updateParent(data)
    }
    const removeFields = (index: number) => {
        const data = [...(initialEntries || [])]
        data.splice(index, 1);
        updateParent(data)
    }
    const disableField = (name: string) => {
        const data = [...(initialEntries || [])].map((v, i) => {
            const val = v
            if (val.data.type === name) {
                val.disabled = !val.disabled
            }
            return val
        });
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
    const entryClasses = (entry: Data<Sugar>) => {
        let out = "sugarEntry"
        if (entry.disabled) {
            out += " disabled"
        } else {
            out += " enabled"
        }
        return out
    }
    // TODO: match this up with AdditiveEntriesGroup

    return <div className={groupClasses()}>{/* TODO: CLASS STYLINGS!!!! */}
        {initialEntries?.map((input, index) => {
            if(readonly){
                return <div className={"readonly "+entryClasses(input)} key={index}>
                    {input.data.amount.toString()+" "+input.data.unit+" of "+input.data.type+(input.disabled ? " (disabled)" : "" /* TODO: remove? */)}
                </div>
            }
            return (
                <div className={"editable "+entryClasses(input)} key={index}> {/* TODO: CLASS STYLINGS!!!! */}
                    {input.disabled ? "disabled" : null /* TODO: remove? */}
                    <label htmlFor="sugar" className="entriesGroupLabel">{"Type"}</label>
                        {readonly ? // TODO: EITHER USE INPUT FOR BOTH, OR SPLIT WITH INPUT AND SELECTOR
                            <input name='sugar' value={input.data.type}
                                   onChange={event => handleFormChangeSugarType(index, event)} readOnly={readonly}/> // TODO: USE SOMETHING SIMILAR TO LiquidOptions
                            : <TestAndValidate
                                todos={["on delete, not properly changing (if top is deleted, then top stays as value it was visually)"]}>
                                <SugarTypeSelector blacklist={blacklist?.map(s => {
                                    return s.data.type
                                })} onSelect={(nt?: string) => {
                                    nt && handleSelSugarType(index, nt)
                                }}/>
                            </TestAndValidate>
                        }
                    <NumericalArea mode="floating" min={0} readonly={readonly} errorMessage={"FIXME"}
                                   value={input.data.amount.toString()} label="Amount" onChange={(val?: string) => {
                        val && handleFormChangeAmt(index, Number(val))
                    }}/>
                        <TextBox label={"Unit"} readonly={readonly} value={input.data.unit} fieldName={"FIXME"}
                                 updateTextHandler={(t) => {
                                     handleFormChangeUnit(index, t)
                                 }}/>{/* TODO: not working*/}
                        {(!readonly) && <button className={input.disabled?"removeButton":"basicButton"} onClick={() => {
                            (preexisting ? disableField(input.data.type) : removeFields(index))
                        }}>
                            {preexisting ? (input.disabled ? "enable" : "disable") : "remove"}
                        </button>}
                </div>)
        })}
        {preexisting ? null : <button className={"basicButton"} onClick={addFields}>Add More..</button>}
    </div>
}
