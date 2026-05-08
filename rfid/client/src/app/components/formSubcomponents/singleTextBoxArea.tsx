import {ChangeEvent, useState} from "react";
import {AreaProps, Data, FormListArea, GroupProps} from "@/app/components/formSubcomponents/shared";
import * as React from "react";
import {RemoveToggle} from "@/app/components/formSubcomponents/commonClient";
import {InputText} from "@/app/components/formSubcomponents/numericInput";

export default function TextBoxArea(props: AreaProps<string>){
    return FormListArea(StringEntriesGroup)(props)
}

function StringEntriesGroup({initialEntries, preexisting, readonly, updateParent}: GroupProps<string>){
    const [inputFields, setInputFields] = useState(initialEntries || [])
    const handleFormChangeText = (index: number, event: ChangeEvent<HTMLInputElement>) => {
        let data = [...inputFields];
        data[index].data = event.target.value
        updateParent(data)
        setInputFields(data);
    }
    const handleFormChangeText2 = (index: number, newVal: string) => {
        let data = [...inputFields];
        data[index].data = newVal
        updateParent(data)
        setInputFields(data);
    }
    const addFields = () => {
        let data: Data<string>[] = [{ data: '', disabled: false }]
        if(inputFields.length!==0){
            data = [...inputFields, { data: '', disabled: false }] // TODO: FIX DEFAULT?
        }
        updateParent(data)
        setInputFields(data)
    }
    const removeFields = (index: number) => {
        return () => {
            let data = [...inputFields];
            data.splice(index, 1) // TODO: THIS WONT WORK PROPERLY WITH INDEX
            updateParent(data)
            setInputFields(data)
        }
    }
    const disableField = (index: number) => {
        return () => {
            let data = [...inputFields]
            data[index].disabled = !data[index].disabled
            updateParent(data)
            setInputFields(data)
        }
    }
    const textAreaClasses = () => {
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
    const txtClasses = (note: Data<string>) => {
        let out = "inlineChildren textBox"
        if(note.disabled){
            out+=" disabled"
        } else {
            out+=" enabled"
        }
        return out
    }
    return <div className={textAreaClasses()}>{/* TODO: CLASS STYLINGS!!!! */}
        {inputFields.map((input, index) => {
            return (
                <div className={txtClasses(input)} key={index}> {/* TODO: CLASS STYLINGS!!!! */}
                    {input.disabled?"disabled":null /* TODO: remove? gray out instead? */}
                    {/* TODO: INPUT TAG */}
                    {readonly ? <div>{input.data}</div> :
                        <InputText value={input.data} readonly={false} onChange={(s)=>{s && handleFormChangeText2(index, s)}} />}
                    {readonly || <button className={input.disabled?"basicButtonSmall":"removeButtonSmall"} onClick={(e)=>{
                            e.stopPropagation();
                            preexisting?disableField(index)():removeFields(index)()
                        }}>{preexisting ? (input.disabled?"enable":"disable") : "remove"}</button>}
                </div>)
        })}
        {preexisting ? null : <button className={"basicButton"} onClick={addFields}>{"Add More.."}</button>}
    </div>
}