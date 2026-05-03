'use client'
// TODO: THESE ARE ALL NON-CLIENT?

import {useState} from "react";
import TestAndValidate from "@/app/components/testing/untested";

export function TopLevelImageSelector({updateParent, buttonText}:{buttonText?:string, updateParent: (f: File | undefined)=> void}) {
    return <div className={"centerH padContent topLevelImageSelector"}>
        <ImageSelector updateParent={updateParent} buttonText={buttonText}/>
    </div>
}

export default function ImageSelector({updateParent, buttonText}:{buttonText?:string, updateParent: (f: File | undefined)=> void}) {
    const [file, setFile] = useState<File | undefined>(undefined)
    const [inputElement, setInputElement] = useState<HTMLInputElement | null>();
    const setInputRef = (element:HTMLInputElement) => {
        setInputElement(element);
    };
    const handleImageSelected = (e: React.ChangeEvent<HTMLInputElement>) => {
        if (e.currentTarget.files != null && e.currentTarget.files.length > 0) {
            let fileToSave = e.currentTarget.files.item(0)
            if (fileToSave !== null) {
                updateParent(fileToSave)
                setFile(fileToSave)
                return
            }
            updateParent(undefined)
            setFile(undefined)
            return
        }
    }
    return <TestAndValidate todos={["ensure working"]}><div className={"imageSelector picLeft"}>
        {file !== undefined && <div className={"preview"}> {/* TODO: FIX SIZE!*/}
            <img className={"picDisplay"} src={URL.createObjectURL(file)} alt="image preview"/>
        </div>}
        <div className={"centerH"}>
            <button onClick={() => {
                if (inputElement) {
                    inputElement.click(); // TODO: ENSURE OK
                }
            }}>{buttonText || "Choose File"}</button>
            <input className="hidden" type="file" ref={setInputRef} accept="image/*;capture=camera" capture="user"
                   onChange={handleImageSelected}/>
        </div>
    </div></TestAndValidate>
}