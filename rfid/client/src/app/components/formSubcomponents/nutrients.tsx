// TODO: THESE ARE ALL NON-CLIENT!

import {AreaProps, FormListArea, GroupProps} from "@/app/components/formSubcomponents/shared";
import {useQuery} from "@tanstack/react-query";
import {SelectorResetsOnSelectFor} from "@/app/components/selector";
import {NumericalArea} from "@/app/components/formSubcomponents/numericInput";
import TextBox from "@/app/components/formSubcomponents/textbox";
import {getOptionsResponse} from "@/app/components/formSubcomponents/server";
import TestAndValidate from "@/app/components/testing/untested";
import {LiquidEntryForNew, NutrientEntryForNew} from "@/app/components/formSubcomponents/commonClient";
import {Liquid, LiquidsTypeSelectorForNew} from "@/app/components/formSubcomponents/liquids";

interface NutrientsAreaProps {
    readonly: boolean,
    initialValues: NutrientData[],
    updateParent: (entries: AllNutrients) => void,
}

export interface Nutrient {
    nutrient: string,
    amount: number,
    unit: string,
}

export const NutrientsList: string[] = ["PotatoFlakes", "LME", "other"]


export function IsValidNutrient(input: any): boolean {
    return (
        typeof input === 'object' &&
        'nutrient' in input && typeof input.nutrient === 'string' &&
        'amount' in input && typeof input.amount === 'number' &&
        'unit' in input && typeof input.unit === 'string'
    )
}

export interface NutrientData {
    data: Nutrient,
    disabled: boolean,
}

export type AllNutrients = {
    existing: NutrientData[],
    new: NutrientData[],
}


