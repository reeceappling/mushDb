'use client'

import {useEffect, useState} from "react";

export function TopLevelImageSelector({updateParent, buttonText}:{buttonText?:string, updateParent: (f: File | undefined)=> void}) {
    return <div className={"centerH padContent topLevelImageSelector"}>
        <ImageSelector updateParent={updateParent} buttonText={buttonText}/>
    </div>
}

export default function ImageSelector({updateParent, buttonText}:{buttonText?:string, updateParent: (f: File | undefined)=> void}) {
    const [hasCamera, setHasCamera] = useState<boolean | undefined>(undefined);
    const [file, setFile] = useState<File | undefined>(undefined)
    const [inputElement, setInputElement] = useState<HTMLInputElement | null>();
    const setInputRef = (element:HTMLInputElement) => {
        setInputElement(element);
    };
    const handleImageSelected = (e: React.ChangeEvent<HTMLInputElement>) => {
        if (e.currentTarget.files != null && e.currentTarget.files.length > 0) {
            const fileToSave = e.currentTarget.files.item(0)
            if (fileToSave !== null) {
                updateParent(fileToSave)
                setFile(fileToSave)
                e.currentTarget.blur();
                (document.activeElement as HTMLElement)?.blur();
                return
            }
            updateParent(undefined)
            setFile(undefined)
            e.currentTarget.blur();
            (document.activeElement as HTMLElement)?.blur();
            return
        }
    }
    // Check if the device has a camera
    if (!navigator.mediaDevices || !navigator.mediaDevices.enumerateDevices){
        console.error("no media devices found")
        setHasCamera(false)
    } else {
        navigator.mediaDevices.enumerateDevices().then(devices=>{
            const cams = devices.filter(device => device.kind === 'videoinput')
            setHasCamera(cams.length > 0)
        }).catch(e=> {
                console.error("failed to get media devices")
                setHasCamera(false)
            }
        );
    }

    return <div className={"imageSelector picLeft"}>
        {file !== undefined && <div className={"preview"}> {/* TODO: FIX SIZE!*/}
            <img className={"picDisplay"} src={URL.createObjectURL(file)} alt="image preview"/>
        </div>}
        <div className={"centerH"}>
            <button className={"basicButtonSmall"} onClick={() => {
                if (inputElement) {
                    inputElement.click();
                }
            }}>{buttonText || "Choose File"}</button>
            <input className="hidden" type="file" ref={setInputRef} accept={"image/*"+(hasCamera ? ",application/octet-stream":"")} // ,application/octet-stream is a fix for google pixel phones, and other androids
                   onChange={handleImageSelected}/>
            {/*TODO: custom dual-button overlay so that we can either pick or take a photo instead of doing the weird octet stream stuff!!!*/}
            {/*<input className="hidden" type="file" ref={setInputRef} accept={"image/*"/*photo picker only*!/*/}
            {/*       onChange={handleImageSelected}/>*/}
            {/*<input className="hidden" type="file" ref={setInputRef} accept={"image/*"} capture={"environment"/*camera only*!/*/}
            {/*       onChange={handleImageSelected}/>*/}
        </div>
    </div>
}