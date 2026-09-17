'use client'
import React, {useState} from "react";
import {NewPlateForm} from "@/app/components/plateClient";
import {NewAgarBatchForm} from "@/app/components/agarBatchClient";
import {AgarBatchData} from "@/app/components/agarBatchServer";
// import {Dictaphone, SayString} from "@/app/components/common";

export function Closeable(props:React.PropsWithChildren<{title:string}>) {
    const [open,setOpen]=useState(false);
    if(!open){
        return <div>
            <button className={"basicButton buttonFullWidth"} onClick={(e)=>{
                e.stopPropagation()
                setOpen(!open);
            }}>{props.title}</button>
        </div>
    }
    return <div>
        <button className={"basicButton buttonFullWidth"} onClick={(e)=>{
            e.stopPropagation()
            setOpen(!open);
        }}>{"collapse "+props.title}</button>
        {props.children}
    </div>
}

export function TestNewPlate(){
    return <NewPlateForm agarBatchIn={new AgarBatchData({_id:"agarBatchId",agarRecipe:"recipeId",color:"aColor",pcRun:"runId",lastUpdated:1000000})} handlers={{isTopLevel:true,onCreate:(pl)=>console.log("created plate")}}  />
}
export function TestNewAgarBatch(){
    return <NewAgarBatchForm handlers={{isTopLevel:true,onCreate:(item)=>console.log("created agar batch")}}  />
}
// export function TextToSpeech(){
//     return <div>
//         <button className={"buttonFullWidth greenButton"} onClick={e=>{
//             e.preventDefault();
//             e.stopPropagation();
//             const utterance = new SpeechSynthesisUtterance("Hello World")
//             window.speechSynthesis.speak(utterance)
//         }}>{"Say 'Hello World'"}</button>
//     </div>
// }
// export function SpeechToText(){
//     return <div>
//         <Dictaphone createNoteHandler={n=>{
//             SayString("note created: "+n)
//         }}/>
//     </div>
// }