export function NutrientTypeSelectorForNew( // TODO: USE THIS!!!!!
    {current, onSelect, readonly, blacklist}:{
        readonly: boolean,
        current?: string,
        onSelect?: (ab?: string)=>void
        blacklist?: string[], // TODO: use?
    }){
    if(readonly){
        return <div>{current || "FIXME"}</div> // TODO: div ok?
    }
    const { isPending, error, data } = useQuery({ // TODO: SO SOMETHING SIMILAR FOR nutrients, sugars, liquid, grain, transferReason
        queryKey: ['nutrientsOptions'],
        queryFn: () => getOptionsResponse("nutrients")
    })
    if(isPending || error !== null){
        return <div>{isPending ? "NUTRIENT SELECTOR LOADING" : "NUTRIENT SELECTOR ERROR: "+error.message}</div>
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

// TODO: UPDATE TO BE LIKE SUGARS, USING GENERICS!!!!
export default function NutrientsArea(props: AreaProps<Nutrient>){
    return FormListArea(NutrientEntriesGroup)(props)
}

export function NutrientsEntriesGroupForNew({currentEntries, updateParent}: {currentEntries: Nutrient[], updateParent: (l: Nutrient[])=>void}){
    const handleSelectNutrient = (v: string) => {
        let data = [...(currentEntries || []), {nutrient: v, amount: 0, unit: ""}];
        updateParent(data)
    }
    return <div>
        {currentEntries.length!==0 && <div className={"inputGrid inputGrid4 gap-8"}>
        {currentEntries.map((n,i)=>{
            return <>
                <NutrientEntryForNew key={n.nutrient+1} currentValue={n} updateParent={(updated: Nutrient) => {
                    updateParent([...(currentEntries || [])].map((existing) => {
                        return existing.nutrient !== n.nutrient ? existing : updated
                    }))
                }}/>
                <button key={n.nutrient+2} className={"removeButton"} onClick={()=>{
                    updateParent([...(currentEntries || [])].filter((existing) => existing.nutrient !== n.nutrient))
                }}>{"Remove"}</button>
            </>
        })}
        </div>}
        <NutrientTypeSelectorForNew onSelect={(val)=>{val && handleSelectNutrient(val)}} blacklist={currentEntries.map((v)=>{return v.nutrient})} readonly={false} />
    </div>
}

export function NutrientEntriesGroup({initialEntries, preexisting, readonly, updateParent}: GroupProps<Nutrient>){
    const handleFormChangeNutrient = (index: number, val: string) => {
        let data = [...(initialEntries || [])];
        data[index].data.nutrient = val
        updateParent(data)
    }
    const handleFormChangeAmt = (index: number, val: number) => {
        let data = [...(initialEntries || [])];
        data[index].data.amount = val
        updateParent(data)
    }
    const handleFormChangeUnit = (index: number, txt: string) => {
        let data = [...(initialEntries || [])];
        data[index].data.unit = txt
        updateParent(data)
    }
    const addFields = (e: React.MouseEvent) => {
        e.preventDefault()
        let data = [...(initialEntries || []), { data: {nutrient: "UNDEFINED", amount: 0.0, unit: 'UNDEFINED'}, disabled: false }] // TODO: FIX DEFAULT
        updateParent(data)
    }
    const removeFields = (index: number) => {
        return (event: MouseEvent) => {
            event.preventDefault()
            let data = [...(initialEntries || [])];
            data.splice(index, 1) // TODO: THIS WONT WORK PROPERLY WITH INDEX
            updateParent(data)
        }
    }
    const disableField = (index: number) => {
        return (event: MouseEvent) => {
            event.preventDefault()
            let data = [...(initialEntries || [])];
            data[index].disabled = !data[index].disabled
            updateParent(data)
        }
    }
    const groupClasses = () => {
        let out = ""
        if(preexisting){
            out = "exists"
        } else {
            out = "new"
        }
        if(readonly){
            out+=" readonly"
        } else {
            out+=" editable"
        }
        return out
    }
    const entryClasses = (entry: NutrientData) => {
        let out = "nutrient"
        if(entry.disabled){
            out+=" disabled"
        } else {
            out+=" enabled"
        }
        return out
    }
    const blacklist = ([...(initialEntries || [])]).map((v)=>{
        return v.data.nutrient
    })
    return <div className={groupClasses()}>{/* TODO: CLASS STYLINGS!!!! */}
        <TestAndValidate todos={["overhaul for non-new (updates)"]}>
        {initialEntries?.map((input, index) => {
            if(readonly){
                return <div className={"readonly "+entryClasses(input)} key={index}>
                    {input.data.amount.toString()+" "+input.data.unit+" of "+input.data.nutrient+"s"+(input.disabled ? " (disabled)" : "" /* TODO: remove? */)}
                </div>
            }
            // TODO: match this up with AdditiveEntriesGroup
            return (
                <div className={"editable "+entryClasses(input)} key={index}> {/* TODO: CLASS STYLINGS!!!! */}
                    {input.disabled?"disabled":null /* TODO: remove? */}
                    {/* TODO: INPUT TAG */}
                    {/* TODO: BLACKLIST AND READONLY? */}
                    <NutrientTypeSelectorForNew onSelect={(val)=>{val && handleFormChangeNutrient(index, val)}} current={input.data.nutrient} blacklist={blacklist} readonly={readonly}/>
                     {/* TODO: INPUT TAG */}
                    <NumericalArea value={input.data.amount.toString()} onChange={(val?:string)=>{val && handleFormChangeAmt(index, Number(val))}} label="Amount" min={0} step={1} errorMessage={'invalid amount'} mode={"floating"} readonly={readonly} />
                    <TextBox readonly={readonly} label={"Unit"} value={input.data.unit} fieldName={"FIXME"} updateTextHandler={(t)=>{handleFormChangeUnit(index, t)}} />

                    {readonly ? null :
                        <button className={input.disabled?"removeButton":"basicButton"} onClick={()=>{preexisting?disableField(index):removeFields(index)}}>{preexisting ? (input.disabled?"enable":"disable") : "remove"}</button>}
                </div>)
        })}
        {preexisting ? null : <button className={"basicButton"} onClick={addFields}>Add More..</button>}
        </TestAndValidate>
    </div>
}