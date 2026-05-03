'use client'
import * as React from "react";

export function TailwindButton({click,txt}:{txt:string,click:()=>void}){
    return <button className={"basicButton"} onClick={click}>{txt}</button>
}